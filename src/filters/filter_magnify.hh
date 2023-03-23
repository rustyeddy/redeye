#pragma once

#include "filter.hh"

class FltResize : public Filter
{
public:
    FltResize() : Filter("magnify") {}
    Mat* filter(Mat* iframe);
};
