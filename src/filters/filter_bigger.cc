#include <opencv2/opencv.hpp>

#include "filter.hh"
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

void bigger_mouse_callback( int event, int x, int y, int flags, void *param )
{
    FltBigger *flt = (FltBigger *) param;
    cout << "Mouse Event: " << event << ", x: " << x << ", y: " << y;

    switch ( event ) {
    case EVENT_MOUSEMOVE: 
        cout << " --- Move " << endl;
        break;

    case EVENT_LBUTTONDOWN:
        cout << " --- left button down - start box. " << endl;
        break;

    case EVENT_LBUTTONUP:
        cout << " --- left button up. " << endl;
        break;

    case EVENT_RBUTTONDOWN:
        cout << " --- right button down. " << endl;
        break;

    case EVENT_RBUTTONUP:
        cout << " --- right button up. " << endl;
        break;

    default:
        cout << " ---  unknown " << event << endl;
    }

}
