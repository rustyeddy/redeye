#include "filter.hh"
#include "filter_default.hh"
#include "filter_bigger.hh"
#include "filter_contour.hh"
#include "filter_face_detect.hh"
#include "filter_resize.hh"

#include "filters.hh"

FltFilters::FltFilters() {
    add( "border", new FltBorder() );
    add( "bigger", new FltBigger() );
    add( "contour", new FltContour() ); 
    add( "canny", new FltCanny() );
    add( "face-detect", new FltHaarCascade() );
    add( "gaussian", new FltGaussianBlur() );
    add( "smaller", new FltSmaller() );
    add( "resize", new FltResize() );
}

void FltFilters::add(string name, Filter* f)
{
    _filters[name] = f;
}

Filter* FltFilters::get(string name)
{
    auto it = _filters.find(name);
    if (it != _filters.end()) {
        Filter *filter = (Filter*) it->second;
        return filter;
    }
    return NULL;
}

string FltFilters::to_json()
{
    string jstr = "{[";
    bool first = false;

    std::map<std::string, Filter*>::iterator it = _filters.begin();
    while ( it != _filters.end() ) {
        if (!first) {
            jstr += ",";
        } else {
            first = false;
        }
        jstr += "\"" + it->first + "\"";
        it++;
    }
    jstr += "]}";
    return jstr;
}
