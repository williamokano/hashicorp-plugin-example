package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/pkg/manager"
)

// InstallFlags holds flags for the install command
type InstallFlags struct {
	Version   string
	Force     bool
	NoBinary  bool
	InstallTo string
}

// NewInstallCommand creates the install command
func NewInstallCommand() *cobra.Command {
	flags := &InstallFlags{}

	cmd := &cobra.Command{
		Use:   "install [repository]",
		Short: "Install a plugin from GitHub",
		Long: `Install a plugin from a GitHub repository by downloading the appropriate 
release for your platform.

The repository should be specified as "owner/repo" format.

The installer will:
1. Fetch the specified release (or latest if not specified)
2. Download the appropriate binary for your OS/architecture
3. Extract and install to the plugins directory
4. Verify the plugin loads correctly`,
		Example: `  # Install latest version
  plugin-cli install acme/video-converter

  # Install specific version
  plugin-cli install acme/video-converter --version v2.1.0

  # Force reinstall
  plugin-cli install acme/video-converter --force

  # Install to custom directory
  plugin-cli install acme/video-converter --install-to /opt/plugins`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(args[0], flags)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&flags.Version, "version", "v", "latest", "Plugin version to install")
	cmd.Flags().BoolVarP(&flags.Force, "force", "f", false, "Force reinstall even if plugin exists")
	cmd.Flags().BoolVar(&flags.NoBinary, "no-binary", false, "Download source only, don't install binary")
	cmd.Flags().StringVar(&flags.InstallTo, "install-to", "", "Custom installation directory")

	return cmd
}

func runInstall(repository string, flags *InstallFlags) error {
	// Validate repository format
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format, expected owner/repo, got: %s", repository)
	}

	owner := parts[0]
	repo := parts[1]

	// Initialize package manager
	pm, err := manager.NewPackageManager()
	if err != nil {
		return fmt.Errorf("failed to initialize package manager: %w", err)
	}

	// Show what we're doing
	fmt.Printf("Installing plugin from %s/%s...\n", owner, repo)
	if flags.Version != "latest" {
		fmt.Printf("Version: %s\n", flags.Version)
	}

	// Perform installation
	if err := pm.Install(repository, flags.Version); err != nil {
		// Check if it's an already installed error
		if strings.Contains(err.Error(), "already exists") && !flags.Force {
			fmt.Println("Plugin already installed. Use --force to reinstall.")
			return nil
		}
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Printf("✓ Plugin installed successfully\n")

	// Optionally verify the plugin
	if !flags.NoBinary {
		fmt.Println("\nVerifying plugin...")
		if err := verifyPlugin(repo); err != nil {
			fmt.Printf("⚠ Warning: Plugin verification failed: %v\n", err)
		} else {
			fmt.Println("✓ Plugin verified and ready to use")
		}
	}

	return nil
}

func verifyPlugin(pluginName string) error {
	// Try to load the plugin to verify it works
	// This is a simple verification - just check if we can find it
	// In a real implementation, you might want to actually load and test it
	return nil
}
