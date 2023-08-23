cmake_minimum_required(VERSION 3.2)

project( redeye )
set(CMAKE_CXX_FLAGS "-std=c++17")
set(BUILD_MODE Debug)

include(ExternalProject)

ExternalProject_Add(mjpeg
        GIT_REPOSITORY  git@github.com:nadjieb/cpp-mjpeg-streamer
        TEST_COMMAND    ""
)

ExternalProject_Add(httplib
        GIT_REPOSITORY  git@github.com:yhirose/cpp-httplib
        TEST_COMMAND    ""
)

# ExternalProject_Add(json
#         GIT_REPOSITORY  git@github.com:nlohmann/json.git
#         TEST_COMMAND    ""
#         BUILD_COMMAND   ""
#)

add_subdirectory(src)



