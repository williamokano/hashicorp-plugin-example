package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "plugin-cli",
	Short: "Event-driven plugin orchestration CLI",
	Long: `A CLI tool that processes events through a pipeline of plugins.
Each plugin decides whether to act and can modify the context for subsequent plugins.

This architecture enables:
  • Behavioral extension through plugins
  • Event processing pipelines
  • Auto-discovery of plugins
  • Version compatibility checking
  • GitHub-based distribution`,
	Version: "1.0.0",
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(
		NewVersionCommand(),
		NewPluginCommand(),
		NewProcessCommand(),
		NewInstallCommand(),
		NewSimulateCommand(),
	)

	// Global flags (if any)
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().String("config", "", "Config file path (default: ~/.config/plugin-cli/config.json)")
}
