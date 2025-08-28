package commands

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download [plugin-name]",
	Short: "Download a plugin from GitHub releases",
	Long: `Download a plugin binary from GitHub releases.

The download command fetches plugin binaries from GitHub releases
using the Terraform-style naming convention:
  <plugin-name>_<version>_<os>_<arch>.tar.gz

Examples:
  # Download latest version of a plugin
  plugin-cli download dummy

  # Download specific version
  plugin-cli download dummy --version 1.0.0

  # Download from custom repository
  plugin-cli download dummy --repo owner/repo

  # Download and verify checksum
  plugin-cli download dummy --verify`,
	Args: cobra.ExactArgs(1),
	RunE: runDownload,
}

var (
	downloadVersion string
	downloadRepo    string
	verifyChecksum  bool
	downloadPath    string
	forceDownload   bool
)

func init() {
	downloadCmd.Flags().StringVar(&downloadVersion, "version", "latest", "Plugin version to download")
	downloadCmd.Flags().StringVarP(&downloadRepo, "repo", "r", "williamokano/hashicorp-plugin-example", "GitHub repository (owner/repo)")
	downloadCmd.Flags().BoolVar(&verifyChecksum, "verify", true, "Verify SHA256 checksum")
	downloadCmd.Flags().StringVarP(&downloadPath, "path", "p", ".plugins", "Directory to download plugin to")
	downloadCmd.Flags().BoolVarP(&forceDownload, "force", "f", false, "Force download even if plugin exists")

	rootCmd.AddCommand(downloadCmd)
}

func runDownload(cmd *cobra.Command, args []string) error {
	pluginName := args[0]
	if !strings.HasPrefix(pluginName, "plugin-") {
		pluginName = "plugin-" + pluginName
	}

	// Get OS and architecture
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Ensure download path exists
	if err := os.MkdirAll(downloadPath, 0750); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	// Check if plugin already exists
	pluginPath := filepath.Join(downloadPath, pluginName)
	if runtime.GOOS == "windows" {
		pluginPath += ".exe"
	}

	if !forceDownload {
		if _, err := os.Stat(pluginPath); err == nil {
			fmt.Printf("Plugin %s already exists at %s\n", pluginName, pluginPath)
			fmt.Println("Use --force to re-download")
			return nil
		}
	}

	// Get release information
	version := downloadVersion
	if version == "latest" {
		var err error
		version, err = getLatestVersion(downloadRepo, pluginName)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
	}

	// Construct download URL
	archiveName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", pluginName, version, osName, archName)
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s",
		downloadRepo, version, archiveName)

	// For plugin-specific releases, use different tag format
	if pluginName != "plugin-cli" {
		shortName := strings.TrimPrefix(pluginName, "plugin-")
		downloadURL = fmt.Sprintf("https://github.com/%s/releases/download/plugin-%s-v%s/%s",
			downloadRepo, shortName, version, archiveName)
	}

	fmt.Printf("Downloading %s from %s...\n", pluginName, downloadURL)

	// Download the archive
	archivePath := filepath.Join(downloadPath, archiveName)
	if err := downloadFile(archivePath, downloadURL); err != nil {
		// Try alternative URL format (general release)
		downloadURL = fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s",
			downloadRepo, version, archiveName)
		if err := downloadFile(archivePath, downloadURL); err != nil {
			return fmt.Errorf("failed to download plugin: %w", err)
		}
	}
	defer func() {
		if err := os.Remove(archivePath); err != nil {
			// Log the error but don't fail the command
			fmt.Printf("Warning: Could not remove temporary file %s: %v\n", archivePath, err)
		}
	}()

	// Verify checksum if requested
	if verifyChecksum {
		checksumURL := downloadURL + ".sha256"
		if err := verifyFileChecksum(archivePath, checksumURL); err != nil {
			fmt.Printf("Warning: Could not verify checksum: %v\n", err)
			// Continue anyway, checksum might not be available
		} else {
			fmt.Println("Checksum verified successfully")
		}
	}

	// Extract the plugin
	fmt.Printf("Extracting %s...\n", archiveName)
	if err := extractTarGz(archivePath, downloadPath); err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	// Make the plugin executable
	if err := os.Chmod(pluginPath, 0755); err != nil { //nolint:gosec // G302: executable files need 0755
		return fmt.Errorf("failed to make plugin executable: %w", err)
	}

	fmt.Printf("Successfully downloaded %s v%s to %s\n", pluginName, version, pluginPath)
	return nil
}

func getLatestVersion(repo, pluginName string) (string, error) {
	// For now, return a default version
	// In a real implementation, this would query the GitHub API
	// to get the latest release version
	return "1.0.0", nil
}

func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifyFileChecksum(filepath, checksumURL string) error {
	// Download checksum file
	resp, err := http.Get(checksumURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("checksum file not found")
	}

	checksumData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse expected checksum (format: "sha256sum filename")
	parts := strings.Fields(string(checksumData))
	if len(parts) < 1 {
		return fmt.Errorf("invalid checksum format")
	}
	expectedChecksum := parts[0]

	// Calculate actual checksum
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s",
			expectedChecksum, actualChecksum)
	}

	return nil
}

func extractTarGz(archivePath, destPath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Validate file path to prevent zip slip attacks
		if strings.Contains(header.Name, "..") || filepath.IsAbs(header.Name) {
			continue // Skip potentially malicious entries
		}
		target := filepath.Join(destPath, header.Name)
		// Ensure target is within destination path
		if !strings.HasPrefix(target, filepath.Clean(destPath)+string(os.PathSeparator)) {
			continue // Skip files outside destination
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			// Limit file size to prevent decompression bombs (100MB limit)
			const maxFileSize = 100 * 1024 * 1024
			if _, err := io.CopyN(outFile, tarReader, maxFileSize); err != nil && err != io.EOF {
				_ = outFile.Close() // Best effort cleanup
				return err
			}
			_ = outFile.Close() // Best effort cleanup

			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	return nil
}
