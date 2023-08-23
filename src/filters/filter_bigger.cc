#include <opencv2/opencv.hpp>

#include "../filter.hh"
#include "filter_bigger.hh"

FltBigger::FltBigger() : Filter("bigger")
{
}

cv::Mat* FltBigger::filter(cv::Mat* img)
{
    cv::pyrUp(*img, *img);
    return img;
}

