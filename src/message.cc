#include "message.hh"

Message::Message(string topic, string value)
{
    _topic = topic;
    _value = value;

    //
    // parse the topic to extract some important
    // values
    //
    // redeye/camera/<host>/<camera>/cmd (play|pause)
    // redeye/camera/<host>/<camera>/filter <filter-name>
    // redeye/camera/<host>/<camera>/record <record-name>
    // redeye/camera/<host>/<camera>/snapshot <snapshot-name>
    //

    // 1. Split topic string into / topic elements
    // 2. set topic type of "camera"
    // 3. set the host
    // 4. set the cammera
    // 5. set the directive (cmd, filter, record or snapshot)
    // 6. set the value
}
