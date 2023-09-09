#pragma once

#include <iostream>
#include <string>
#include <vector>

using namespace std;

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

    void dump() {
        cout << _topic << ": " << _value << endl;
    }
};
