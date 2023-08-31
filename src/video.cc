#include <iostream>
#include <fstream>
#include <string>

#include <opencv2/opencv.hpp>
#include "nadjieb/mjpeg_streamer.hpp"

#include "config.hh"
#include "video.hh"

using MJPEGStreamer = nadjieb::MJPEGStreamer;
using namespace cv;
using namespace std;

std::string get_tegra_pipeline1(int width, int height, int fps) {
    return "nvarguscamerasrc sensor_id=0 ! video/x-raw(memory:NVMM), width=(int)1920, height=(int)1080, format=(string)NV12, framerate=(fraction)30/1 ! nvvidconv flip-method=0 ! video/x-raw, format=(string)BGRx ! videoconvert ! video/x-raw, format=(string)BGR ! appsink";
}

std::string gstreamer_pipeline (int sensor, int capture_width, int capture_height, int display_width, int display_height, int framerate, int flip_method) {
    return "nvarguscamerasrc sensor_id=" + std::to_string(sensor) + " ! video/x-raw(memory:NVMM), width=(int)" + std::to_string(capture_width) + ", height=(int)" +
           std::to_string(capture_height) + ", format=(string)NV12, framerate=(fraction)" + std::to_string(framerate) +
           "/1 ! nvvidconv flip-method=" + std::to_string(flip_method) + " ! video/x-raw, width=(int)" + std::to_string(display_width) + ", height=(int)" +
           std::to_string(display_height) + ", format=(string)BGRx ! videoconvert ! video/x-raw, format=(string)BGR ! appsink";
}

Video::Video( int devnum )
{
    _cap.open( devnum );
}

Video::Video( string camstr )
{
    cout << "Opening camstr " << camstr << endl;
    _name = camstr;

    if ( camstr == "tegra0" || camstr == "tegra1" ) {
        
	std::string pipeline = gstreamer_pipeline(1,
                                                  dims.capture_width,
						  dims.capture_height,
						  dims.display_width,
						  dims.display_height,
						  dims.framerate,
						  dims.flip_method);

	std::cout << "Using pipeline: \n\t";
	std::cout << "--------------------------------" << std::endl;
	std::cout << pipeline << std::endl;
	std::cout << "--------------------------------" << std::endl;

	// string t = get_tegra(1024, 768, 60);
	// cout << "TEGRA String: " << t << endl;
	_cap.open( pipeline, cv::CAP_GSTREAMER );
        
    } else if ( camstr == "dev0" ) {

        cout << "Opening camera device 0\n";
        _cap.open( 0 );

    } else if ( camstr == "dev1" ) {

        cout << "Opening camera device 1\n" ;
	_cap.open( 1 );

    }

    if ( !_cap.isOpened() ) {
        cerr << "ERROR - the camera is not open. exiting ... " << endl;
        exit(-3);
    }
}

cv::Mat* Video::get_frame()
{
    if ( !_cap.isOpened() ) {
        cerr << "ERROR - the camera is not open. exiting ... " << endl;
        exit(-3);
    }

    Mat* iframe = new cv::Mat();
    if (!_cap.read( *iframe )) {
	cerr << "ERROR - reading cap frame" << endl;
	return iframe;
    }
    return iframe;
}

