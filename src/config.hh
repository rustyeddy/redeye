#pragma once

#include <string>

#include "filters/filter.hh"
#include "filters/filters.hh"
#include "image.hh"
#include "video.hh"

using namespace std;

class Config
{
private:
    string      _filter_name    = "";
    string      _file_name      = "";
    string      _gstreamer      = "";
    string      _iface          = "eth0";
    string      _mqtt_broker    = "localhost";
    int         _mjpg_port      = 1234;
    string      _video_name     = "";
    string      _video_uri      = "/video0";
    int         _web_port       = 8000;

public:
    Config( int argc, char *argv[], char *envp[] );
    int parse_args( int argc, char *argv[], char *envp[] );

    string      get_filter_name()       { return _filter_name; }
    string      get_video_name()        { return _video_name; }

    string      get_mqtt_broker()       { return _mqtt_broker; }
    int         get_mjpg_port()         { return _mjpg_port; }
    string      get_video_uri()         { return _video_uri; } 
    int         get_web_port()          { return _web_port; }

    string	get_iface()		{ return _iface; }

    Video*      get_video();
    Image*      get_image();
};

extern Config *config;
