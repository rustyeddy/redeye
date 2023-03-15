cmake_minimum_required(VERSION 3.2)

project( redeye )
add_executable( redeye main.cc )

find_package( OpenCV REQUIRED )
include_directories( ${OpenCV_INCLUDE_DIRS} )

target_link_libraries( redeye ${OpenCV_LIBS} )
