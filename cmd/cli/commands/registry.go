package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// NewRegistryCommand creates the registry command
func NewRegistryCommand() *cobra.Command {
	var registryRepo string
	var showAllVersions bool

	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Interact with the plugin registry",
		Long:  `Commands for interacting with the plugin registry on GitHub.`,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available plugins from the registry",
		Long: `List all available plugins from the GitHub releases registry.

This command queries GitHub releases to find available plugins
and their versions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegistryList(cmd, args, registryRepo, showAllVersions)
		},
	}

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for plugins in the registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegistrySearch(cmd, args, registryRepo)
		},
	}

	cmd.PersistentFlags().StringVarP(&registryRepo, "repo", "r", "williamokano/hashicorp-plugin-example", "GitHub repository (owner/repo)")
	listCmd.Flags().BoolVar(&showAllVersions, "all-versions", false, "Show all available versions")

	cmd.AddCommand(listCmd)
	cmd.AddCommand(searchCmd)

	return cmd
}

// GitHub API structures
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	Assets      []GitHubAsset `json:"assets"`
	PublishedAt string        `json:"published_at"`
}

type GitHubAsset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"browser_download_url"`
}

func runRegistryList(_ *cobra.Command, _ []string, registryRepo string, showAllVersions bool) error {
	releases, err := fetchReleases(registryRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}

	plugins := extractPluginInfo(releases)

	if len(plugins) == 0 {
		fmt.Println("No plugins found in registry")
		return nil
	}

	fmt.Printf("Available plugins from %s:\n\n", registryRepo)

	// Group plugins by name
	pluginVersions := make(map[string][]string)
	for _, plugin := range plugins {
		pluginVersions[plugin.Name] = append(pluginVersions[plugin.Name], plugin.Version)
	}

	// Display plugins
	fmt.Printf("%-20s %-15s %s\n", "PLUGIN", "LATEST VERSION", "AVAILABLE VERSIONS")
	fmt.Printf("%-20s %-15s %s\n", "------", "--------------", "------------------")

	for name, versions := range pluginVersions {
		if len(versions) > 0 {
			latest := versions[0] // Assumes versions are sorted (newest first)
			otherVersions := ""

			if showAllVersions && len(versions) > 1 {
				otherVersions = strings.Join(versions[1:], ", ")
			} else if len(versions) > 1 {
				otherVersions = fmt.Sprintf("(+%d more)", len(versions)-1)
			}

			fmt.Printf("%-20s %-15s %s\n", name, latest, otherVersions)
		}
	}

	fmt.Println("\nUse 'plugin-cli download <plugin-name>' to download a plugin")
	return nil
}

func runRegistrySearch(_ *cobra.Command, args []string, registryRepo string) error {
	query := strings.ToLower(args[0])

	releases, err := fetchReleases(registryRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}

	plugins := extractPluginInfo(releases)

	// Filter plugins by query
	matches := make([]PluginInfo, 0, len(plugins))
	for _, plugin := range plugins {
		if strings.Contains(strings.ToLower(plugin.Name), query) {
			matches = append(matches, plugin)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No plugins found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Plugins matching '%s':\n\n", query)
	fmt.Printf("%-20s %-15s\n", "PLUGIN", "VERSION")
	fmt.Printf("%-20s %-15s\n", "------", "-------")

	for _, plugin := range matches {
		fmt.Printf("%-20s %-15s\n", plugin.Name, plugin.Version)
	}

	return nil
}

type PluginInfo struct {
	Name    string
	Version string
}

func fetchReleases(repo string) ([]GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)

	resp, err := http.Get(url) //nolint:gosec // G107: URL is constructed for GitHub API access from validated inputs
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log but don't return error as response body was already read
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}

func extractPluginInfo(releases []GitHubRelease) []PluginInfo {
	plugins := make([]PluginInfo, 0, len(releases)*2) // Estimate capacity
	seen := make(map[string]bool)

	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}

		// Extract version from tag
		version := strings.TrimPrefix(release.TagName, "v")

		// Check if this is a plugin-specific release
		if strings.HasPrefix(release.TagName, "plugin-") {
			// Format: plugin-<name>-v<version>
			parts := strings.Split(release.TagName, "-")
			if len(parts) >= 3 {
				pluginName := "plugin-" + parts[1]
				version = strings.TrimPrefix(parts[2], "v")

				key := fmt.Sprintf("%s-%s", pluginName, version)
				if !seen[key] {
					plugins = append(plugins, PluginInfo{
						Name:    pluginName,
						Version: version,
					})
					seen[key] = true
				}
			}
		} else {
			// General release - extract all plugin assets
			for _, asset := range release.Assets {
				if strings.HasSuffix(asset.Name, ".tar.gz") && strings.HasPrefix(asset.Name, "plugin-") {
					// Extract plugin name from asset filename
					// Format: plugin-<name>_<version>_<os>_<arch>.tar.gz
					parts := strings.Split(asset.Name, "_")
					if len(parts) >= 4 {
						pluginName := parts[0]

						// Skip if it's the CLI
						if pluginName == "plugin-cli" {
							continue
						}

						key := fmt.Sprintf("%s-%s", pluginName, version)
						if !seen[key] {
							plugins = append(plugins, PluginInfo{
								Name:    pluginName,
								Version: version,
							})
							seen[key] = true
						}
					}
				}
			}
		}
	}

	return plugins
}
