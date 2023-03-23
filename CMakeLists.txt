cmake_minimum_required(VERSION 3.2)

project( redeye )
set(CMAKE_CXX_FLAGS "-std=c++17")

include(ExternalProject)

ExternalProject_Add(mjpeg-streamer
        GIT_REPOSITORY  git@github.com:nadjieb/cpp-mjpeg-streamer
        TEST_COMMAND    ""
)

ExternalProject_Add(cpp-httplib
        GIT_REPOSITORY  git@github.com:yhirose/cpp-httplib
        TEST_COMMAND    ""
)


add_subdirectory(src)



