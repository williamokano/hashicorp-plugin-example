package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Interact with the plugin registry",
	Long:  `Commands for interacting with the plugin registry on GitHub.`,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugins from the registry",
	Long: `List all available plugins from the GitHub releases registry.

This command queries GitHub releases to find available plugins
and their versions.`,
	RunE: runRegistryList,
}

var registrySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for plugins in the registry",
	Args:  cobra.ExactArgs(1),
	RunE:  runRegistrySearch,
}

var (
	registryRepo    string
	showAllVersions bool
)

func init() {
	registryCmd.PersistentFlags().StringVarP(&registryRepo, "repo", "r", "williamokano/hashicorp-plugin-example", "GitHub repository (owner/repo)")

	registryListCmd.Flags().BoolVar(&showAllVersions, "all-versions", false, "Show all available versions")

	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registrySearchCmd)

	rootCmd.AddCommand(registryCmd)
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

func runRegistryList(cmd *cobra.Command, args []string) error {
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

func runRegistrySearch(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])

	releases, err := fetchReleases(registryRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch releases: %w", err)
	}

	plugins := extractPluginInfo(releases)

	// Filter plugins by query
	var matches []PluginInfo
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

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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
	var plugins []PluginInfo
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
