cmake_minimum_required(VERSION 3.2)

project( filters )

find_package( OpenCV REQUIRED )
include_directories( ${OpenCV_INCLUDE_DIRS} )
add_library( Filters
	filter_bigger.cc
	filter_contour.cc
	filter_default.cc
	filter_face_detect.cc
	filter_resize.cc
)

set(CMAKE_BUILD_TYPE Debug)
