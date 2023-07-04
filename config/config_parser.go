// Package internal for internal details
package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/validator.v2"
)

// Configuration struct to hold app configurations
type Configuration struct {
	Port    string `json:"port"  validate:"nonzero"`
	DBFile  string `json:"db_file"  validate:"nonzero"`
	Version string `json:"version" validate:"nonzero"`
}

// ReadConfFile read configurations of json file
func ReadConfFile(path string) (Configuration, error) {
	config := Configuration{}
	file, err := os.Open(path)
	if err != nil {
		return Configuration{}, fmt.Errorf("failed to open config file: %w", err)
	}

	dec := json.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		return Configuration{}, fmt.Errorf("failed to load config: %w", err)
	}

	return config, validator.Validate(config)
}
