#pragma once

class FltNULL : public Filter {
public:
    FltNULL() : Filter("") {}
    bool init() { return true; }
    Mat* filter(Mat* iframe);
};

class FltSmaller : public Filter {
public:
    FltSmaller() : Filter("smaller") {}
    bool init() { return true; }
    Mat* filter(Mat* iframe);
};
    
class FltGaussianBlur : public Filter
{
public:
    FltGaussianBlur() : Filter("gaussian") {}
    bool init() { return true; }
    Mat* filter(Mat* iframe);
}; 
    
class FltCanny : public Filter
{
public:
    FltCanny() : Filter("canny") {}
    bool init() { return true; }
    Mat* filter(Mat* iframe);
};

class FltBorder : public Filter
{
  public:
    FltBorder() : Filter("border") {}
    bool init() { return true; }
    Mat* filter(Mat* iframe);
};
