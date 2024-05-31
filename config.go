package redeye

import (
	"fmt"

	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Configuration struct {
	HTTPAddr	string `json:"addr"`		// http address and port
	HTMLPath	string `json:"basepath"`	// html basepath
	MQTTBroker	string `json:"broker"`		// MQTT Broker
	VideoDevice int	   `json:video-device`	// Capture device

	Debug	 bool	`json:"debug"`
	ID		 string `json:"id"`
	Pipeline string `json:"pipeline"`
	Thumb    string `json:"thumb"`
	Vidsrc   string `json:"vidsrc"`
	Vidaddr  string `json:"vidaddr"`
}

var (
	Config Configuration
)

func (c *Configuration) Save(path string) (err error) {

	buf, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("Config Save [%s] failed json.Marshal config [%w]", path, err)
	}

	err = ioutil.WriteFile(path, buf, 0644)
	if err != nil {
		return fmt.Errorf("Config Save [%s] failed to save file: [%w]", path, err)
	}
	return err
}

// ServeHTTP provides the Web service for the configuration module
func (c Configuration) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(c)
}
