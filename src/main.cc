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

Config*         config  = NULL;
Player*         player  = NULL;
FltFilters*     filters = NULL;

string ID       = "";

using namespace std;

int start_server(Config *config);
int process_file(Config *config);
void* hello_loop(void *);

int main(int argc, char *argv[], char *envp[])
{
    config = new Config( argc, argv, envp );
    config->dump();

    filters = new FltFilters();
    Filter *flt = filters->get(config->get_filter_name());
    if (flt == NULL) {
        cout << "Could not find the filter " << config->get_filter_name() << endl;
        exit(1);
    }

    // Start the server if we have been configured to do so.
    if (config->start_server()) {
        start_server(config);
        exit(0);
    }
    
    // We must have a file
    if (process_file(config)) {
        cerr << "Failed to process file: " << config->get_file_name() << endl;
        exit(1);
    }
    exit(0);
}

int process_file(Config *config)
{
    Mat frame = imread(config->get_file_name());

    // check if we succeeded
    if (frame.empty()) {
        cerr << "ERROR! blank frame grabbed\n";
        return -1;
    }
    // show live and wait for a key with timeout long enough to show images
    imshow("Live", frame);
    waitKey(5) >= 0;

    return 0;
}

int start_server(Config *config)
{
    // TODO: this will need to be fixed for other machines
    ID = get_ip_address(config->get_iface()); 
    pthread_t t_mqtt;
    pthread_t t_player;
    pthread_t t_web;
    pthread_t t_hello;
    
    pthread_create(&t_mqtt, NULL, mqtt_loop, (char *)ID.c_str());
    pthread_create(&t_web,  NULL, web_start, NULL);
    pthread_create(&t_hello, NULL, hello_loop, NULL);

    player  = new Player( config->get_filter_name() );
    player->set_filter( config->get_filter_name() );
    player->add_imgsrc( config->get_video() );

    cv::startWindowThread();
    pthread_create(&t_player, NULL, &play_video, player);
    cv::destroyAllWindows();
    
    pthread_join(t_web, NULL);
    pthread_join(t_mqtt, NULL);
    pthread_join(t_player, NULL); 
    pthread_join(t_hello, NULL);

    cout << "Goodbye, all done. " << endl;
    exit(0);
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
        mqtt_publish("redeye/announce/camera", jstr.c_str());
    }
    return NULL;
}
