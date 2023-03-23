#pragma once

#include <string>

#include <mosquitto.h>

int     mqtt_publish(std::string topic, std::string msg);
void*   mqtt_loop(void *p);
