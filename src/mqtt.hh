#pragma once

#include <string>
#include <vector>
#include <mosquitto.h>

#include "player.hh"

class MQTT
{
private:
    struct mosquitto *_mosq = NULL;

    string _broker;
    string _ID;

public:
    MQTT(string broker);

    int subscribe(string Topic);
    int publish(string Topic, string msg);

    void message_handler(string topic, string msg);
    void loop(void *p);
};

// redeye/player/<hostid>/<playerid>
class Topic
{
    string              _player_name;
    string              _cmd;
    vector<string>      _items;

public:
    Topic(string tstr);
    
    string cmd() { return _cmd; }
    Player *player();

    vector<string> items();
};

void *mqtt_loop(void *p);

extern MQTT *mqtt;
