#pragma once

#include <queue>

#include "message.hh"

class EventLoop
{
private:
    bool _running    = true;
    queue<Message*> _messages;

public:
    EventLoop() {}

    void add(Message* msg);

    void loop();
};

extern EventLoop events;
extern void* event_loop(void *p);
