package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Plugins      []PluginConfig `json:"plugins"`
	PluginPaths  []string       `json:"plugin_paths"`
	AutoDownload bool           `json:"auto_download"`
}

type PluginConfig struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Version    string `json:"version"`
	Enabled    bool   `json:"enabled"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = getDefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Plugins:      []PluginConfig{},
				PluginPaths:  []string{},
				AutoDownload: false,
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func (c *Config) Save(path string) error {
	if path == "" {
		path = getDefaultConfigPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func getDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "plugin-config.json"
	}
	return filepath.Join(homeDir, ".config", "plugin-cli", "config.json")
}
