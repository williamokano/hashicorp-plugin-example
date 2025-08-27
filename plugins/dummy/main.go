package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/williamokano/hashicorp-plugin-example/pkg/protocol"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
	"github.com/williamokano/hashicorp-plugin-example/shared"
)

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

type DummyPlugin struct{}

func (p *DummyPlugin) ShouldExecute(ctx context.Context, context *types.Context) types.ExecutionDecision {
	// Always execute for demonstration
	return types.ExecutionDecision{
		ShouldExecute: true,
		Reason:        "Dummy plugin always executes",
	}
}

func (p *DummyPlugin) Process(ctx context.Context, pipelineCtx *types.Context) (*types.Context, error) {
	// Add some dummy processing
	pipelineCtx.Properties["dummy_processed"] = fmt.Sprintf("Event type %s processed at %s", pipelineCtx.Event.Type, time.Now().Format(time.RFC3339))
	pipelineCtx.Properties["dummy_message"] = "Hello from dummy plugin!"

	// Add a response
	pipelineCtx.Responses = append(pipelineCtx.Responses, types.Response{
		PluginName: p.Name(),
		Content:    "Dummy plugin processed the event successfully",
		Type:       "text",
	})

	return pipelineCtx, nil
}

func (p *DummyPlugin) Priority() int {
	return 100
}

func (p *DummyPlugin) Name() string {
	return "dummy-plugin"
}

func (p *DummyPlugin) Version() string {
	return Version
}

func (p *DummyPlugin) BuildTime() string {
	if BuildTime == "unknown" {
		return time.Now().Format(time.RFC3339)
	}
	return BuildTime
}

func (p *DummyPlugin) MinCLIVersion() string {
	return "1.0.0"
}

func (p *DummyPlugin) MaxCLIVersion() string {
	return "2.0.0"
}

func (p *DummyPlugin) Description() string {
	return "A dummy plugin for demonstration purposes"
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &protocol.GRPCPlugin{Impl: &DummyPlugin{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
	os.Exit(0)
}
