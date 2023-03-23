#pragma once

class FltContour : public Filter
{
public:
    FltContour();
    Mat* filter(Mat* iframe);

    int get_threshold() { return 100; }
};

