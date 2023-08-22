#pragma once

class FltContour : public Filter
{
public:
    FltContour();
    int get_threshold() { return 100; }
    bool init() { return true; }
    Mat* filter(Mat* iframe);
};

