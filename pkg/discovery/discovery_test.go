package discovery

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverPlugins(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		want        int
		wantPlugins []string
		wantErr     bool
	}{
		{
			name: "discovers plugins with correct prefix",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				createExecutableFile(t, filepath.Join(dir, "plugin-test"))
				createExecutableFile(t, filepath.Join(dir, "plugin-another"))
				createExecutableFile(t, filepath.Join(dir, "not-a-plugin"))
				return dir
			},
			want:        2,
			wantPlugins: []string{"test", "another"},
			wantErr:     false,
		},
		{
			name: "ignores directories",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				createExecutableFile(t, filepath.Join(dir, "plugin-valid"))
				err := os.Mkdir(filepath.Join(dir, "plugin-directory"), 0o755)
				require.NoError(t, err)
				return dir
			},
			want:        1,
			wantPlugins: []string{"valid"},
			wantErr:     false,
		},
		{
			name: "handles non-existent directory",
			setup: func(t *testing.T) string {
				return "/non/existent/path"
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "ignores non-executable files",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				createNonExecutableFile(t, filepath.Join(dir, "plugin-noexec"))
				createExecutableFile(t, filepath.Join(dir, "plugin-exec"))
				return dir
			},
			want:        1,
			wantPlugins: []string{"exec"},
			wantErr:     false,
		},
		{
			name: "handles windows exe extension",
			setup: func(t *testing.T) string {
				if runtime.GOOS != "windows" {
					t.Skip("Windows-specific test")
				}
				dir := t.TempDir()
				createExecutableFile(t, filepath.Join(dir, "plugin-test.exe"))
				return dir
			},
			want:        1,
			wantPlugins: []string{"test"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)

			got, err := DiscoverPlugins([]string{dir})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got, tt.want)

			// Check that all expected plugins are found (order doesn't matter)
			foundPlugins := make(map[string]bool)
			for _, plugin := range got {
				foundPlugins[plugin.Name] = true
				assert.Contains(t, plugin.Path, dir)
			}

			for _, wantPlugin := range tt.wantPlugins {
				assert.True(t, foundPlugins[wantPlugin], "Expected plugin %s not found", wantPlugin)
			}
		})
	}
}

func TestFindPlugin(t *testing.T) {
	// Setup test directory with plugins
	dir := t.TempDir()
	createExecutableFile(t, filepath.Join(dir, "plugin-target"))
	createExecutableFile(t, filepath.Join(dir, "plugin-other"))

	// Save original environment and set test path
	oldPluginPath := os.Getenv("PLUGIN_PATH")
	_ = os.Setenv("PLUGIN_PATH", dir)
	defer func() {
		_ = os.Setenv("PLUGIN_PATH", oldPluginPath)
	}()

	tests := []struct {
		name       string
		pluginName string
		wantFound  bool
		wantErr    bool
	}{
		{
			name:       "finds existing plugin",
			pluginName: "target",
			wantFound:  true,
			wantErr:    false,
		},
		{
			name:       "returns error for non-existent plugin",
			pluginName: "nonexistent",
			wantFound:  false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := FindPlugin(tt.pluginName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, plugin)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, plugin)
			assert.Equal(t, tt.pluginName, plugin.Name)
		})
	}
}

func TestGetPluginPaths(t *testing.T) {
	// Save original environment
	originalHome := os.Getenv("HOME")
	originalPluginPath := os.Getenv("PLUGIN_PATH")
	defer func() {
		_ = os.Setenv("HOME", originalHome)
		_ = os.Setenv("PLUGIN_PATH", originalPluginPath)
	}()

	tests := []struct {
		name      string
		setup     func()
		wantPaths []string
	}{
		{
			name: "includes home directory path",
			setup: func() {
				if runtime.GOOS == "windows" {
					_ = os.Setenv("USERPROFILE", "C:\\test\\home")
				} else {
					_ = os.Setenv("HOME", "/test/home")
				}
				_ = os.Unsetenv("PLUGIN_PATH")
			},
			wantPaths: []string{
				getExpectedHomePath(),
			},
		},
		{
			name: "includes PLUGIN_PATH environment variable",
			setup: func() {
				if runtime.GOOS == "windows" {
					_ = os.Setenv("PLUGIN_PATH", "C:\\custom\\path;C:\\another\\path")
				} else {
					_ = os.Setenv("PLUGIN_PATH", "/custom/path:/another/path")
				}
			},
			wantPaths: []string{
				getExpectedCustomPath1(),
				getExpectedCustomPath2(),
			},
		},
		{
			name: "includes system path",
			setup: func() {
				_ = os.Unsetenv("PLUGIN_PATH")
				if runtime.GOOS == "windows" {
					_ = os.Setenv("ProgramData", "C:\\ProgramData")
				}
			},
			wantPaths: []string{
				getExpectedSystemPath(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			paths := GetPluginPaths()

			for _, wantPath := range tt.wantPaths {
				assert.Contains(t, paths, wantPath)
			}
		})
	}
}

// Helper functions for cross-platform paths
func getExpectedHomePath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("C:", "test", "home", ".local", "share", "plugins")
	}
	return "/test/home/.local/share/plugins"
}

func getExpectedCustomPath1() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("C:", "custom", "path")
	}
	return "/custom/path"
}

func getExpectedCustomPath2() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("C:", "another", "path")
	}
	return "/another/path"
}

func getExpectedSystemPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("C:", "ProgramData", "plugins")
	}
	return "/usr/local/lib/plugins"
}

// Helper functions
func createExecutableFile(t *testing.T, path string) {
	t.Helper()
	
	// On Windows, ensure the file has .exe extension
	if runtime.GOOS == "windows" && !strings.HasSuffix(path, ".exe") {
		path = path + ".exe"
	}
	
	file, err := os.Create(path)
	require.NoError(t, err)
	defer func() {
		if err := file.Close(); err != nil {
			// Log but don't fail test as file operations are complete
		}
	}()

	// On Unix-like systems, set executable permissions
	if runtime.GOOS != "windows" {
		err = os.Chmod(path, 0o755)
		require.NoError(t, err)
	}
	// Note: On Windows, files are executable by default based on extension
}

func createNonExecutableFile(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	defer func() {
		if err := file.Close(); err != nil {
			// Log but don't fail test as file operations are complete
		}
	}()

	err = os.Chmod(path, 0o644)
	require.NoError(t, err)
}
