#include <iostream>

#include "cmd.hh"
#include "player.hh"

using namespace std;

// Brute force it for now
void cmd_runner( char *cmdstr )
{
    string cmd(cmdstr);

    if ( cmd == "snap" ) {

        cout << "CMD snap " << endl;
        player->command_request("snap");

    } else if ( cmd == "filter" ) {

        int pos = cmd.find_first_of(' ');
        if (pos == string::npos) {
            cerr << "WARN filter command with not filter name. " << endl;
            return;
        }

        string fname = cmd.substr(pos);
        if (fname == "") {
            cerr << "WARN CMD filter failed to get filter name. " << endl;
            return;
        }

        cout << "CMD filter " << endl;
        player->set_filter( fname );

    } else if ( cmd == "record" ) {

        cout << "CMD record " << endl;            
        player->command_request("record");

    } else if ( cmd == "pause" ) {

        cout << "CMD pause " << endl;
        player->command_request("pause");

    } else if ( cmd == "stop" ) {

        cout << "CMD stop " << endl;
        player->command_request("stop");

    } else if ( cmd == "stop" ) {

        cout << "CMD stop " << endl;

    } else {

        cerr << "Error unsupported command: " << cmd << endl;

    }
}
