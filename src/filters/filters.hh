#pragma once 

#include "filter.hh"
#include "filter_default.hh"

class FltFilters
{
private:
    map<string,Filter*> _filters;

public:
    FltFilters();

    void add(string name, Filter* f);
    Filter* get(string name);

    string to_json();
};

extern FltFilters* filters;
