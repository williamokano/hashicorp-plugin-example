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

// NewRemoveCommand creates the remove command
func NewRemoveCommand() *cobra.Command {
	var keepBinary bool

	cmd := &cobra.Command{
		Use:   "remove [plugin-name]",
		Short: "Remove a plugin from the project",
		Long: `Remove a plugin from the project.

This command:
  1. Removes the plugin binary from .plugins/
  2. Updates plugins.json to remove the plugin
  3. Updates plugins.lock to remove the plugin entry

Examples:
  plugin-cli remove dummy
  plugin-cli remove plugin-filter`,
		Aliases: []string{"rm", "uninstall"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemove(cmd, args, keepBinary)
		},
	}

	cmd.Flags().BoolVar(&keepBinary, "keep-binary", false, "Keep the plugin binary in .plugins/")
	return cmd
}

func runRemove(_ *cobra.Command, args []string, keepBinary bool) error {
	// Check if project is initialized
	if !config.IsProjectInitialized() {
		return fmt.Errorf("no plugins.json found. Run 'plugin-cli init' first")
	}

	pluginName := args[0]

	// Ensure plugin name has correct prefix
	if !strings.HasPrefix(pluginName, "plugin-") {
		pluginName = "plugin-" + pluginName
	}

	// Load current configuration
	cfg, err := config.LoadPluginsConfig()
	if err != nil {
		return fmt.Errorf("failed to load plugins.json: %w", err)
	}

	// Check if plugin exists in configuration
	if _, exists := cfg.GetPluginVersion(pluginName); !exists {
		return fmt.Errorf("plugin %s is not in plugins.json", pluginName)
	}

	fmt.Printf("Removing %s...\n", pluginName)

	// Remove plugin binary if not keeping it
	if !keepBinary {
		pluginPath := filepath.Join(config.GetPluginsDirectory(), pluginName)
		if runtime.GOOS == "windows" {
			pluginPath += ".exe"
		}

		if err := os.Remove(pluginPath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("Warning: Failed to remove plugin binary: %v\n", err)
			}
		} else {
			fmt.Printf("✓ Removed plugin binary from .plugins/\n")
		}
	}

	// Update plugins.json
	cfg.RemovePlugin(pluginName)

	if err := config.SavePluginsConfig(cfg); err != nil {
		return fmt.Errorf("failed to update plugins.json: %w", err)
	}
	fmt.Printf("✓ Removed %s from plugins.json\n", pluginName)

	// Update lock file
	if err := removeFromLockFile(pluginName); err != nil {
		// Non-fatal error
		fmt.Printf("Warning: Failed to update lock file: %v\n", err)
	} else {
		fmt.Printf("✓ Updated plugins.lock\n")
	}

	// Show remaining plugins
	if len(cfg.Plugins) > 0 {
		fmt.Println("\nRemaining plugins:")
		for name, ver := range cfg.Plugins {
			fmt.Printf("  - %s: %s\n", name, ver)
		}
	} else {
		fmt.Println("\nNo plugins remaining in plugins.json")
	}

	return nil
}

func removeFromLockFile(pluginName string) error {
	lock, err := config.LoadPluginsLock()
	if err != nil {
		return err
	}

	// Remove entry
	newPlugins := []config.PluginLockEntry{}
	for _, entry := range lock.Plugins {
		if entry.Name != pluginName {
			newPlugins = append(newPlugins, entry)
		}
	}

	lock.Plugins = newPlugins
	return config.SavePluginsLock(lock)
}
