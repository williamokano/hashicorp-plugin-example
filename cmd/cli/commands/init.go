package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/pkg/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new plugins configuration",
	Long: `Initialize a new plugins.json configuration file in the current directory.

This command creates:
  - plugins.json: Plugin dependency configuration
  - .plugins/: Directory for plugin binaries

Similar to 'npm init', this sets up your project for plugin management.`,
	RunE: runInit,
}

var (
	forceInit bool
)

func init() {
	initCmd.Flags().BoolVarP(&forceInit, "force", "f", false, "Force initialization, overwrite existing files")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Check if already initialized
	if config.IsProjectInitialized() && !forceInit {
		fmt.Println("Project already initialized with plugins.json")
		fmt.Println("Use --force to reinitialize")
		return nil
	}

	// Remove existing files if force flag is set
	if forceInit {
		_ = os.Remove(config.PluginsConfigFile) // Best effort cleanup
		_ = os.Remove(config.PluginsLockFile)   // Best effort cleanup
	}

	// Initialize the configuration
	if err := config.InitPluginsConfig(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	fmt.Println("Initialized plugins configuration")
	fmt.Println("")
	fmt.Println("Created:")
	fmt.Println("  - plugins.json (plugin dependencies)")
	fmt.Println("  - .plugins/ (plugin binaries directory)")
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  plugin-cli add <plugin-name>    # Add a plugin")
	fmt.Println("  plugin-cli install              # Install all plugins from plugins.json")
	fmt.Println("  plugin-cli list                 # List installed plugins")
	fmt.Println("")
	fmt.Println("Note: Consider adding .plugins/ and plugins.lock to your .gitignore file")

	return nil
}
