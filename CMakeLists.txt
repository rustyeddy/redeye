cmake_minimum_required(VERSION 2.8)

project( redeye )

find_package( OpenCV REQUIRED )

include_directories( ${OpenCV_INCLUDE_DIRS} )

add_executable( redeye main.cc )

target_link_libraries( redeye ${OpenCV_LIBS} )
