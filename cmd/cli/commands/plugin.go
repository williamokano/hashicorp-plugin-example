package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/pkg/discovery"
	"github.com/williamokano/hashicorp-plugin-example/pkg/manager"
	"github.com/williamokano/hashicorp-plugin-example/pkg/plugin"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

// NewPluginCommand creates the plugin management command
func NewPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
		Long: `Plugin management commands for listing, inspecting, and removing plugins.

Available subcommands:
  list    - List all discovered plugins
  info    - Show detailed information about a plugin
  remove  - Remove an installed plugin
  paths   - Show plugin discovery paths`,
	}

	cmd.AddCommand(
		newPluginListCommand(),
		newPluginInfoCommand(),
		newPluginRemoveCommand(),
		newPluginPathsCommand(),
	)

	return cmd
}

func newPluginListCommand() *cobra.Command {
	var showPaths bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discovered plugins with their priorities",
		Long: `List all plugins found in the configured discovery paths.
Shows plugin name, priority, version, and description.`,
		Example: `  plugin-cli plugin list
  plugin-cli plugin list --show-paths`,
		RunE: func(cmd *cobra.Command, args []string) error {
			plugins, err := discovery.DiscoverPlugins(discovery.GetPluginPaths())
			if err != nil {
				return fmt.Errorf("failed to discover plugins: %w", err)
			}

			if len(plugins) == 0 {
				fmt.Println("No plugins found.")
				if showPaths {
					fmt.Printf("\nSearch paths:\n")
					for _, path := range discovery.GetPluginPaths() {
						fmt.Printf("  - %s\n", path)
					}
				}
				return nil
			}

			// Load each plugin to get metadata
			mgr := plugin.NewManager()
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tPRIORITY\tVERSION\tDESCRIPTION")
			_, _ = fmt.Fprintln(w, "----\t--------\t-------\t-----------")

			for _, p := range plugins {
				client, plugin, err := mgr.LoadPluginFromPath(p.Path)
				if err != nil {
					_, _ = fmt.Fprintf(w, "%s\t?\t?\tError: %v\n", p.Name, err)
					continue
				}

				_, _ = fmt.Fprintf(w, "%s\t%d\t%s\t%s\n",
					plugin.Name(),
					plugin.Priority(),
					plugin.Version(),
					plugin.Description())

				client.Kill()
			}
			_ = w.Flush() // Best effort

			if showPaths {
				fmt.Printf("\nPlugin paths:\n")
				for _, path := range discovery.GetPluginPaths() {
					fmt.Printf("  - %s\n", path)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showPaths, "show-paths", false, "Show plugin discovery paths")

	return cmd
}

func newPluginInfoCommand() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "info [plugin-name]",
		Short: "Show detailed plugin information",
		Long:  `Display comprehensive metadata about a specific plugin including version requirements and description.`,
		Example: `  plugin-cli plugin info dummy
  plugin-cli plugin info filter --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]

			mgr := plugin.NewManager()
			client, p, err := mgr.LoadPlugin(pluginName)
			if err != nil {
				return fmt.Errorf("failed to load plugin: %w", err)
			}
			defer client.Kill()

			metadata := types.PluginMetadata{
				Name:          p.Name(),
				Version:       p.Version(),
				BuildTime:     p.BuildTime(),
				MinCLIVersion: p.MinCLIVersion(),
				MaxCLIVersion: p.MaxCLIVersion(),
				Description:   p.Description(),
				Priority:      p.Priority(),
			}

			if outputJSON {
				data, err := json.MarshalIndent(metadata, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal metadata: %w", err)
				}
				fmt.Println(string(data))
			} else {
				fmt.Printf("Plugin: %s\n", metadata.Name)
				fmt.Printf("Version: %s\n", metadata.Version)
				fmt.Printf("Build Time: %s\n", metadata.BuildTime)
				fmt.Printf("Priority: %d\n", metadata.Priority)
				fmt.Printf("Description: %s\n", metadata.Description)
				fmt.Printf("\nCompatibility:\n")
				fmt.Printf("  Minimum CLI Version: %s\n", metadata.MinCLIVersion)
				fmt.Printf("  Maximum CLI Version: %s\n", metadata.MaxCLIVersion)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	return cmd
}

func newPluginRemoveCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove [plugin-name]",
		Short: "Remove an installed plugin",
		Long:  `Remove a plugin from the local installation directory.`,
		Example: `  plugin-cli plugin remove dummy
  plugin-cli plugin remove converter --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginName := args[0]

			if !force {
				fmt.Printf("Are you sure you want to remove plugin '%s'? (y/N): ", pluginName)
				var response string
				_, _ = fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Canceled")
					return nil
				}
			}

			pm, err := manager.NewPackageManager()
			if err != nil {
				return fmt.Errorf("failed to initialize package manager: %w", err)
			}

			if err := pm.Remove(pluginName); err != nil {
				return fmt.Errorf("failed to remove plugin: %w", err)
			}

			fmt.Printf("Plugin '%s' removed successfully\n", pluginName)
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

func newPluginPathsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "paths",
		Short: "Show plugin discovery paths",
		Long:  `Display all paths where the CLI searches for plugins.`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("Plugin discovery paths:")
			for i, path := range discovery.GetPluginPaths() {
				// Check if path exists
				if _, err := os.Stat(path); err == nil {
					fmt.Printf("%d. %s ✓\n", i+1, path)
				} else {
					fmt.Printf("%d. %s (not found)\n", i+1, path)
				}
			}

			fmt.Println("\nEnvironment variables:")
			if pluginPath := os.Getenv("PLUGIN_PATH"); pluginPath != "" {
				fmt.Printf("  PLUGIN_PATH: %s\n", pluginPath)
			} else {
				fmt.Println("  PLUGIN_PATH: (not set)")
			}
		},
	}
}
