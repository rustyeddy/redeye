#pragma once

#include <string>
#include <map>
#include <opencv2/opencv.hpp>

#include "imgsrc.hh"

using namespace std;

struct Dimensions {
    int capture_width = 1280 ;
    int capture_height = 720 ;
    int display_width = 1280 ;
    int display_height = 720 ;
    int framerate = 60 ;
    int flip_method = 0 ;

    Dimensions( int w, int h ) {
        display_width = capture_width = w;
        display_height = capture_height = h;
    }

    Dimensions( int w = 1280, int h = 720, int f = 60 ) {
        display_width = capture_width = w;
        display_height = capture_height = h;
        framerate = f;
    }
};

class Video : public Imgsrc
{
    string              _name;
    cv::VideoCapture    _cap;
    
public:
    Video( string camstr );
    Video( int devnum );

    Dimensions  dims = Dimensions( 1280, 720, 60 );
    string	get_tegra();
    cv::Mat*    get_frame();        // add the << operator for reading frames
    double      get_fps()    { return _cap.get( cv::CAP_PROP_FRAME_COUNT ); }
    int         get_width()  { return _cap.get( cv::CAP_PROP_FRAME_WIDTH ); }
    int         get_height() { return _cap.get( cv::CAP_PROP_FRAME_HEIGHT ); }
};

