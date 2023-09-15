#include <iostream>
#include <thread>
#include <unistd.h>
#include <opencv2/opencv.hpp>

#include "config.hh"
#include "event.hh"
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
EventLoop       events;

Players         video_players;

int start_server(Config *config);
int process_file(Config *config);
void* hello_loop(void *);

int main(int argc, char *argv[], char *envp[])
{
    pthread_t t_events;
    pthread_t t_mqtt;
    pthread_t t_player;
    pthread_t t_web;
    pthread_t t_hello;

    config = new Config( argc, argv, envp );
    // config->dump();

    mqtt = new MQTT("localhost");

    // Move this initialization into a better class
    filters = new FltFilters();
    flt = filters->get(config->get_filter_name());
    if (config->get_filter_name() != "" && flt == NULL) {
        cout << "Could not find the filter " << config->get_filter_name() << endl;
        exit(1);
    }
    if (flt != NULL) flt->init();

    cout << "Getting video sources\n" ;
    auto vsrcs = config->get_video_sources();
    if (vsrcs.size() < 1) {
        cerr << "No video sources have been identified, exiting..." << endl;
        return -1;
    }
    
    for ( string vname : config->get_video_sources() ) {

        Player* player = video_players.add(vname);
        if (flt != NULL) {
            player->set_filter(flt);
        }

        cv::startWindowThread();
        pthread_create(&t_player, NULL, &play_video, player);
        cv::destroyAllWindows();
    }

    pthread_create(&t_events, NULL, event_loop, &events);
    pthread_create(&t_mqtt, NULL, mqtt_loop, (char *)NULL);
    pthread_create(&t_web,  NULL, web_start, NULL);
    pthread_create(&t_hello, NULL, hello_loop, NULL);

    // Catch all player threads here
    pthread_join(t_mqtt, NULL);
    pthread_join(t_web, NULL);
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
        // imshow("Live", *f2);
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

