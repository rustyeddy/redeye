#pragma once

#include <opencv2/opencv.hpp>

class Imgsrc
{
public:
    virtual cv::Mat* get_frame() = 0;
};
