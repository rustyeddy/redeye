cmake_minimum_required(VERSION 3.10)

add_subdirectory( filters )

find_package( OpenCV REQUIRED )
include_directories( ${OpenCV_INCLUDE_DIRS} )
add_executable( redeye
                buffers.cc
                config.cc
                cmd.cc
                filter.cc
                filters.cc
                image.cc
                main.cc
                message.cc
                mqtt.cc
                net.cc
                player.cc
                video.cc
                web.cc
                )       

set(CMAKE_C_FLAGS_DEBUG "${CMAKE_C_FLAGS_DEBUG} -ggdb -ffunction-sections")
set(CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} -ggdb -ffunction-sections")
set(CMAKE_BUILD_TYPE Debug)
                     
target_link_libraries( redeye ${OpenCV_LIBS} -lmosquitto -lpthread Filters )
