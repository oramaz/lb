package util

import (
	"encoding/json"
	"log"
	"os"
)

// Config
type Config struct {
	Port  int      `json:"port"`  // Load-balancer server port
	Hosts []string `json:"hosts"` // Array of services' URLs
}

// New parses config.json file and returns the Config object
func New() (config *Config) {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal("opening config file", err)
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		log.Fatal("parsing config file", err.Error())
	}

	return
}
