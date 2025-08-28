package manager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type PluginRegistry struct {
	Plugins []PluginPackage `json:"plugins"`
}

type PluginPackage struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Repository  string `json:"repository"`
	Author      string `json:"author"`
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type PackageManager struct {
	installDir string
	httpClient *http.Client
}

func NewPackageManager() (*PackageManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	installDir := filepath.Join(homeDir, ".local", "share", "plugins")
	if err := os.MkdirAll(installDir, 0750); err != nil {
		return nil, err
	}

	return &PackageManager{
		installDir: installDir,
		httpClient: &http.Client{},
	}, nil
}

func (pm *PackageManager) Install(repository, version string) error {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository format, expected owner/repo")
	}

	owner := parts[0]
	repo := parts[1]

	release, err := pm.getRelease(owner, repo, version)
	if err != nil {
		return fmt.Errorf("failed to get release: %w", err)
	}

	asset := pm.findAsset(release)
	if asset == nil {
		return fmt.Errorf("no compatible asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	downloadPath := filepath.Join(pm.installDir, asset.Name)
	if err := pm.downloadFile(asset.BrowserDownloadURL, downloadPath); err != nil {
		return fmt.Errorf("failed to download plugin: %w", err)
	}

	if err := pm.extractPlugin(downloadPath, repo); err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	_ = os.Remove(downloadPath) // Best effort cleanup

	return nil
}

func (pm *PackageManager) getRelease(owner, repo, version string) (*GitHubRelease, error) {
	var url string
	if version == "latest" || version == "" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, version)
	}

	resp, err := pm.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func (pm *PackageManager) findAsset(release *GitHubRelease) *struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
} {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) && strings.Contains(name, arch) {
			return &asset
		}
	}

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, osName) {
			return &asset
		}
	}

	return nil
}

func (pm *PackageManager) downloadFile(url, dest string) error {
	resp, err := pm.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (pm *PackageManager) extractPlugin(archivePath, pluginName string) error {
	ext := filepath.Ext(archivePath)

	switch ext {
	case ".gz", ".tgz":
		return pm.extractTarGz(archivePath, pluginName)
	case ".zip":
		return pm.extractZip(archivePath, pluginName)
	default:
		binaryPath := filepath.Join(pm.installDir, "plugin-"+pluginName)
		if runtime.GOOS == "windows" {
			binaryPath += ".exe"
		}

		if err := os.Rename(archivePath, binaryPath); err != nil {
			return err
		}
		return os.Chmod(binaryPath, 0755) //nolint:gosec // G302: executable files need 0755
	}
}

func (pm *PackageManager) extractTarGz(archivePath, pluginName string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		if strings.Contains(header.Name, "plugin-") || header.Name == pluginName {
			targetPath := filepath.Join(pm.installDir, "plugin-"+pluginName)
			if runtime.GOOS == "windows" {
				targetPath += ".exe"
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// Limit file size to prevent decompression bombs (100MB limit)
			const maxFileSize = 100 * 1024 * 1024
			if _, err := io.CopyN(outFile, tr, maxFileSize); err != nil && err != io.EOF {
				return err
			}

			return os.Chmod(targetPath, 0755) //nolint:gosec // G302: executable files need 0755
		}
	}

	return fmt.Errorf("plugin binary not found in archive")
}

func (pm *PackageManager) extractZip(archivePath, pluginName string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "plugin-") || f.Name == pluginName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			targetPath := filepath.Join(pm.installDir, "plugin-"+pluginName)
			if runtime.GOOS == "windows" {
				targetPath += ".exe"
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// Limit file size to prevent decompression bombs (100MB limit)
			const maxFileSize = 100 * 1024 * 1024
			if _, err := io.CopyN(outFile, rc, maxFileSize); err != nil && err != io.EOF {
				return err
			}

			return os.Chmod(targetPath, 0755) //nolint:gosec // G302: executable files need 0755
		}
	}

	return fmt.Errorf("plugin binary not found in archive")
}

func (pm *PackageManager) List() ([]string, error) {
	entries, err := os.ReadDir(pm.installDir)
	if err != nil {
		return nil, err
	}

	var plugins []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "plugin-") {
			name := strings.TrimPrefix(entry.Name(), "plugin-")
			if runtime.GOOS == "windows" {
				name = strings.TrimSuffix(name, ".exe")
			}
			plugins = append(plugins, name)
		}
	}

	return plugins, nil
}

func (pm *PackageManager) Remove(pluginName string) error {
	pluginPath := filepath.Join(pm.installDir, "plugin-"+pluginName)
	if runtime.GOOS == "windows" {
		pluginPath += ".exe"
	}

	return os.Remove(pluginPath)
}
