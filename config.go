package main

import (
	"encoding/json"
	"os"
)

// Configuration holds information about application configuration.
type Configuration struct {
	RedisHostName   string
	RedisPortNumber int32

	DNSServerAddr         string
	DNSPortNumber         int32
	DNSCacheTime          int64
	UpstreamDNSServerAddr string
}

// Config holds configuration values.
var Config Configuration

// ReadConfigurationFile builds configuration instance from json config file.
// This method must be called at the beginning of main method.
func ReadConfigurationFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	var config Configuration
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	Config = config

	return nil
}
