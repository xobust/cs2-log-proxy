package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Server struct {
		Port int `json:"port"`
	} `json:"server"`

	Storage struct {
		Type        string `json:"type"`
		Path        string `json:"path"`
		MaxFileSize int64  `json:"maxFileSize"`
		MaxFiles    int    `json:"maxFiles"`
	} `json:"storage"`

	Receivers []struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Config map[string]interface{} `json:"config"`
	} `json:"receivers"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = filepath.Join(os.Getenv("HOME"), ".cs2-log-manager", "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
