#pragma once

#include <opencv2/objdetect.hpp>
#include <opencv2/highgui.hpp>
#include <opencv2/imgproc.hpp>
#include <opencv2/videoio.hpp>
#include <iostream>

#include "../filter.hh"

using namespace std;

// Face detector using Haar Cascade Classifier
class FltHaarCascade : public Filter {
private:
    bool                _filter_ok = false;
    string              _face_cascade_name;
    string              _eyes_cascade_name;
    double              _scale = 1.0;
    bool                _try_flip = false;

    cv::CascadeClassifier _face_cascade;
    cv::CascadeClassifier _eyes_cascade;

    cv::Mat* detectAndDraw( cv::Mat *iframe );

public:
    FltHaarCascade();
    cv::Mat* filter(cv::Mat* iframe);
};


