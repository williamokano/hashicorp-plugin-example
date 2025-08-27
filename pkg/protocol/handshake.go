package protocol

import "github.com/hashicorp/go-plugin"

// Handshake is the shared handshake config for plugins
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello-plugin",
}

// PluginMap is the map of plugins we support
var PluginMap = map[string]plugin.Plugin{
	"plugin": &GRPCPlugin{},
}
