package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadPluginsConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   bool
		fileContent string
		want        *PluginsConfig
		wantErr     bool
	}{
		{
			name:      "load empty config when file doesn't exist",
			setupFile: false,
			want: &PluginsConfig{
				Plugins: map[string]string{},
			},
			wantErr: false,
		},
		{
			name:      "load valid config",
			setupFile: true,
			fileContent: `{
				"plugins": {
					"plugin-dummy": "1.0.0",
					"plugin-filter": "latest"
				}
			}`,
			want: &PluginsConfig{
				Plugins: map[string]string{
					"plugin-dummy":  "1.0.0",
					"plugin-filter": "latest",
				},
			},
			wantErr: false,
		},
		{
			name:        "handle invalid json",
			setupFile:   true,
			fileContent: `{invalid json}`,
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "handle empty plugins field",
			setupFile:   true,
			fileContent: `{}`,
			want: &PluginsConfig{
				Plugins: map[string]string{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			require.NoError(t, os.Chdir(tempDir))

			// Setup file if needed
			if tt.setupFile {
				err := os.WriteFile(PluginsConfigFile, []byte(tt.fileContent), 0644)
				require.NoError(t, err)
			}

			// Test
			got, err := LoadPluginsConfig()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSavePluginsConfig(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	require.NoError(t, os.Chdir(tempDir))

	config := &PluginsConfig{
		Plugins: map[string]string{
			"plugin-dummy":  "1.0.0",
			"plugin-filter": "^2.0.0",
		},
	}

	// Save config
	err := SavePluginsConfig(config)
	assert.NoError(t, err)

	// Verify file exists and is valid
	data, err := os.ReadFile(PluginsConfigFile)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "plugin-dummy")
	assert.Contains(t, string(data), "1.0.0")

	// Load it back
	loaded, err := LoadPluginsConfig()
	assert.NoError(t, err)
	assert.Equal(t, config, loaded)
}

func TestInitPluginsConfig(t *testing.T) {
	tests := []struct {
		name          string
		existingFile  bool
		wantErr       bool
		errorContains string
	}{
		{
			name:         "create new config",
			existingFile: false,
			wantErr:      false,
		},
		{
			name:          "error if file exists",
			existingFile:  true,
			wantErr:       true,
			errorContains: "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			require.NoError(t, os.Chdir(tempDir))

			// Create existing file if needed
			if tt.existingFile {
				err := os.WriteFile(PluginsConfigFile, []byte("{}"), 0644)
				require.NoError(t, err)
			}

			// Test
			err := InitPluginsConfig()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Verify files were created
				assert.FileExists(t, PluginsConfigFile)
				assert.DirExists(t, ".plugins")

				// Verify config is valid
				config, err := LoadPluginsConfig()
				assert.NoError(t, err)
				assert.NotNil(t, config.Plugins)
				assert.Empty(t, config.Plugins)
			}
		})
	}
}

func TestPluginsConfig_AddPlugin(t *testing.T) {
	config := &PluginsConfig{}

	// Add to nil map
	config.AddPlugin("plugin-dummy", "1.0.0")
	assert.Equal(t, "1.0.0", config.Plugins["plugin-dummy"])

	// Update existing
	config.AddPlugin("plugin-dummy", "2.0.0")
	assert.Equal(t, "2.0.0", config.Plugins["plugin-dummy"])

	// Add another
	config.AddPlugin("plugin-filter", "latest")
	assert.Len(t, config.Plugins, 2)
}

func TestPluginsConfig_RemovePlugin(t *testing.T) {
	config := &PluginsConfig{
		Plugins: map[string]string{
			"plugin-dummy":  "1.0.0",
			"plugin-filter": "2.0.0",
		},
	}

	// Remove existing
	config.RemovePlugin("plugin-dummy")
	assert.Len(t, config.Plugins, 1)
	assert.NotContains(t, config.Plugins, "plugin-dummy")

	// Remove non-existing (should not panic)
	config.RemovePlugin("plugin-nonexistent")
	assert.Len(t, config.Plugins, 1)
}

func TestPluginsConfig_GetPluginVersion(t *testing.T) {
	config := &PluginsConfig{
		Plugins: map[string]string{
			"plugin-dummy": "1.0.0",
		},
	}

	// Get existing
	version, exists := config.GetPluginVersion("plugin-dummy")
	assert.True(t, exists)
	assert.Equal(t, "1.0.0", version)

	// Get non-existing
	version, exists = config.GetPluginVersion("plugin-filter")
	assert.False(t, exists)
	assert.Empty(t, version)
}

func TestIsProjectInitialized(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	require.NoError(t, os.Chdir(tempDir))

	// Not initialized
	assert.False(t, IsProjectInitialized())

	// Create file
	err := os.WriteFile(PluginsConfigFile, []byte("{}"), 0644)
	require.NoError(t, err)

	// Now initialized
	assert.True(t, IsProjectInitialized())
}

func TestGetPluginsDirectory(t *testing.T) {
	dir := GetPluginsDirectory()
	assert.Equal(t, filepath.Join(".", ".plugins"), dir)
}

func TestLoadPluginsLock(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   bool
		fileContent string
		want        *PluginsLock
		wantErr     bool
	}{
		{
			name:      "empty lock when file doesn't exist",
			setupFile: false,
			want: &PluginsLock{
				Plugins: []PluginLockEntry{},
			},
			wantErr: false,
		},
		{
			name:      "load valid lock",
			setupFile: true,
			fileContent: `{
				"plugins": [
					{
						"name": "plugin-dummy",
						"version": "1.0.0",
						"checksum": "abc123",
						"url": "https://github.com/test/test"
					}
				]
			}`,
			want: &PluginsLock{
				Plugins: []PluginLockEntry{
					{
						Name:     "plugin-dummy",
						Version:  "1.0.0",
						Checksum: "abc123",
						URL:      "https://github.com/test/test",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			defer func() { _ = os.Chdir(oldDir) }()
			require.NoError(t, os.Chdir(tempDir))

			// Setup file if needed
			if tt.setupFile {
				err := os.WriteFile(PluginsLockFile, []byte(tt.fileContent), 0644)
				require.NoError(t, err)
			}

			// Test
			got, err := LoadPluginsLock()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSavePluginsLock(t *testing.T) {
	// Setup test directory
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	require.NoError(t, os.Chdir(tempDir))

	lock := &PluginsLock{
		Plugins: []PluginLockEntry{
			{
				Name:     "plugin-dummy",
				Version:  "1.0.0",
				Checksum: "sha256:abc123",
				URL:      "https://github.com/test/releases/plugin-dummy",
			},
		},
	}

	// Save lock
	err := SavePluginsLock(lock)
	assert.NoError(t, err)

	// Verify file exists and is valid
	data, err := os.ReadFile(PluginsLockFile)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "plugin-dummy")
	assert.Contains(t, string(data), "sha256:abc123")

	// Load it back
	loaded, err := LoadPluginsLock()
	assert.NoError(t, err)
	assert.Equal(t, lock, loaded)
}
