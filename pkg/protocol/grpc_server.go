package protocol

import (
	"context"

	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

// GRPCServer implements the gRPC server
type GRPCServer struct {
	Impl types.VersionedPlugin
	UnimplementedPluginServer
}

// ShouldExecute decides if the plugin should run
func (m *GRPCServer) ShouldExecute(ctx context.Context, req *ContextProto) (*ExecutionDecisionProto, error) {
	context := ProtoToContext(req)
	decision := m.Impl.ShouldExecute(ctx, context)
	return &ExecutionDecisionProto{
		ShouldExecute: decision.ShouldExecute,
		Reason:        decision.Reason,
	}, nil
}

// Process handles the event processing
func (m *GRPCServer) Process(ctx context.Context, req *ContextProto) (*ContextProto, error) {
	inputContext := ProtoToContext(req)
	outputContext, err := m.Impl.Process(ctx, inputContext)
	if err != nil {
		return nil, err
	}
	return ContextToProto(outputContext), nil
}

// GetMetadata returns plugin metadata
func (m *GRPCServer) GetMetadata(ctx context.Context, req *Empty) (*Metadata, error) {
	return &Metadata{
		Name:          m.Impl.Name(),
		Version:       m.Impl.Version(),
		BuildTime:     m.Impl.BuildTime(),
		MinCliVersion: m.Impl.MinCLIVersion(),
		MaxCliVersion: m.Impl.MaxCLIVersion(),
		Description:   m.Impl.Description(),
		Priority:      int32(m.Impl.Priority()),
	}, nil
}
