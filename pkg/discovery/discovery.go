package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	osWindows = "windows"
	exeSuffix = ".exe"
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

			if runtime.GOOS == osWindows {
				if !strings.HasSuffix(name, exeSuffix) {
					continue
				}
				name = strings.TrimSuffix(name, exeSuffix)
			}

			if !strings.HasPrefix(name, PluginPrefix) {
				continue
			}

			pluginPath := filepath.Join(absPath, entry.Name())

			info, err := os.Stat(pluginPath)
			if err != nil {
				continue
			}

			// Check if file is executable
			if runtime.GOOS == osWindows {
				// On Windows, we rely on the .exe extension check above
				// All .exe files are considered executable
			} else {
				// On Unix-like systems, check executable permissions
				if info.Mode()&0o111 == 0 {
					continue
				}
			}

			pluginName := strings.TrimPrefix(name, PluginPrefix)

			// Skip the CLI itself (plugin-cli is not a plugin)
			if pluginName == "cli" {
				continue
			}

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

	// Priority 1: Local .plugins directory (like .terraform)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, ".plugins"))
	}

	// Priority 2: Environment variable paths
	if envPath := os.Getenv("PLUGIN_PATH"); envPath != "" {
		for _, p := range strings.Split(envPath, string(os.PathListSeparator)) {
			if p != "" {
				paths = append(paths, p)
			}
		}
	}

	// Priority 3: Local plugins directory (for development)
	if cwd, err := os.Getwd(); err == nil {
		paths = append(paths, filepath.Join(cwd, "plugins"))
	}

	// Priority 4: User home directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(homeDir, ".local", "share", "plugins"))
	}

	// Priority 5: System-wide location
	if runtime.GOOS == osWindows {
		// On Windows, use ProgramData for system-wide plugins
		paths = append(paths, filepath.Join(os.Getenv("ProgramData"), "plugins"))
	} else {
		// On Unix-like systems
		paths = append(paths, "/usr/local/lib/plugins")
	}

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
