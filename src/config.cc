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
    while ((opt = getopt(argc, argv, "b:df:i:mo:sv")) != -1) {
        switch (opt) {
        case 'b':
            _mqtt_broker = optarg;
            break;

        case 'd':
            _display = true;
            break;

        case 'f':
            _filter_name = optarg;
            break;

        case 'i':
            _id = optarg;
            break;

        case 'm':
            _mjpg = true;
            break;

        case 'o':
            _outdir = optarg;
            break;

        case 's':
            _start_server = true;
            break;

        case 'v':
            verbose = true;
            break;

        default:
            cerr << "ERROR unknown option: " << to_string(opt) << endl;
            cerr << "Usage: re [ -v <videodev> | -i interface ] [-f <filter-name>]" << endl;
            exit(1);
            break;
        }
    }

    for (; optind < argc; optind++) {
        string vidname(argv[optind]);
        _video_srcs.push_back(vidname);
    }

    return 1;
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
    cout << "Web Port: "        << _web_port << endl;
}
