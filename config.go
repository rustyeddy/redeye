package redeye

import (
	"errors"
	"fmt"

	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

type Configuration struct {
	HTTPAddr    string `json:"addr"`         // http address and port
	HTMLPath    string `json:"basepath"`     // html basepath
	MQTTBroker       string `json:"broker"`       // MQTT broker URL
	MQTTTopicPrefix  string `json:"topic-prefix"` // MQTT topic namespace
	VideoDevice string `json:"video-device"` // Capture device: index, name, or path
	Image       string `json:"image"`        // Single image
	Video    string `json:"video"`
	RTSPUrl  string `json:"rtsp-url"`
	CascadeFile string `json:"cascade-file"`
	Pipeline    string `json:"pipeline"`
	WaitTime    int    `json:"wait-time"`

	ListFilters bool `json:"list-filters"` // List filters

	ID    string `json:"id"`
	Thumb string `json:"thumb"`
	Debug bool   `json:"debug"`
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

	err = os.WriteFile(path, buf, 0644)
	if err != nil {
		return fmt.Errorf("Config Save [%s] failed to save file: [%w]", path, err)
	}
	return err
}

func (c *Configuration) Load(path string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Config Load [%s] failed to read file: [%w]", path, err)
	}

	if err := json.Unmarshal(buf, c); err != nil {
		return fmt.Errorf("Config Load [%s] failed json.Unmarshal config [%w]", path, err)
	}

	return nil
}

func (c *Configuration) LoadDefault() error {
	paths := []string{"redeye.json"}
	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(homeDir, ".redeye.json"))
	}

	for _, path := range paths {
		err := c.Load(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// ServeHTTP provides the Web service for the configuration module
func (c Configuration) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	writeJSON(w, c)
}
