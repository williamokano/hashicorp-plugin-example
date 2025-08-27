package types

import "context"

// Plugin interface for event-driven architecture
type Plugin interface {
	// Decide whether this plugin should process the event
	ShouldExecute(ctx context.Context, context *Context) ExecutionDecision

	// Process the event and potentially modify the context
	Process(ctx context.Context, context *Context) (*Context, error)

	// Plugin metadata
	Name() string
	Description() string
	Priority() int // Lower numbers run first
}

// VersionedPlugin adds version information
type VersionedPlugin interface {
	Plugin
	Version() string
	BuildTime() string
	MinCLIVersion() string
	MaxCLIVersion() string
}

// PluginMetadata for serialization
type PluginMetadata struct {
	Name          string `json:"name"`
	Version       string `json:"version"`
	BuildTime     string `json:"build_time"`
	MinCLIVersion string `json:"min_cli_version"`
	MaxCLIVersion string `json:"max_cli_version"`
	Description   string `json:"description"`
	Priority      int    `json:"priority"`
}
