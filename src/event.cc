#include <unistd.h>

#include "event.hh"

void EventLoop::add(Message* msg)
{
    _messages.push(msg);
}

void EventLoop::loop()
{
    while (_running) {
        if (_messages.size()) {
            usleep(20);
        }

        Message *msg = _messages.front();
        msg->dump();

        // // XXX Add these bits to the main event loop
        // string pname = msg->get_player();
        // if (pname == "" ) {
        //     cerr << "Uknown player for topic: " << mqsg->topic << endl;
        //     return;
        // }

        // Player* player = video_players[pname];
        // if (player == NULL) {
        //     cerr << "Could not find player from message: " << pname << endl;
        //     return;
        // }

        // player->add_message(msg);
    }
}

void *event_loop(void *p)
{
    events.loop();
    return p;
}
