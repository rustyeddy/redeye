#pragma once

#include <map>
#include <string>
#include <opencv2/opencv.hpp>

using namespace std;
using namespace cv;

class Filter
{
protected:
    string      _name;

public:
    Filter(string n);

    string Name();       
    string to_string() { return _name; }

    // All derivative classes must implement the filter method to
    // simply transform an image into another one.
    virtual Mat* filter(Mat* iframe) = 0;
};
