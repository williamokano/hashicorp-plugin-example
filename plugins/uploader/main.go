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

type UploaderPlugin struct{}

func (p *UploaderPlugin) ShouldExecute(ctx context.Context, context *shared.Context) shared.ExecutionDecision {
	// Check if upload is needed based on context properties
	needsUpload, ok := context.Properties["needs_upload"].(bool)
	if !ok || !needsUpload {
		return shared.ExecutionDecision{
			ShouldExecute: false,
			Reason:        "Upload not required",
		}
	}

	// Check if there's a file to upload (set by previous plugin)
	if _, hasFile := context.Properties["file_path"]; !hasFile {
		return shared.ExecutionDecision{
			ShouldExecute: false,
			Reason:        "No file to upload",
		}
	}

	return shared.ExecutionDecision{
		ShouldExecute: true,
		Reason:        "File ready for upload",
	}
}

func (p *UploaderPlugin) Process(ctx context.Context, context *shared.Context) (*shared.Context, error) {
	filePath, _ := context.Properties["file_path"].(string)

	// Simulate uploading to S3
	uploadedURL := fmt.Sprintf("https://s3.example.com/uploads/%d/%s",
		time.Now().Unix(),
		filePath)

	// Add upload URL to context for other plugins to use
	context.Properties["uploaded_url"] = uploadedURL
	context.Properties["upload_timestamp"] = time.Now().Unix()

	// Add response
	context.Responses = append(context.Responses, shared.Response{
		PluginName: p.Name(),
		Type:       "upload",
		Content:    fmt.Sprintf("File uploaded successfully to %s", uploadedURL),
		Data: map[string]interface{}{
			"url":        uploadedURL,
			"original":   filePath,
			"size_bytes": 1024 * 50,   // Simulated
			"mime_type":  "video/mp4", // Simulated
		},
	})

	return context, nil
}

func (p *UploaderPlugin) Name() string {
	return "s3-uploader"
}

func (p *UploaderPlugin) Version() string {
	return Version
}

func (p *UploaderPlugin) BuildTime() string {
	if BuildTime == "unknown" {
		return time.Now().Format(time.RFC3339)
	}
	return BuildTime
}

func (p *UploaderPlugin) MinCLIVersion() string {
	return "1.0.0"
}

func (p *UploaderPlugin) MaxCLIVersion() string {
	return "2.0.0"
}

func (p *UploaderPlugin) Description() string {
	return "Uploads files to S3 when needed"
}

func (p *UploaderPlugin) Priority() int {
	return 50 // Runs after processing plugins
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &shared.GRPCPlugin{Impl: &UploaderPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
	os.Exit(0)
}
