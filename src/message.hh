#pragma once

#include <iostream>
#include <string>

using namespace std;

class Message
{
private:
    string      _topic;
    string      _value;
    string     _player;

public:
    Message(string topic, string value = "");

    string get_topic()  { return _topic; }
    string get_value()  { return _value; }

    string get_player() { return _player; }

    void dump() {
        cout << _topic << ": " << _value << endl;
    }
};
