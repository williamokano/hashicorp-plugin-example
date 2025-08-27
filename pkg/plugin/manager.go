package plugin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/williamokano/hashicorp-plugin-example/internal/version"
	"github.com/williamokano/hashicorp-plugin-example/pkg/discovery"
	"github.com/williamokano/hashicorp-plugin-example/pkg/protocol"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

type Manager struct {
	logger hclog.Logger
}

func NewManager() *Manager {
	// Get log level from environment, default to Error
	level := hclog.Error
	if envLevel := os.Getenv("PLUGIN_LOG_LEVEL"); envLevel != "" {
		switch envLevel {
		case "trace", "TRACE":
			level = hclog.Trace
		case "debug", "DEBUG":
			level = hclog.Debug
		case "info", "INFO":
			level = hclog.Info
		case "warn", "WARN":
			level = hclog.Warn
		case "error", "ERROR":
			level = hclog.Error
		case "off", "OFF":
			level = hclog.Off
		}
	}

	return &Manager{
		logger: hclog.New(&hclog.LoggerOptions{
			Name:   "plugin-manager",
			Output: os.Stderr,
			Level:  level,
		}),
	}
}

func (m *Manager) LoadPlugin(name string) (*plugin.Client, types.VersionedPlugin, error) {
	discoveredPlugin, err := discovery.FindPlugin(name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to discover plugin: %w", err)
	}

	return m.LoadPluginFromPath(discoveredPlugin.Path)
}

func (m *Manager) LoadPluginFromPath(path string) (*plugin.Client, types.VersionedPlugin, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: protocol.Handshake,
		Plugins:         protocol.PluginMap,
		Cmd:             exec.Command(path),
		Logger:          m.logger,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("failed to create RPC client: %w", err)
	}

	raw, err := rpcClient.Dispense("plugin")
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}

	p, ok := raw.(types.VersionedPlugin)
	if !ok {
		client.Kill()
		return nil, nil, fmt.Errorf("plugin does not implement VersionedPlugin interface")
	}

	minVersion := p.MinCLIVersion()
	maxVersion := p.MaxCLIVersion()
	compatible, err := version.IsCompatible(version.CLIVersion, minVersion, maxVersion)
	if err != nil {
		client.Kill()
		return nil, nil, fmt.Errorf("failed to check version compatibility: %w", err)
	}

	if !compatible {
		client.Kill()
		return nil, nil, fmt.Errorf("plugin version incompatible: CLI version %s, plugin requires %s-%s",
			version.CLIVersion, minVersion, maxVersion)
	}

	return client, p, nil
}

func (m *Manager) ListPlugins() ([]discovery.DiscoveredPlugin, error) {
	return discovery.DiscoverPlugins(discovery.GetPluginPaths())
}

func (m *Manager) GetPluginMetadata(p types.VersionedPlugin) types.PluginMetadata {
	return types.PluginMetadata{
		Name:          p.Name(),
		Version:       p.Version(),
		BuildTime:     p.BuildTime(),
		MinCLIVersion: p.MinCLIVersion(),
		MaxCLIVersion: p.MaxCLIVersion(),
		Description:   p.Description(),
		Priority:      p.Priority(),
	}
}
