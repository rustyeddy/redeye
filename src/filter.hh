#pragma once

#include <map>
#include <string>
#include <opencv2/opencv.hpp>
#include <nlohmann/json.hpp>

using namespace std;
using namespace cv;
using json = nlohmann::json;

class Filter
{
protected:
    string      _name;
    json        _config = NULL;

public:
    Filter(string n);

    string Name();       
    string to_string() { return _name; }

    // Some derivative class can override the init function
    virtual bool init() { return true; }

    // All derivative classes must implement the filter method to
    // simply transform an image into another one.
    virtual Mat* filter(Mat* iframe) = 0;
};
