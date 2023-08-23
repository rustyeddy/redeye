#include <unistd.h>

#include "config.hh"
#include "filter.hh"
#include "filters.hh"
#include "video.hh"

Config::Config( int argc, char *argv[], char *envp[ ] )
{
    parse_args( argc, argv, envp );
}

int Config::parse_args( int argc, char *argv[], char *envp[] )
{
    int opt;
    while ((opt = getopt(argc, argv, "b:f:i:o:sv:")) != -1) {
        switch (opt) {
        case 'b':
            _mqtt_broker = optarg;
            break;

        case 'f':
            _filter_name = optarg;
            break;

        case 'i':
            _iface = optarg;
            break;

        case 'o':
            _outdir = optarg;
            break;

        case 's':
            _start_server = true;
            break;

        case 'v':
            _video_name = optarg;
            break;

        default:
            cerr << "ERROR unknown option: " << to_string(opt) << endl;
            cerr << "Usage: re [ -v <videodev> | -i interface ] [-f <filter-name>]" << endl;
            exit(1);
            break;
        }
    }

    cout << "O: " << optind << " - A: " << argc << endl;

    // Allow multiple images
    if (argc > optind) {
        _file_name = argv[optind];
    }

    return 1;
}

Video*  Config::get_video()
{
    if ( _video_name == "" ) return NULL;
    Video *vid = new Video( _video_name );
    return vid;
}

Image*  Config::get_image()
{
    if ( _file_name  == "" ) {
        cerr << "Must specify an image to display. " << endl;
        exit(-5);
    }
    Image *img = new Image( _file_name );
    return img;
}

void    Config::dump()
{
    cout << "Filter: "          << _filter_name << endl;
    cout << "File: "            << _file_name << endl;
    cout << "Gstreamer: "       << _gstreamer << endl;
    cout << "Interface: "       << _iface << endl;
    cout << "MQTT Broker: "     << _mqtt_broker << endl;
    cout << "Outdir: "          << _outdir << endl;
    cout << "MJPG Port: "       << _mjpg_port << endl;
    cout << "Video URI: "       << _video_uri << endl;
    cout << "Web Port: "        << _web_port << endl;
}
