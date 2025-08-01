package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFile = ".task-cli-config.json"

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configFile)
}

func loadConfig() (*Config, error) {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config := &Config{}
			saveConfig(config)
			return config, nil
		}
		return nil, err
	}
	
	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func saveConfig(config *Config) error {
	configPath := getConfigPath()
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
