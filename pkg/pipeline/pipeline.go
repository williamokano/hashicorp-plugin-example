package pipeline

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/williamokano/hashicorp-plugin-example/pkg/discovery"
	pluginpkg "github.com/williamokano/hashicorp-plugin-example/pkg/plugin"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

type Pipeline struct {
	manager *pluginpkg.Manager
	logger  hclog.Logger
}

type LoadedPlugin struct {
	Client *plugin.Client
	Plugin types.VersionedPlugin
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		manager: pluginpkg.NewManager(),
		logger: hclog.New(&hclog.LoggerOptions{
			Name:  "pipeline",
			Level: hclog.Info,
		}),
	}
}

// ProcessEvent runs all plugins in priority order
func (p *Pipeline) ProcessEvent(ctx context.Context, event types.Event) (*types.Context, error) {
	// Initialize context
	context := &types.Context{
		Event:      event,
		Properties: make(map[string]interface{}),
		Responses:  []types.Response{},
	}

	// Discover and load all plugins
	plugins, err := p.loadAllPlugins()
	if err != nil {
		return context, fmt.Errorf("failed to load plugins: %w", err)
	}
	defer p.cleanupPlugins(plugins)

	// Sort plugins by priority
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Plugin.Priority() < plugins[j].Plugin.Priority()
	})

	// Execute plugins in order
	for _, loadedPlugin := range plugins {
		pluginName := loadedPlugin.Plugin.Name()
		p.logger.Info("checking plugin", "name", pluginName, "priority", loadedPlugin.Plugin.Priority())

		// Check if plugin should execute
		decision := loadedPlugin.Plugin.ShouldExecute(ctx, context)
		if !decision.ShouldExecute {
			p.logger.Info("plugin skipped", "name", pluginName, "reason", decision.Reason)
			continue
		}

		p.logger.Info("executing plugin", "name", pluginName)

		// Execute plugin
		newContext, err := loadedPlugin.Plugin.Process(ctx, context)
		if err != nil {
			p.logger.Error("plugin execution failed", "name", pluginName, "error", err)
			// Continue with other plugins even if one fails
			continue
		}

		// Update context for next plugin
		context = newContext
		p.logger.Info("plugin executed successfully", "name", pluginName)
	}

	return context, nil
}

// ProcessMessage is a convenience method for processing text messages
func (p *Pipeline) ProcessMessage(ctx context.Context, source, content, userID, channelID string) (*types.Context, error) {
	event := types.Event{
		Type:      types.EventMessage,
		Source:    source,
		Content:   content,
		UserID:    userID,
		ChannelID: channelID,
		Metadata:  make(map[string]interface{}),
	}
	return p.ProcessEvent(ctx, event)
}

// ProcessCommand is a convenience method for processing commands
func (p *Pipeline) ProcessCommand(ctx context.Context, source, command, userID, channelID string) (*types.Context, error) {
	event := types.Event{
		Type:      types.EventCommand,
		Source:    source,
		Content:   command,
		UserID:    userID,
		ChannelID: channelID,
		Metadata:  make(map[string]interface{}),
	}
	return p.ProcessEvent(ctx, event)
}

func (p *Pipeline) loadAllPlugins() ([]LoadedPlugin, error) {
	discovered, err := discovery.DiscoverPlugins(discovery.GetPluginPaths())
	if err != nil {
		return nil, err
	}

	var plugins []LoadedPlugin
	for _, disc := range discovered {
		p.logger.Debug("loading plugin", "name", disc.Name, "path", disc.Path)

		client, plugin, err := p.manager.LoadPluginFromPath(disc.Path)
		if err != nil {
			p.logger.Error("failed to load plugin", "name", disc.Name, "error", err)
			continue
		}

		plugins = append(plugins, LoadedPlugin{
			Client: client,
			Plugin: plugin,
		})
	}

	return plugins, nil
}

func (p *Pipeline) cleanupPlugins(plugins []LoadedPlugin) {
	for _, plugin := range plugins {
		plugin.Client.Kill()
	}
}
