#pragma once

#include "../filter.hh"

class FltResize : public Filter
{
public:
    FltResize() : Filter("resize") {}
    bool init() { return true; }
    Mat* filter(Mat* iframe);
};
