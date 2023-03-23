#include <string>
#include <opencv2/opencv.hpp>

#include "image.hh"

Image::Image(std::string fname)
{
    _file_name = fname;
}

cv::Mat* Image::get_frame()
{
    auto m = new cv::Mat( cv::imread( _file_name, cv::IMREAD_COLOR ) );
    return m;
}
