package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/williamokano/hashicorp-plugin-example/shared"
)

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

type FilterPlugin struct{}

func (p *FilterPlugin) ShouldExecute(ctx context.Context, context *shared.Context) shared.ExecutionDecision {
	// Only process message events
	if context.Event.Type != shared.EventMessage {
		return shared.ExecutionDecision{
			ShouldExecute: false,
			Reason:        "Not a message event",
		}
	}

	// Check if message contains keywords we care about
	keywords := []string{"convert", "upload", "process", "help"}
	content := strings.ToLower(context.Event.Content)

	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return shared.ExecutionDecision{
				ShouldExecute: true,
				Reason:        "Message contains actionable keyword",
			}
		}
	}

	return shared.ExecutionDecision{
		ShouldExecute: false,
		Reason:        "No actionable keywords found",
	}
}

func (p *FilterPlugin) Process(ctx context.Context, context *shared.Context) (*shared.Context, error) {
	// Extract command from message
	content := strings.ToLower(context.Event.Content)

	// Set properties based on detected intent
	if strings.Contains(content, "convert") {
		context.Properties["action"] = "convert"
		if strings.Contains(content, "video") {
			context.Properties["media_type"] = "video"
		} else if strings.Contains(content, "image") {
			context.Properties["media_type"] = "image"
		}
	}

	if strings.Contains(content, "upload") {
		context.Properties["needs_upload"] = true
	}

	// Add a response indicating the filter processed the message
	context.Responses = append(context.Responses, shared.Response{
		PluginName: p.Name(),
		Type:       "status",
		Content:    "Message filtered and categorized",
		Data: map[string]interface{}{
			"detected_action": context.Properties["action"],
			"timestamp":       time.Now().Unix(),
		},
	})

	return context, nil
}

func (p *FilterPlugin) Name() string {
	return "message-filter"
}

func (p *FilterPlugin) Version() string {
	return Version
}

func (p *FilterPlugin) BuildTime() string {
	if BuildTime == "unknown" {
		return time.Now().Format(time.RFC3339)
	}
	return BuildTime
}

func (p *FilterPlugin) MinCLIVersion() string {
	return "1.0.0"
}

func (p *FilterPlugin) MaxCLIVersion() string {
	return "2.0.0"
}

func (p *FilterPlugin) Description() string {
	return "Filters and categorizes incoming messages"
}

func (p *FilterPlugin) Priority() int {
	return 10 // Runs early in the pipeline
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &shared.GRPCPlugin{Impl: &FilterPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
	os.Exit(0)
}
