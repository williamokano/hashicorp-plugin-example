package commands

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/williamokano/hashicorp-plugin-example/internal/version"
)

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CLI version information",
		Long:  `Display the CLI version, build time, and other version-related information.`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("CLI Version: %s\n", version.CLIVersion)
			fmt.Printf("Build Time: %s\n", version.CLIBuildTime)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
