package protocol

import (
	"context"

	"github.com/hashicorp/go-plugin"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
	"google.golang.org/grpc"
)

// GRPCPlugin is the gRPC plugin implementation
type GRPCPlugin struct {
	plugin.Plugin
	Impl types.VersionedPlugin
}

// GRPCServer registers the gRPC server
func (p *GRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// GRPCClient returns the gRPC client
func (p *GRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: NewPluginClient(c)}, nil
}
