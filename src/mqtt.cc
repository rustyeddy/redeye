#include <string>
#include <iostream>
#include <string.h>
#include <assert.h>

#include <mosquitto.h>

#include "config.hh"
#include "cmd.hh"
#include "message.hh"
#include "mqtt.hh"
#include "player.hh"

using namespace std;


static void mqtt_log_callback(struct mosquitto *mosq, void *userdata, int level, const char *str)
{
    clog << "MQTT" << str << endl;
}

void* mqtt_loop(void *p)
{
    mqtt->loop(p);
    return p;
}

static void mqtt_connect_callback(struct mosquitto *mosq, void *userdata, int result)
{
    int i;
    if( result ) {
        fprintf(stderr, "Connect failed\n");
        return;
    }
    /* Subscribe to broker information topics on successful connect. */
    string id((char *) userdata);
    string tbase = "re/camera/" + id + "/+/";

    string topic(tbase + "cmd");
    mosquitto_subscribe(mosq, NULL, topic.c_str(), 2);
    topic = tbase + "filter";
    mosquitto_subscribe(mosq, NULL, topic.c_str(), 2);
    
    mqtt->publish("re/announce/camera", id);
}

static void mqtt_message_callback(struct mosquitto *mosq, void *obj, const struct mosquitto_message *mqsg)
{
    bool match = 0;
    printf("MQTT Message topic: %s - %d - %s\n", mqsg->topic, mqsg->payloadlen, (char *) mqsg->payload);

    Message *msg = (mqsg->payloadlen) ?
        new Message(string(mqsg->topic), string((char *)mqsg->payload)) :
        new Message(string(mqsg->topic));

    string pname = msg->get_player();
    if (pname == "" ) {
        cerr << "Uknown player for topic: " << mqsg->topic << endl;
        return;
    }

    Player* player = video_players[pname];
    if (player == NULL) {
        cerr << "Could not find player from message: " << pname << endl;
        return;
    }

    player->add_message(msg);
}


MQTT::MQTT(string broker)
{
    _broker = broker;
}

int MQTT::publish(string topic, string msg) 
{
    int mid;
    return mosquitto_publish( _mosq, NULL, topic.c_str(), topic.length(), msg.c_str(), 0, false );
}

int MQTT::subscribe(string topic)
{
    
    return 0;
}

void MQTT::loop(void *p)
{
    int i;
    int keepalive = 60;
    bool clean_session = true;

    mosquitto_lib_init();
    _mosq = mosquitto_new(NULL, clean_session, NULL);
    if (!_mosq) {
        cerr << "MQTT New Error: Out of memory." << endl;
        return;
    }
    mosquitto_log_callback_set(_mosq, mqtt_log_callback);
    mosquitto_connect_callback_set(_mosq, mqtt_connect_callback);
    mosquitto_message_callback_set(_mosq, mqtt_message_callback);
    // mosquitto_subscribe_callback_set(_mosq, mqtt_subscribe_callback);

    // string broker = config->get_mqtt_broker();
    if (mosquitto_connect(_mosq, _broker.c_str(), 1883, keepalive)) {
        cerr << "MQTT Error: Failed to connect." << endl;
        return;
    }

    mosquitto_loop_forever(_mosq, -1, 1);

    mosquitto_destroy(_mosq);
    mosquitto_lib_cleanup();
}

Topic::Topic(string tstr)
{
    int start = 0;
    for (int i = 0; i < tstr.length(); i++) {
        if (tstr[i] == '/') {
            _items.push_back(tstr.substr(start, i - start));
            start = i;
        }
    }
}

Player *Topic::player()
{
    Player* p = video_players[_player_name];
    return p;
}
