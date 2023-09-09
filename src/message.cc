#include "message.hh"

Message::Message(string topic, string value)
{
    _topic = topic;
    _value = value;

    parse_topic();

    //
    // parse the topic to extract some important
    // values
    //
    // redeye/video/<host>/<camera>/cmd (play|pause)
    // redeye/video/<host>/<camera>/filter <filter-name>
    // redeye/video/<host>/<camera>/record <record-name>
    // redeye/video/<host>/<camera>/snapshot <snapshot-name>
    //

    // 1. Split topic string into / topic elements
    // 2. set topic type of "camera"
    // 3. set the host
    // 4. set the cammera
    // 5. set the directive (cmd, filter, record or snapshot)
    // 6. set the value
}

void Message::parse_topic()
{
    // parse the topic into elements
    size_t start = 0;
    size_t end = 0;
    string token;

    while (end != string::npos) {

        end = _topic.find('/', start);
        token = _topic.substr(start, end - start);
        _elements.push_back(token);
        start = end + 1;
    }
}

void Message::dump()
{
    for (int i = 0; i < _elements.size(); i++) {
        if ( i == 0 ) {
            cout << _elements[i];
        } else {
            cout << " :: " << _elements[i];             
        }
    }

    cout << " = " << _value << endl;
}

string Message::get_element(int i)
{
    if (_elements.size() < i) {
        return "";
    }
    return _elements[i-1];
}

string Message::get_type()
{
    if (_elements.size() >= 2) {
        return _elements[1];
    }
    return "";
}

string Message::get_host()
{
    if (_elements.size() >= 3) {
        return _elements[2];
    }
    return "";
}

