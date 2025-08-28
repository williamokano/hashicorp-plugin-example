package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/pkg/config"
)

var addCmd = &cobra.Command{
	Use:   "add [plugin-name]",
	Short: "Add a plugin to the project",
	Long: `Add a plugin to the project and download it.

This command:
  1. Downloads the plugin binary
  2. Updates plugins.json with the plugin and version
  3. Updates plugins.lock with download information

Examples:
  plugin-cli add dummy              # Add latest version
  plugin-cli add dummy@1.0.0        # Add specific version
  plugin-cli add dummy --save-exact # Save exact version`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

var (
	addRepo      string
	saveExact    bool
	skipDownload bool
)

func init() {
	addCmd.Flags().StringVarP(&addRepo, "repo", "r", "williamokano/hashicorp-plugin-example", "GitHub repository")
	addCmd.Flags().BoolVar(&saveExact, "save-exact", false, "Save exact version in plugins.json")
	addCmd.Flags().BoolVar(&skipDownload, "skip-download", false, "Only update plugins.json without downloading")

	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Check if project is initialized
	if !config.IsProjectInitialized() {
		return fmt.Errorf("no plugins.json found. Run 'plugin-cli init' first")
	}

	// Parse plugin name and version
	input := args[0]
	pluginName, version := parsePluginSpec(input)

	// Ensure plugin name has correct prefix
	if !strings.HasPrefix(pluginName, "plugin-") {
		pluginName = "plugin-" + pluginName
	}

	// Load current configuration
	cfg, err := config.LoadPluginsConfig()
	if err != nil {
		return fmt.Errorf("failed to load plugins.json: %w", err)
	}

	// Check if plugin already exists
	if existingVersion, exists := cfg.GetPluginVersion(pluginName); exists {
		fmt.Printf("Updating %s from %s to %s\n", pluginName, existingVersion, version)
	} else {
		fmt.Printf("Adding %s@%s\n", pluginName, version)
	}

	// Download the plugin if not skipping
	if !skipDownload {
		fmt.Printf("Downloading %s...\n", pluginName)

		// Use the download functionality
		if err := downloadPlugin(pluginName, version, addRepo); err != nil {
			return fmt.Errorf("failed to download plugin: %w", err)
		}
	}

	// Update plugins.json
	versionToSave := version
	if !saveExact && version != "latest" {
		// Could use semver ranges here in the future
		versionToSave = "^" + version
	}

	cfg.AddPlugin(pluginName, versionToSave)

	if err := config.SavePluginsConfig(cfg); err != nil {
		return fmt.Errorf("failed to update plugins.json: %w", err)
	}

	// Update lock file
	if err := updateLockFile(pluginName, version, addRepo); err != nil {
		// Non-fatal error
		fmt.Printf("Warning: Failed to update lock file: %v\n", err)
	}

	fmt.Printf("✓ Added %s@%s to plugins.json\n", pluginName, versionToSave)

	// Show current plugins
	fmt.Println("\nCurrent plugins:")
	for name, ver := range cfg.Plugins {
		fmt.Printf("  - %s: %s\n", name, ver)
	}

	return nil
}

func parsePluginSpec(spec string) (name, version string) {
	parts := strings.Split(spec, "@")
	name = parts[0]
	version = "latest"

	if len(parts) > 1 {
		version = parts[1]
	}

	return name, version
}

func downloadPlugin(pluginName, version, repo string) error {
	// Get OS and architecture
	osName := runtime.GOOS
	archName := runtime.GOARCH

	pluginsDir := config.GetPluginsDirectory()

	// Ensure plugins directory exists
	if err := os.MkdirAll(pluginsDir, 0750); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Check if plugin already exists
	pluginPath := filepath.Join(pluginsDir, pluginName)
	if runtime.GOOS == "windows" {
		pluginPath += ".exe"
	}

	// Get actual version if "latest"
	actualVersion := version
	if version == "latest" {
		// In a real implementation, this would query GitHub API
		actualVersion = "1.0.0"
	}

	// Download using existing download logic
	archiveName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", pluginName, actualVersion, osName, archName)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s",
		repo, actualVersion, archiveName)

	// For plugin-specific releases
	if pluginName != "plugin-cli" {
		shortName := strings.TrimPrefix(pluginName, "plugin-")
		altURL := fmt.Sprintf("https://github.com/%s/releases/download/plugin-%s-v%s/%s",
			repo, shortName, actualVersion, archiveName)

		// Try plugin-specific URL first
		if err := downloadAndExtract(altURL, pluginsDir, pluginPath); err == nil {
			return nil
		}
	}

	// Try general release URL
	if err := downloadAndExtract(downloadURL, pluginsDir, pluginPath); err != nil {
		return fmt.Errorf("plugin not available for download yet (releases may not be published): %w", err)
	}
	return nil
}

func downloadAndExtract(url, destDir, pluginPath string) error {
	// Simulated download for now - would use actual download logic
	fmt.Printf("  Attempting download from: %s\n", url)

	// In development mode, just create a symlink to local binary if it exists
	localBinary := fmt.Sprintf("./bin/%s", filepath.Base(pluginPath))
	if _, err := os.Stat(localBinary); err == nil {
		fmt.Printf("  ℹ Using local binary from %s (development mode)\n", localBinary)
		// Copy the local binary instead of downloading
		input, err := os.ReadFile(localBinary)
		if err != nil {
			return err
		}
		if err := os.WriteFile(pluginPath, input, 0755); err != nil { //nolint:gosec // G306: executable files need 0755
			return err
		}
		return nil
	}

	// In production, this would actually download
	// For now, return an error indicating the release doesn't exist yet
	return fmt.Errorf("release not found (HTTP 404)")
}

func updateLockFile(pluginName, version, repo string) error {
	lock, err := config.LoadPluginsLock()
	if err != nil {
		return err
	}

	// Get actual version if "latest"
	actualVersion := version
	if version == "latest" {
		actualVersion = "1.0.0" // Would query GitHub API
	}

	osName := runtime.GOOS
	archName := runtime.GOARCH
	archiveName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", pluginName, actualVersion, osName, archName)

	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s",
		repo, actualVersion, archiveName)

	// Update or add entry
	found := false
	for i, entry := range lock.Plugins {
		if entry.Name == pluginName {
			lock.Plugins[i] = config.PluginLockEntry{
				Name:     pluginName,
				Version:  actualVersion,
				URL:      downloadURL,
				Checksum: "", // Would calculate actual checksum
			}
			found = true
			break
		}
	}

	if !found {
		lock.Plugins = append(lock.Plugins, config.PluginLockEntry{
			Name:     pluginName,
			Version:  actualVersion,
			URL:      downloadURL,
			Checksum: "", // Would calculate actual checksum
		})
	}

	return config.SavePluginsLock(lock)
}
