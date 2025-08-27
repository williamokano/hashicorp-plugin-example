package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PluginsConfig represents the plugins.json configuration file
type PluginsConfig struct {
	Plugins map[string]string `json:"plugins"` // name -> version
}

const PluginsConfigFile = "plugins.json"

// LoadPluginsConfig loads the plugins configuration from plugins.json
func LoadPluginsConfig() (*PluginsConfig, error) {
	configPath := PluginsConfigFile

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &PluginsConfig{
				Plugins: make(map[string]string),
			}, nil
		}
		return nil, fmt.Errorf("failed to read plugins config: %w", err)
	}

	var config PluginsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse plugins config: %w", err)
	}

	if config.Plugins == nil {
		config.Plugins = make(map[string]string)
	}

	return &config, nil
}

// SavePluginsConfig saves the plugins configuration to plugins.json
func SavePluginsConfig(config *PluginsConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugins config: %w", err)
	}

	if err := os.WriteFile(PluginsConfigFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write plugins config: %w", err)
	}

	return nil
}

// InitPluginsConfig creates a new plugins.json file if it doesn't exist
func InitPluginsConfig() error {
	// Check if file already exists
	if _, err := os.Stat(PluginsConfigFile); err == nil {
		return fmt.Errorf("plugins.json already exists")
	}

	// Create initial config
	config := &PluginsConfig{
		Plugins: make(map[string]string),
	}

	// Save the config
	if err := SavePluginsConfig(config); err != nil {
		return fmt.Errorf("failed to create plugins.json: %w", err)
	}

	// Create .plugins directory if it doesn't exist
	pluginsDir := ".plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create .plugins directory: %w", err)
	}

	return nil
}

// AddPlugin adds or updates a plugin in the configuration
func (c *PluginsConfig) AddPlugin(name, version string) {
	if c.Plugins == nil {
		c.Plugins = make(map[string]string)
	}
	c.Plugins[name] = version
}

// RemovePlugin removes a plugin from the configuration
func (c *PluginsConfig) RemovePlugin(name string) {
	delete(c.Plugins, name)
}

// GetPluginVersion returns the configured version for a plugin
func (c *PluginsConfig) GetPluginVersion(name string) (string, bool) {
	version, exists := c.Plugins[name]
	return version, exists
}

// PluginLockEntry represents an entry in the lock file
type PluginLockEntry struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Checksum string `json:"checksum"`
	URL      string `json:"url"`
}

// PluginsLock represents the plugins.lock file
type PluginsLock struct {
	Plugins []PluginLockEntry `json:"plugins"`
}

const PluginsLockFile = "plugins.lock"

// LoadPluginsLock loads the plugins lock file
func LoadPluginsLock() (*PluginsLock, error) {
	data, err := os.ReadFile(PluginsLockFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &PluginsLock{
				Plugins: []PluginLockEntry{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read plugins lock: %w", err)
	}

	var lock PluginsLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse plugins lock: %w", err)
	}

	return &lock, nil
}

// SavePluginsLock saves the plugins lock file
func SavePluginsLock(lock *PluginsLock) error {
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plugins lock: %w", err)
	}

	if err := os.WriteFile(PluginsLockFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write plugins lock: %w", err)
	}

	return nil
}

// IsProjectInitialized checks if the project has been initialized with plugins.json
func IsProjectInitialized() bool {
	_, err := os.Stat(PluginsConfigFile)
	return err == nil
}

// GetPluginsDirectory returns the path to the plugins directory
func GetPluginsDirectory() string {
	return filepath.Join(".", ".plugins")
}
