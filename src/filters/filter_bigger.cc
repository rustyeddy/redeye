#include <opencv2/opencv.hpp>

#include "../filter.hh"
#include "filter_bigger.hh"

FltBigger::FltBigger() : Filter("bigger")
{
}

cv::Mat* FltBigger::filter(cv::Mat* img)
{
    static bool init = true;
    if (init) {
        cout << "Intializing mouse callback" << endl;
        cv::setMouseCallback( _name, bigger_mouse_callback, (void *) this ); 
        init = false;
    }

    cv::pyrUp(*img, *img);
    return img;
}

