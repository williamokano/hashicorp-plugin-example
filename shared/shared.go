// Package shared provides backward compatibility by re-exporting types from their new locations.
// New code should import directly from pkg/types and pkg/protocol.
package shared

import (
	"github.com/williamokano/hashicorp-plugin-example/pkg/protocol"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

// Re-export types for backward compatibility
type (
	Event             = types.Event
	EventType         = types.EventType
	Context           = types.Context
	Response          = types.Response
	ExecutionDecision = types.ExecutionDecision
	Plugin            = types.Plugin
	VersionedPlugin   = types.VersionedPlugin
	PluginMetadata    = types.PluginMetadata
)

// Re-export event type constants
const (
	EventMessage   = types.EventMessage
	EventCommand   = types.EventCommand
	EventWebhook   = types.EventWebhook
	EventScheduled = types.EventScheduled
)

// Re-export protocol elements
var (
	Handshake = protocol.Handshake
	PluginMap = protocol.PluginMap
)

type GRPCPlugin = protocol.GRPCPlugin
