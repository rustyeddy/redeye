#pragma once

#include "../filter.hh"

class FltResize : public Filter
{
public:
    FltResize() : Filter("resize") {}
    Mat* filter(Mat* iframe);
};
