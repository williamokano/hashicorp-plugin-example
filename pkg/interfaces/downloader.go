package interfaces

import "io"

// Downloader defines the interface for downloading plugins
//
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Downloader
type Downloader interface {
	// Download downloads a file from a URL
	Download(url string) (io.ReadCloser, error)
	
	// DownloadToFile downloads a file from a URL to a local path
	DownloadToFile(url, filepath string) error
	
	// GetLatestVersion queries for the latest version of a plugin
	GetLatestVersion(repo, pluginName string) (string, error)
}

// PluginRegistry defines the interface for plugin registry operations
//
//counterfeiter:generate . PluginRegistry
type PluginRegistry interface {
	// ListAvailable lists all available plugins from a repository
	ListAvailable(repo string) ([]PluginInfo, error)
	
	// SearchPlugins searches for plugins matching a query
	SearchPlugins(repo, query string) ([]PluginInfo, error)
	
	// GetPluginReleases gets all releases for a specific plugin
	GetPluginReleases(repo, pluginName string) ([]ReleaseInfo, error)
}

// PluginInstaller defines the interface for installing plugins
//
//counterfeiter:generate . PluginInstaller
type PluginInstaller interface {
	// Install installs a plugin from a repository
	Install(pluginName, version, repo string) error
	
	// Uninstall removes a plugin
	Uninstall(pluginName string) error
	
	// IsInstalled checks if a plugin is installed
	IsInstalled(pluginName string) bool
	
	// ExtractArchive extracts a tar.gz archive
	ExtractArchive(archivePath, destPath string) error
}

// PluginInfo represents basic plugin information
type PluginInfo struct {
	Name        string
	Version     string
	Description string
}

// ReleaseInfo represents a plugin release
type ReleaseInfo struct {
	Version     string
	URL         string
	Checksum    string
	PublishedAt string
}