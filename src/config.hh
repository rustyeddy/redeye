#pragma once

#include <string>
#include <vector>

#include "filter.hh"
#include "filters.hh"
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
    string      _outdir         = "redout";
    bool        _start_server   = false;
    string      _video_uri      = "";
    int         _web_port       = 8000;

    vector<string>  _video_srcs;     


public:
    Config( int argc, char *argv[], char *envp[] );
    int parse_args( int argc, char *argv[], char *envp[] );

    string      get_filter_name()       { return _filter_name; }
    string      get_file_name()         { return _file_name; }

    int         start_server()          { return _start_server; }
    string      get_mqtt_broker()       { return _mqtt_broker; }
    int         get_mjpg_port()         { return _mjpg_port; }
    int         get_web_port()          { return _web_port; }

    string	get_iface()		{ return _iface; }
    Image*      get_image();
    string      get_video_uri()         { return _video_uri; }
    void        dump();

    vector<string> get_video_sources()     { return _video_srcs; }
};

extern Config *config;
