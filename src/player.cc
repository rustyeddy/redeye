#include <list>
#include <vector>
#include <opencv2/opencv.hpp>

#include "nadjieb/mjpeg_streamer.hpp"

#include "config.hh"
#include "externs.hh"
#include "filter.hh"
#include "mqtt.hh"
#include "player.hh"

using MJPEGStreamer = nadjieb::MJPEGStreamer;

using namespace cv;
using namespace std;

static FltFilters* get_filters()
{
    if ( filters == NULL ) {
        filters = new FltFilters();
    }
    return filters;
}

extern void mjpeg_iframe_q(cv::Mat& iframe);

Player::Player( string vname, string fname )
{
    Video* vid = new Video(vname);

    // XXX figure out better scheme for filter
    if ( fname != "" ) {
        set_filter( fname );        
    }
    _imgsrc = vid;

    // Subscribe to MQTT messages for this player
    mqtt->subscribe("redeye/player/" + config->id() + "/" + _name);
}

void Player::command_request(string s)
{
    cout << "Putting: " << s << " on command Q" << endl;
    _cmdlist.push_back(s);
}

void Player::record( )
{
    _recording = true;

    if ( _video_writer == NULL ) {
        _video_writer = get_video_writer();
    }
    assert( _video_writer );
}

void Player::stop()
{
    _recording = false;
    if ( _video_writer != NULL ) {
        delete _video_writer;
    }
}

void Player::add_message( Message* msg )
{
    _messageQ.push(msg);
}

void Player::check_commands( )
{
    
    // Check for incoming commands
    if ( _cmdlist.empty() ) {
        return;
    }

    // If there are incoming commands, handle them here
    string cmd = _cmdlist.back();
    _cmdlist.pop_back();

    if ( cmd == "snap" ) {

        // Save image to file.
        cout << "We have an iframe to save to file ... " << endl;
        //save_image( iframe );

    } else if ( cmd == "pause" ) {

        cout << "We are being paused ... " << endl;
        _paused = true;

    } else if ( cmd == "play" ) {

        cout << "Play has been pushed ... " << endl;
        _paused = false;

    } else if ( cmd == "record" ) {

        cout << "We have a frame from video to save ... " << endl;
        record();

    } else if ( cmd == "stop" ) {
        cout << "We have recieved a stop command " << endl;
        stop();

    } else {

        cerr << "We have no support for: " << cmd << endl;
    }
}

void Player::stream( cv::Mat* mat )
{
    std::vector<int> params = { cv::IMWRITE_JPEG_QUALITY, 90 };
    std::vector<uchar> buff_bgr;
    cv::imencode(".jpg", *mat, buff_bgr, params);

    _streamer.publish("/", std::string(buff_bgr.begin(), buff_bgr.end()));
}

void Player::play_loop()
{
    // cout << ">>> Play loop: " << _recording << endl;
    _playing = true;
    while ( _playing ) {

        if ( ! _messageQ.empty() ) {

            Message* msg = _messageQ.front();
            _messageQ.pop();

            cout << "Player Play Loop got a message\n";
            msg->dump();

            
        } else if ( ! _frameQ.empty() ) {
        
            cv::Mat* iframe = _frameQ.front();
            _frameQ.pop();

            // move this up
            if ( _filter ) {
                iframe = _filter->filter( iframe );
            }

            stream ( iframe );
            if ( ! _paused && _local_display ) {
                display( iframe );
            }
            if ( _recording ) {
                record();
            }
            delete iframe;

        } else {
            usleep(1000);
            continue;
        }

    }
}

void *play_loop( void *p )
{
    Player *player = (Player *)p;

    cout << "Found player PLay loop ! " << player->get_name() << endl;
    player->play_loop();
    cout << "PLay loop returning " << endl;
    return NULL;
}

void *play_video( void *p )
{
    Player *player = (Player *)p;

    cout << "Found player PLay loop ! " << player->get_name() << endl;
    player->play();
    cout << "PLay loop returning " << endl;
    return NULL;
}

void Player::play( )
{
    _recording = true;

    cout << "Player " << _name << endl;
    
    // Start the streamer 
    _streamer.start( config->get_mjpg_port() );
    _streaming = true;

    pthread_t t_playloop;
    pthread_create( &t_playloop, NULL, ::play_loop, this );
    while ( _recording ) {

	cv::Mat* iframe = _imgsrc->get_frame();
        if ( iframe == NULL || iframe->empty() ) {
            cout << "Iframe empty - stopping video..." << endl;
            _recording = false;
            continue;
        }

        int size = _frameQ.size();
        if ( size > _frameQ_max ) {
            _frameQ_max = size;
        }
        if ( size > 4 ) {
            _frameQ_dropped++;
            delete iframe;
            continue;
        }

        _frameQ.push( iframe );
    }

    pthread_join( t_playloop, NULL );

    _streamer.stop();
    cerr << "Video has stopped playing.. " << endl;
}

VideoWriter* Player::get_video_writer()
{
    if ( _video_writer == NULL ) {
        _video_writer = new VideoWriter("redeye-video.mp4",
                                        VideoWriter::fourcc('m', 'p', '4', 'v'),
                                        30.0,
                                        Size(640, 480),
                                        true);
    }
    return _video_writer;
}

int Player::save_image( Mat& img )
{
    std::vector<int> params = { cv::IMWRITE_JPEG_QUALITY, 90 };
    vector<int> compression_params;
    compression_params.push_back(IMWRITE_PNG_COMPRESSION);
    compression_params.push_back(9);

    int result = imwrite("redeye-image.png", img, compression_params);
    return result;
}

void Player::set_filter( string name )
{
    if ( _filter == NULL || name != _filter->Name() ) {
        filters = get_filters();
        assert( filters );

        cout << "Setting filter to " << name << endl;
        _filter = filters->get(name);

        if ( _filter == NULL ) {
            cerr << "filter fialed probably not known: " << name << endl;
        }
    }
}

void
Player::display( Mat* img )
{
    // imshow( _name, *img );
}

Player* 
Players::add(string name)
{
    cout << "Opening video source: " << name << endl;

    Player* player  = new Player( name );
    _players[name] = player;

    return player;
}

Player*
Players::get(string name)
{
    return _players[name];
}

void
Players::process_message(Message* msg)
{
    string name = msg->get_element(MessageVideoPlayer);
    if (name == "") {
        cerr << "Failed to find video with name: " << name << endl;
        return;
    }

    Player* player = _players[name];
    if (player == NULL) {
        cerr << "Failed to find video with name: " << name << endl;
        return;
    }

    cout << "Player is adding a message: " << endl;
    player->add_message(msg);
}


void
mouse_callback( int event, int x, int y, int flags, void *param )
{
    Filter *f = NULL;
    if (param != NULL) {
        f = (Filter *) param;        
        cout << "Mouse Event ~  " << event << ", x: " << x << ", y: " << y << endl;
    }

    switch ( event ) {
    case EVENT_MOUSEMOVE: 
        //cout << "Mouse Move " << endl;
        break;

    case EVENT_LBUTTONDOWN:
        cout << "Mouse left button down. " << endl;
        break;

    case EVENT_LBUTTONUP:
        cout << "Mouse left button up. " << endl;
        break;

    case EVENT_RBUTTONDOWN:
        cout << "Mouse right button down. " << endl;
        break;

    case EVENT_RBUTTONUP:
        cout << "Mouse right button up. " << endl;
        break;

    default:
        cout << "Mouse event unknown " << event << endl;
    }

}

