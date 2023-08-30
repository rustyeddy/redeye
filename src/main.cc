#include <iostream>
#include <thread>
#include <unistd.h>
#include <opencv2/opencv.hpp>

#include "config.hh"
#include "mqtt.hh"
#include "net.hh"
#include "player.hh"
#include "video.hh"
#include "web.hh"

#include "filter.hh"

using namespace std;

Config*         config  = NULL;
FltFilters*     filters = NULL;
Filter*         flt = NULL;
string ID       = "";
MQTT*           mqtt = NULL;

map<string, Player*> video_players;

int start_server(Config *config);
int process_file(Config *config);
void* hello_loop(void *);

int main(int argc, char *argv[], char *envp[])
{
    config = new Config( argc, argv, envp );
    config->dump();

    mqtt = new MQTT("localhost");

    filters = new FltFilters();
    flt = filters->get(config->get_filter_name());
    if (config->get_filter_name() != "" && flt == NULL) {
        cout << "Could not find the filter " << config->get_filter_name() << endl;
        exit(1);
    }
    if (flt != NULL) flt->init();

    int rc = start_server(config);
    if (rc == -1) {
        exit(1);
    }

    // // Start the server if we have been configured to do so.
    // if (config->start_server()) {
    //     exit(0);
    // }
    
    // We must have a file
    // if (process_file(config)) {
    //     cerr << "Failed to process file: " << config->get_file_name() << endl;
    //     exit(1);
    // }
    exit(0);
}

int start_server(Config *config)
{
    cout << "Getting video sources\n" ;
    auto vsrcs = config->get_video_sources();
    if (vsrcs.size() < 1) {
        cerr << "No video sources have been identified, exiting..." << endl;
        return -1;
    }

    // TODO: this will need to be fixed for other machines
    ID = get_ip_address(config->get_iface()); 
    pthread_t t_mqtt;
    pthread_t t_player;
    pthread_t t_web;
    pthread_t t_hello;
    
    for ( string vname : config->get_video_sources() ) {

        cout << "Opening video source: " << vname << endl;

        Video* vid = new Video(vname);

        Player* player  = new Player();
        player->set_filter( config->get_filter_name() );
        player->add_imgsrc( vid );
        video_players[vname] = player;

        // mqtt_add_player()
        // cv::startWindowThread();
        pthread_create(&t_player, NULL, &play_video, player);
        // cv::destroyAllWindows();
    }
    
    pthread_create(&t_mqtt, NULL, mqtt_loop, (char *)ID.c_str());
    pthread_create(&t_web,  NULL, web_start, NULL);
    pthread_create(&t_hello, NULL, hello_loop, NULL);

    pthread_join(t_web, NULL);
    pthread_join(t_mqtt, NULL);
    pthread_join(t_player, NULL); 
    pthread_join(t_hello, NULL);

    cout << "Goodbye, all done. " << endl;
    exit(0);
}


int process_file(Config *config)
{
    VideoCapture cap;
    Mat frame;

    cap.open(config->get_file_name());
    if (!cap.isOpened()) {
        cerr << "ERROR! Unable to open camera\n";
        return -1;
    }

    //--- GRAB AND WRITE LOOP
    cout << "Start grabbing" << endl
         << "Press any key to terminate" << endl;

    for (;;) {
        // wait for a new frame from camera and store it into 'frame'
        cap.read(frame);
        if (frame.empty()) {
            cerr << "ERROR! blank frame grabbed\n";
            break;
        }

        Mat *f2 = flt->filter(&frame);

        // show live and wait for a key with timeout long enough to show images
        imshow("Live", *f2);
        waitKey(0) >= 0;
    }
    return 0;
}

void* hello_loop(void *)
{
    int running = true;

    string jstr = "{";
    jstr += "\"addr\":\"" + ID + "\",";
    jstr += "\"port\":" + to_string(config->get_mjpg_port()) + ",";
    jstr += "\"name\":\"" + ID + "\",";
    jstr += "\"uri\": \"" + config->get_video_uri() + "\"";
    jstr += "}";

    while (running) {
        sleep(10);              // announce every 10 seconds
        mqtt->publish("redeye/announce/camera", jstr.c_str());
    }
    return NULL;
}

