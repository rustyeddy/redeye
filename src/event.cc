#include <unistd.h>

#include "event.hh"
#include "player.hh"

void EventLoop::add(Message* msg)
{
    _messages.push(msg);
}

void EventLoop::loop()
{
    while (_running) {
        if (_messages.size() < 1) {
            usleep(20);
            continue;
        }

        Message *msg = _messages.front();
        _messages.pop();
        msg->dump();

        string msg_type = msg->get_type();
        if (msg_type == "video") {

            cout << "We have a video message" << endl;
            video_players.process_message(msg);

        } else {
            cout << "We have an unknown message type: "<<  msg_type << endl;
        }
    }
}

void *event_loop(void *p)
{
    events.loop();
    return p;
}

