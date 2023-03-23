#pragma once

class FltNULL : public Filter {
public:
    FltNULL() : Filter("") {}
    Mat* filter(Mat* iframe);
};

class FltSmaller : public Filter {
public:
    FltSmaller() : Filter("smaller") {}
    Mat* filter(Mat* iframe);
};
    
class FltGaussianBlur : public Filter
{
public:
    FltGaussianBlur() : Filter("gaussian") {}
    Mat* filter(Mat* iframe);
}; 
    
class FltCanny : public Filter
{
public:
    FltCanny() : Filter("canny") {}
    Mat* filter(Mat* iframe);
};

class FltBorder : public Filter
{
  public:
    FltBorder() : Filter("border") {}
    Mat* filter(Mat* iframe);
};
