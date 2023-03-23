#include "filter.hh"
#include "filter_default.hh"

cv::Mat* FltGaussianBlur::filter(cv::Mat* img)
{
    cv::GaussianBlur( *img, *img, cv::Size(5,5), 3, 3);
    cv::GaussianBlur( *img, *img, cv::Size(5,5), 3, 3);
    return img;
}

cv::Mat* FltCanny::filter(cv::Mat* img)
{
    cv::cvtColor(*img, *img, cv::COLOR_BGR2GRAY);
    cv::Canny( *img, *img, 10, 100, 3, true );
    return img;
}

cv::Mat* FltSmaller::filter(cv::Mat* img)
{
    cv::pyrDown( *img, *img );
    return img;
}

cv::Mat* FltBorder::filter(cv::Mat* img)
{
    cv::copyMakeBorder(*img, *img, 10, 10, 10, 10, cv::BORDER_CONSTANT, cv::Scalar(33,33,33));
    return img;
}
