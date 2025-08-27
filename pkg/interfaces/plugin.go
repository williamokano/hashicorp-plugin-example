package interfaces

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/williamokano/hashicorp-plugin-example/pkg/discovery"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . PluginManager
type PluginManager interface {
	LoadPlugin(name string) (*plugin.Client, types.VersionedPlugin, error)
	LoadPluginFromPath(path string) (*plugin.Client, types.VersionedPlugin, error)
	ListPlugins() ([]discovery.DiscoveredPlugin, error)
	GetPluginMetadata(p types.VersionedPlugin) types.PluginMetadata
}

//counterfeiter:generate . PluginDiscovery
type PluginDiscovery interface {
	DiscoverPlugins(paths []string) ([]discovery.DiscoveredPlugin, error)
	FindPlugin(name string) (*discovery.DiscoveredPlugin, error)
	GetPluginPaths() []string
}

//counterfeiter:generate . Pipeline
type Pipeline interface {
	ProcessEvent(ctx context.Context, event types.Event) (*types.Context, error)
	ProcessMessage(ctx context.Context, source, content, userID, channelID string) (*types.Context, error)
	ProcessCommand(ctx context.Context, source, command, userID, channelID string) (*types.Context, error)
}

//counterfeiter:generate . PackageManager
type PackageManager interface {
	Install(repository, version string) error
	Remove(pluginName string) error
	List() ([]string, error)
}

//counterfeiter:generate . VersionChecker
type VersionChecker interface {
	IsCompatible(cliVersion, minVersion, maxVersion string) (bool, error)
	Parse(version string) (*Version, error)
	Compare(v1, v2 *Version) int
}

// Version represents a semantic version
type Version struct {
	Major int
	Minor int
	Patch int
}

//counterfeiter:generate . ConfigManager
type ConfigManager interface {
	Load(path string) (*Config, error)
	Save(config *Config, path string) error
	GetDefaultPath() string
}

// Config represents the application configuration
type Config struct {
	Plugins      []PluginConfig `json:"plugins"`
	PluginPaths  []string       `json:"plugin_paths"`
	AutoDownload bool           `json:"auto_download"`
}

// PluginConfig represents a plugin configuration
type PluginConfig struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Version    string `json:"version"`
	Enabled    bool   `json:"enabled"`
}
