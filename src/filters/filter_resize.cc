#include <opencv2/opencv.hpp>

#include "../filter.hh"
#include "filter_resize.hh"

cv::Mat* FltResize::filter(cv::Mat* img)
{
    cv::resize(*img, *img, cv::Size(img->rows, img->cols), 2, 2);
    return img;
}
