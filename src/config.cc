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
    while ((opt = getopt(argc, argv, "b:f:i:v:")) != -1) {
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

