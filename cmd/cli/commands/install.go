package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/pkg/config"
	"github.com/williamokano/hashicorp-plugin-example/pkg/download"
)

// NewInstallCommand creates the install command
func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install all plugins from plugins.json",
		Long: `Install all plugins specified in plugins.json.

This command reads plugins.json and downloads all specified plugins
to the .plugins/ directory, similar to 'npm install'.

If no plugins.json exists, it will suggest running 'plugin-cli init' first.`,
		Example: `  # Install all plugins from plugins.json
  plugin-cli install

  # Install and update lock file
  plugin-cli install --update-lock`,
		Args: cobra.NoArgs,
		RunE: runInstallAll,
	}

	// Add flags
	cmd.Flags().BoolP("force", "f", false, "Force reinstall all plugins (ignores lock file)")
	cmd.Flags().Bool("update-lock", true, "Update plugins.lock file")
	cmd.Flags().StringP("repo", "r", "williamokano/hashicorp-plugin-example", "Default GitHub repository")
	cmd.Flags().IntP("parallel", "p", 4, "Number of parallel downloads (1-10)")
	cmd.Flags().Bool("verify-checksums", true, "Verify checksums from lock file")
	cmd.Flags().Bool("ignore-lock", false, "Ignore lock file and download latest versions")

	return cmd
}

func runInstallAll(cmd *cobra.Command, args []string) error {
	// Check if project is initialized
	if !config.IsProjectInitialized() {
		fmt.Println("No plugins.json found in current directory")
		fmt.Println("")
		fmt.Println("To initialize a new project:")
		fmt.Println("  plugin-cli init")
		fmt.Println("")
		fmt.Println("To add plugins:")
		fmt.Println("  plugin-cli add <plugin-name>")
		return fmt.Errorf("project not initialized")
	}

	// Load plugins configuration
	cfg, err := config.LoadPluginsConfig()
	if err != nil {
		return fmt.Errorf("failed to load plugins.json: %w", err)
	}

	if len(cfg.Plugins) == 0 {
		fmt.Println("No plugins specified in plugins.json")
		fmt.Println("Add plugins with: plugin-cli add <plugin-name>")
		return nil
	}

	// Get flags
	force, _ := cmd.Flags().GetBool("force")
	updateLock, _ := cmd.Flags().GetBool("update-lock")
	repo, _ := cmd.Flags().GetString("repo")
	parallel, _ := cmd.Flags().GetInt("parallel")
	// verifyChecksums, _ := cmd.Flags().GetBool("verify-checksums")  // TODO: Implement checksum verification
	// ignoreLock, _ := cmd.Flags().GetBool("ignore-lock")  // TODO: Implement lock file checking

	// Limit parallel downloads
	if parallel < 1 {
		parallel = 1
	} else if parallel > 10 {
		parallel = 10
	}

	fmt.Printf("Installing %d plugin(s) from plugins.json", len(cfg.Plugins))
	if parallel > 1 {
		fmt.Printf(" (up to %d in parallel)", parallel)
	}
	fmt.Println()

	// Ensure .plugins directory exists
	pluginsDir := config.GetPluginsDirectory()
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Prepare download items
	var downloadItems []download.DownloadItem
	skipped := 0

	for pluginName, versionSpec := range cfg.Plugins {
		version := parseVersionSpec(versionSpec)
		pluginPath := filepath.Join(pluginsDir, pluginName)
		if runtime.GOOS == "windows" {
			pluginPath += ".exe"
		}

		// Check if already installed (unless force)
		if !force {
			if _, err := os.Stat(pluginPath); err == nil {
				fmt.Printf("  ✓ %s already installed (skipping)\n", pluginName)
				skipped++
				continue
			}
		}

		downloadItems = append(downloadItems, download.DownloadItem{
			Name:     pluginName,
			Version:  version,
			DestPath: pluginPath,
		})
	}

	if len(downloadItems) == 0 {
		fmt.Printf("\nAll plugins already installed (%d skipped)\n", skipped)
		return nil
	}

	// Create download queue with progress reporting
	queue := download.NewDownloadQueue(parallel)

	// Add progress callback
	queue.SetProgressCallback(func(completed, total int, current string) {
		if current != "" {
			fmt.Printf("[%d/%d] Downloading %s...\n", completed+1, total, current)
		}
	})

	// Add error callback
	failed := []string{}
	queue.SetErrorCallback(func(name string, err error) {
		fmt.Printf("  ✗ Failed to install %s: %v\n", name, err)
		failed = append(failed, name)
	})

	// Add all items to queue
	for _, item := range downloadItems {
		queue.Add(item)
	}

	// Execute downloads
	lock := &config.PluginsLock{
		Plugins: []config.PluginLockEntry{},
	}

	_ = queue.Execute(func(item download.DownloadItem) error {
		if err := installPluginWithItem(item, repo); err != nil {
			return err
		}

		fmt.Printf("  ✓ %s@%s installed successfully\n", item.Name, item.Version)

		// Add to lock file
		if updateLock {
			lock.Plugins = append(lock.Plugins, config.PluginLockEntry{
				Name:    item.Name,
				Version: item.Version,
				// URL and checksum would be filled by actual download
			})
		}

		return nil
	})

	// Save lock file if requested
	if updateLock && len(lock.Plugins) > 0 {
		if err := config.SavePluginsLock(lock); err != nil {
			fmt.Printf("Warning: Failed to save lock file: %v\n", err)
		}
	}

	// Summary
	installed := len(downloadItems) - len(failed)
	fmt.Println("")
	fmt.Printf("Installation complete: %d succeeded", installed)
	if skipped > 0 {
		fmt.Printf(", %d skipped", skipped)
	}
	if len(failed) > 0 {
		fmt.Printf(", %d failed\n", len(failed))
		fmt.Println("Failed plugins:")
		for _, name := range failed {
			fmt.Printf("  - %s\n", name)
		}
	} else {
		fmt.Println("")
	}

	return nil
}

func parseVersionSpec(spec string) string {
	// Remove version range prefixes for now
	// In a real implementation, this would handle semver ranges
	spec = strings.TrimPrefix(spec, "^")
	spec = strings.TrimPrefix(spec, "~")
	spec = strings.TrimPrefix(spec, ">=")
	spec = strings.TrimPrefix(spec, ">")
	spec = strings.TrimPrefix(spec, "<=")
	spec = strings.TrimPrefix(spec, "<")
	spec = strings.TrimPrefix(spec, "=")

	if spec == "" || spec == "*" {
		return "latest"
	}

	return spec
}

func installPluginWithItem(item download.DownloadItem, repo string) error {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Get actual version if "latest"
	actualVersion := item.Version
	if item.Version == "latest" {
		// Would query GitHub API for latest release
		actualVersion = "1.0.0"
	}

	// In development mode, copy from local bin if available
	localBinary := fmt.Sprintf("./bin/%s", item.Name)
	if _, err := os.Stat(localBinary); err == nil {
		fmt.Printf("  ℹ Using local binary from %s (development mode)\n", localBinary)
		input, err := os.ReadFile(localBinary)
		if err != nil {
			return fmt.Errorf("failed to read local binary: %w", err)
		}
		if err := os.WriteFile(item.DestPath, input, 0755); err != nil {
			return fmt.Errorf("failed to copy binary: %w", err)
		}
		return nil
	}

	fmt.Printf("  Downloading %s_%s_%s_%s...\n", item.Name, actualVersion, osName, archName)

	// In production, this would:
	// 1. Download the tar.gz from GitHub releases
	// 2. Verify checksum
	// 3. Extract to .plugins/
	// 4. Set executable permissions

	// For now, return a clear message
	return fmt.Errorf("GitHub releases not yet available (will work after first release)")
}
