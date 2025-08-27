package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	PluginPrefix = "plugin-"
)

type DiscoveredPlugin struct {
	Name string
	Path string
}

func DiscoverPlugins(paths []string) ([]DiscoveredPlugin, error) {
	var plugins []DiscoveredPlugin

	for _, searchPath := range paths {
		absPath, err := filepath.Abs(searchPath)
		if err != nil {
			continue
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(absPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()

			if runtime.GOOS == "windows" {
				if !strings.HasSuffix(name, ".exe") {
					continue
				}
				name = strings.TrimSuffix(name, ".exe")
			}

			if !strings.HasPrefix(name, PluginPrefix) {
				continue
			}

			pluginPath := filepath.Join(absPath, entry.Name())

			info, err := os.Stat(pluginPath)
			if err != nil {
				continue
			}

			if info.Mode()&0111 == 0 {
				continue
			}

			pluginName := strings.TrimPrefix(name, PluginPrefix)
			plugins = append(plugins, DiscoveredPlugin{
				Name: pluginName,
				Path: pluginPath,
			})
		}
	}

	return plugins, nil
}

func GetPluginPaths() []string {
	paths := []string{}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(homeDir, ".local", "share", "plugins"))
	}

	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, "plugins"))
		paths = append(paths, filepath.Join(cwd, ".plugins"))
	}

	if envPath := os.Getenv("PLUGIN_PATH"); envPath != "" {
		for _, p := range strings.Split(envPath, string(os.PathListSeparator)) {
			if p != "" {
				paths = append(paths, p)
			}
		}
	}

	paths = append(paths, "/usr/local/lib/plugins")

	return paths
}

func FindPlugin(name string) (*DiscoveredPlugin, error) {
	plugins, err := DiscoverPlugins(GetPluginPaths())
	if err != nil {
		return nil, err
	}

	for _, plugin := range plugins {
		if plugin.Name == name {
			return &plugin, nil
		}
	}

	return nil, fmt.Errorf("plugin '%s' not found", name)
}
