package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/williamokano/hashicorp-plugin-example/shared"
)

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

type ConverterPlugin struct{}

func (p *ConverterPlugin) ShouldExecute(ctx context.Context, context *shared.Context) shared.ExecutionDecision {
	// Check if conversion is needed
	action, hasAction := context.Properties["action"]
	if !hasAction || action != "convert" {
		return shared.ExecutionDecision{
			ShouldExecute: false,
			Reason:        "No conversion action required",
		}
	}

	// Check media type
	mediaType, hasMedia := context.Properties["media_type"]
	if !hasMedia {
		return shared.ExecutionDecision{
			ShouldExecute: false,
			Reason:        "No media type specified",
		}
	}

	// We can handle video and image
	if mediaType == "video" || mediaType == "image" {
		return shared.ExecutionDecision{
			ShouldExecute: true,
			Reason:        fmt.Sprintf("Ready to convert %s", mediaType),
		}
	}

	return shared.ExecutionDecision{
		ShouldExecute: false,
		Reason:        fmt.Sprintf("Cannot convert media type: %s", mediaType),
	}
}

func (p *ConverterPlugin) Process(ctx context.Context, context *shared.Context) (*shared.Context, error) {
	mediaType := context.Properties["media_type"].(string)

	// Simulate conversion process
	var outputFile string
	var conversionDetails map[string]interface{}

	if mediaType == "video" {
		outputFile = fmt.Sprintf("/tmp/converted_%d.mp4", time.Now().Unix())
		conversionDetails = map[string]interface{}{
			"format":     "mp4",
			"codec":      "h264",
			"resolution": "1920x1080",
			"duration":   "120s",
		}
	} else {
		outputFile = fmt.Sprintf("/tmp/converted_%d.jpg", time.Now().Unix())
		conversionDetails = map[string]interface{}{
			"format":     "jpeg",
			"quality":    "95",
			"resolution": "1920x1080",
		}
	}

	// Add file path to context for next plugins
	context.Properties["file_path"] = outputFile
	context.Properties["conversion_complete"] = true
	context.Properties["conversion_details"] = conversionDetails

	// Add response
	context.Responses = append(context.Responses, shared.Response{
		PluginName: p.Name(),
		Type:       "conversion",
		Content:    fmt.Sprintf("%s converted successfully to %s", mediaType, outputFile),
		Data:       conversionDetails,
	})

	return context, nil
}

func (p *ConverterPlugin) Name() string {
	return "media-converter"
}

func (p *ConverterPlugin) Version() string {
	return Version
}

func (p *ConverterPlugin) BuildTime() string {
	if BuildTime == "unknown" {
		return time.Now().Format(time.RFC3339)
	}
	return BuildTime
}

func (p *ConverterPlugin) MinCLIVersion() string {
	return "1.0.0"
}

func (p *ConverterPlugin) MaxCLIVersion() string {
	return "2.0.0"
}

func (p *ConverterPlugin) Description() string {
	return "Converts media files (video/image) to optimized formats"
}

func (p *ConverterPlugin) Priority() int {
	return 30 // Runs after filter, before uploader
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &shared.GRPCPlugin{Impl: &ConverterPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
	os.Exit(0)
}
