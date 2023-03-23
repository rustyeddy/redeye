#pragma once

#include <string>
#include <opencv2/opencv.hpp>

using namespace std;

class Window
{
    string _name = "";
    bool _slider = false;
    int  _slider_pos = -1;

public:
    Window(string name, int w, int h);

    string get_name()   { return _name; }
    int get_frames() { return (int) _cap.get(cv::CAP_PROP_FRAME_COUNT); }
    int get_width()  { return (int) _cap.get(cv::CAP_PROP_FRAME_WIDTH); }
    int get_height() { return (int) _cap.get(cv::CAP_PROP_FRAME_HEIGHT); }
};

