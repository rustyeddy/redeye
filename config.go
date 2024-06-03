package redeye

import (
	"fmt"

	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Configuration struct {
	HTTPAddr    string `json:"addr"`       // http address and port
	HTMLPath    string `json:"basepath"`   // html basepath
	MQTTBroker  string `json:"broker"`     // MQTT Broker
	VideoDevice int    `json:video-device` // Capture device
	CascadeFile string `json:cascade-file`
	Pipeline    string `json:"pipeline"`

	ListFilters bool `json:"list-filters"` // List filters

	ID      string `json:"id"`
	Thumb   string `json:"thumb"`
	Debug   bool   `json:"debug"`

}

var (
	Config *Configuration = &Configuration{}
)

func GetConfig() *Configuration {
	return Config
}

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
