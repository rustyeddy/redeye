#include <string>

#include <opencv2/opencv.hpp>

#include "filter.hh"

Filter::Filter(string n)
{
    _name = n;
}

string Filter::Name() {
    return _name;
}

cv::Mat* Filter::filter(cv::Mat* iframe)
{
    return iframe;
}

// bool Filter::save_avi(string fname, cv::Mat* iframe)
// {
//     cv::VideoWriter wr;
//     //wr.open(fname, cv::FOURCC('M','J','P','G'), fps, size);
//     return true;
// }

