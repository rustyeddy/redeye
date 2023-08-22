#include <unistd.h>

#include "filter_face_detect.hh"

using namespace std;

FltHaarCascade::FltHaarCascade() : Filter("face-detect")
{
    _face_cascade_name = "/usr/local/share/opencv4/haarcascades/haarcascade_frontalface_alt.xml";
    _eyes_cascade_name = "/usr/local/share/opencv4/haarcascades/haarcascade_eye_tree_eyeglasses.xml";

    if (!_face_cascade.load(_face_cascade_name)) {
        cerr << "WARNING: Could not load Face cascade for nested objects" << endl;
        return;
    }

    if (!_eyes_cascade.load(_eyes_cascade_name)) {
        cerr << "WARNING: Could not load Eyes cascade for nested objects" << endl;
        return;
    }

    _filter_ok = true;
}

cv::Mat* FltHaarCascade::filter(cv::Mat* iframe)
{
    if ( _filter_ok == false ) {
        cerr << "Filter not ready" << endl;
        return iframe;
    }

    if ( iframe->empty() ) {
        cerr << "Haar Cascasde filter frame is empty" << endl;
        return iframe;
    }
    
    detectAndDraw( iframe );
    return iframe;
}

cv::Mat* FltHaarCascade::detectAndDraw( cv::Mat* img )
{
    double t = 0;
    vector<Rect> faces, faces2;
    const static Scalar colors[] =
    {
        Scalar(255,0,0),
        Scalar(255,128,0),
        Scalar(255,255,0),
        Scalar(0,255,0),
        Scalar(0,128,255),
        Scalar(0,255,255),
        Scalar(0,0,255),
        Scalar(255,0,255)
    };

    cv::Mat gray, smallImg;
    cvtColor( *img, gray, COLOR_BGR2GRAY );
    double fx = 1 / _scale;
    resize( gray, smallImg, Size(), fx, fx, INTER_LINEAR_EXACT );
    equalizeHist( smallImg, smallImg );

    t = (double) getTickCount();

    _face_cascade.detectMultiScale( smallImg, faces, 1.1, 2, 0|CASCADE_SCALE_IMAGE, Size(30, 30), Size(30, 30) );
    if ( _try_flip ) {
        flip( smallImg, smallImg, 1 );
        _face_cascade.detectMultiScale( smallImg, faces2, 1.1, 2, 0|CASCADE_SCALE_IMAGE, Size(30, 30), Size(30, 30) );
        for ( vector<Rect>::const_iterator r = faces2.begin(); r != faces2.end(); ++r ) {
            faces.push_back( Rect( smallImg.cols - r->x - r->width, r->y, r->width, r->height ) );
        }
    }

    t = (double) getTickCount() - t;
    printf("Detection time: %g ms\n", t*1000/getTickFrequency());

    for ( size_t i = 0; i < faces.size(); i++ ) {
        Rect r = faces[i];
        cv::Mat smallImgROI;
        vector<Rect> nestedObjects;
        Point center;
        Scalar color = colors[i%8];
        int radius;

        double aspect_ratio = (double) r.width / r.height;
        if ( 0.75 < aspect_ratio && aspect_ratio < 1.3 ) {
            center.x = cvRound( (r.x + r.width * 0.5) * _scale );
            center.y = cvRound( (r.y + r.height * 0.5) * _scale );            
            radius = cvRound( (r.width + r.height) * 0.25 * _scale );
            circle( *img, center, radius, color, 3, 8, 0 );
        } else {
            rectangle( *img,
                       Point( cvRound( r.x * _scale ), cvRound( r.y * _scale ) ),
                       Point( cvRound(( r.x + r.width-1 ) * _scale ), cvRound(( r.y + r.height-1 ) * _scale) ),
                       color, 3, 8, 0 );
        }
        if ( _eyes_cascade.empty() ) {
            cout << "Eyes cascade is empty continuing ... " << endl;
            continue;
        }

        smallImgROI = smallImg( r );
        _eyes_cascade.detectMultiScale( smallImgROI, nestedObjects, 1.1, 2, 0|CASCADE_SCALE_IMAGE, Size(30, 30), Size(30, 30) );

        cout << "Nested object count: " << ::to_string(nestedObjects.size()) << endl;
        for ( size_t j = 0; j < nestedObjects.size(); j++ ){
            Rect nr = nestedObjects[j];
            center.x = cvRound((r.x + nr.x + nr.width * 0.5) * _scale );
            center.y = cvRound((r.y + nr.y + nr.height * 0.5) * _scale );
            radius = cvRound( (nr.width + nr.height) * 0.25 * _scale );
            circle( *img, center, radius, color, 3, 8, 0 );
        }
    }
    return img;
}
