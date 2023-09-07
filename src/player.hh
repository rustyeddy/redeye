#pragma once

#include <string>
#include <queue>

#include "nadjieb/mjpeg_streamer.hpp"

#include <opencv2/opencv.hpp>
#include "filter.hh"
#include "video.hh"
#include "message.hh"

using namespace std;
using namespace cv;
using MJPEGStreamer = nadjieb::MJPEGStreamer;

class Player
{
private:
    string              _name = "";
    list<string>        _windows;
    int                 _xpos = 100, _ypos = 100;

    Filter*             _filter = NULL;
    Imgsrc*             _imgsrc;

    list<string>        _cmdlist;
    VideoWriter*        _video_writer;

    MJPEGStreamer       _streamer;

    bool                _streaming = false;
    bool                _playing = false;
    bool                _recording = false;
    bool                _paused = false;
    bool                _local_display = false;

    queue<cv::Mat*>     _frameQ;

    int                 _frameQ_max = 0;
    int                 _frameQ_size = 0;
    int                 _frameQ_dropped = 0;

    queue<Message*>     _messageQ;;

public:
    Player( string name, string filter_name = "" );

    string      get_name() { return _name; }

    void        set_filter( string name );

    string      snapshot_filename()  { return "redeye-snapshot.png"; }
    string      video_filename()        { return "redeye-video.mp4"; }

    VideoWriter* get_video_writer();

    void        play();
    void        pause();

    bool        is_recording()  { return _recording; }
    void        record();
    void        stop();

    void        stream( Mat* frame );
    void        display( Mat* frame );
    int         save_image( Mat& frame );
    void        add_message( Message* msg );

    void        play_loop();
    void        command_request(string s);
    void	check_commands();
    string      to_string() { return _name; }
};

extern map<string, Player*> video_players;
extern void* play_video( void *p ); // callback for pthreads
extern void mouse_callback( int event, int x, int y, int flags, void *param );

