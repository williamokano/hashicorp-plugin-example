package protocol

import (
	"context"

	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

// GRPCClient implements the gRPC client
type GRPCClient struct {
	client PluginClient
}

// ShouldExecute checks if the plugin should execute
func (m *GRPCClient) ShouldExecute(ctx context.Context, context *types.Context) types.ExecutionDecision {
	resp, err := m.client.ShouldExecute(ctx, ContextToProto(context))
	if err != nil {
		return types.ExecutionDecision{ShouldExecute: false, Reason: err.Error()}
	}
	return types.ExecutionDecision{
		ShouldExecute: resp.ShouldExecute,
		Reason:        resp.Reason,
	}
}

// Process executes the plugin processing
func (m *GRPCClient) Process(ctx context.Context, context *types.Context) (*types.Context, error) {
	resp, err := m.client.Process(ctx, ContextToProto(context))
	if err != nil {
		return nil, err
	}
	return ProtoToContext(resp), nil
}

// Name returns the plugin name
func (m *GRPCClient) Name() string {
	metadata, err := m.client.GetMetadata(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return metadata.Name
}

// Version returns the plugin version
func (m *GRPCClient) Version() string {
	metadata, err := m.client.GetMetadata(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return metadata.Version
}

// BuildTime returns the build time
func (m *GRPCClient) BuildTime() string {
	metadata, err := m.client.GetMetadata(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return metadata.BuildTime
}

// MinCLIVersion returns the minimum CLI version
func (m *GRPCClient) MinCLIVersion() string {
	metadata, err := m.client.GetMetadata(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return metadata.MinCliVersion
}

// MaxCLIVersion returns the maximum CLI version
func (m *GRPCClient) MaxCLIVersion() string {
	metadata, err := m.client.GetMetadata(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return metadata.MaxCliVersion
}

// Description returns the plugin description
func (m *GRPCClient) Description() string {
	metadata, err := m.client.GetMetadata(context.Background(), &Empty{})
	if err != nil {
		return ""
	}
	return metadata.Description
}

// Priority returns the plugin priority
func (m *GRPCClient) Priority() int {
	metadata, err := m.client.GetMetadata(context.Background(), &Empty{})
	if err != nil {
		return 100
	}
	return int(metadata.Priority)
}
