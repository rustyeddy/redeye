#include <iostream>

#include "vid.hpp"
#include "win.hpp"

extern cv::VideoCapture g_cap;
extern int g_run;
extern int g_dontset;

void onTrackbarSlide( int pos, void * ) {
    g_cap.set( cv::CAP_PROP_POS_FRAMES, pos );

    if ( !g_dontset ) {
        g_run = 1;
        g_dontset = 0;
    }
}

Window::Window(string name)
{
    _name = name;
    int frames = (int) g_cap.get(cv::CAP_PROP_FRAME_COUNT);
    if (frames > 0) {
        _slider = true;
    }

    if (frames > 0) {
        cv::createTrackbar("Position", "X", &_slider_pos, frames, onTrackbarSlide);
    }
    int curpos = (int) g_cap.get(cv::CAP_PROP_POS_FRAMES);
    g_dontset = 1;
    if (curpos > 0) {
        cv::setTrackbarPos("Postion", "X", curpos);                
    }
}
