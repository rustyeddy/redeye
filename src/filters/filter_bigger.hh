#pragma once

#include "../filter.hh"

class FltBigger : public Filter
{
private:
    int         _x0 = -1, _y0 = -1;
    int         _x1 = -1, _y1 = -1;
    bool        _drawing = false;

public:
    FltBigger();
    Mat* filter(Mat* iframe);
};

extern void bigger_mouse_callback( int event, int x, int y, int flags, void *param );
