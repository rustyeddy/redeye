#pragma once

#include <iostream>
#include <string>
#include <vector>

using namespace std;

//
// Messages are constructed re/<message-type>/<host>/...
//
// message-types are: "video"
//
class Message
{
private:
    string              _topic;
    string              _value;

    vector<string>      _elements;

public:
    Message(string topic, string value = "");

    void parse_topic();
    vector<string> topic_elements() { return _elements; }

    string get_topic()  { return _topic; }
    string get_value()  { return _value; }

    string get_element(int i);

    string get_type();
    string get_host();

    void dump();
};

// topic: re/<msg-type>/<host>
const u_int MessageElementRE            = 0;
const u_int MessageElementType          = 1;
const u_int MessageElementHost          = 2;

// topic: re/video/<host>/<player-name>/<command>
const u_int MessageVideoPlayer          = 4;
const u_int MessageVideoCmd             = 5;
