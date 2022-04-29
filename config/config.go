package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/adrg/xdg"
)

const (
	ConfigPath  = "mata/config.json"
	EnvEndpoint = "MATAROA_ENDPOINT"
	EnvKey      = "MATAROA_KEY"
)

type Config struct {
	Endpoint string `json:"endpoint"`
	Key      string `json:"key"`
}

func LoadConfig() (Config, error) {
	var err error
	if os.Getenv(EnvEndpoint) != "" && os.Getenv(EnvKey) != "" {
		return Config{
			Key:      os.Getenv(EnvKey),
			Endpoint: os.Getenv(EnvEndpoint),
		}, nil
	}

	var path string
	path = os.Getenv("MATAROA_CONFIG")
	if path == "" {
		path, err = xdg.ConfigFile(ConfigPath)
		if err != nil {
			return Config{}, fmt.Errorf("error finding config file: %s", err)
		}
	}

	f, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("error reading config file: %s", err)
	}

	var config Config
	err = json.Unmarshal(f, &config)
	if err != nil {
		return Config{}, fmt.Errorf("error unmarshaling config file: %s", err)
	}

	if config.Endpoint == "" {
		config.Endpoint = "https://mataroa.blog/api"
	}

	if config.Key == "" {
		return Config{}, fmt.Errorf(`'key' cannot be empty on 'config.json'`)
	}

	return config, nil
}
