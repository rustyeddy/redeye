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
            string name = msg->get_element(MessageVideoPlayer);
            if (name == "") {
                cerr << "Failed to find video with name: " << name << endl;
                return;
            }

            Player* player = video_players.get(name);
            if (player == NULL) {
                cerr << "Failed to find video with name: " << name << endl;
                return;
            }

            cout << "Player is adding a message: " << endl;
            player->add_message(msg);

            // video_players.process_message(msg);

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

