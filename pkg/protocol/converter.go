package protocol

import (
	"encoding/json"

	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

// ContextToProto converts a Context to protobuf format
func ContextToProto(ctx *types.Context) *ContextProto {
	propsJSON, _ := json.Marshal(ctx.Properties)
	metadataJSON, _ := json.Marshal(ctx.Event.Metadata)

	responses := make([]*ResponseProto, len(ctx.Responses))
	for i, resp := range ctx.Responses {
		dataJSON, _ := json.Marshal(resp.Data)
		responses[i] = &ResponseProto{
			PluginName: resp.PluginName,
			Content:    resp.Content,
			Type:       resp.Type,
			DataJson:   string(dataJSON),
		}
	}

	return &ContextProto{
		Event: &EventProto{
			Type:         string(ctx.Event.Type),
			Source:       ctx.Event.Source,
			Content:      ctx.Event.Content,
			UserId:       ctx.Event.UserID,
			ChannelId:    ctx.Event.ChannelID,
			MetadataJson: string(metadataJSON),
		},
		PropertiesJson: string(propsJSON),
		Responses:      responses,
	}
}

// ProtoToContext converts protobuf format to Context
func ProtoToContext(proto *ContextProto) *types.Context {
	var props map[string]interface{}
	_ = json.Unmarshal([]byte(proto.PropertiesJson), &props)

	var metadata map[string]interface{}
	_ = json.Unmarshal([]byte(proto.Event.MetadataJson), &metadata)

	responses := make([]types.Response, len(proto.Responses))
	for i, resp := range proto.Responses {
		var data map[string]interface{}
		_ = json.Unmarshal([]byte(resp.DataJson), &data)
		responses[i] = types.Response{
			PluginName: resp.PluginName,
			Content:    resp.Content,
			Type:       resp.Type,
			Data:       data,
		}
	}

	return &types.Context{
		Event: types.Event{
			Type:      types.EventType(proto.Event.Type),
			Source:    proto.Event.Source,
			Content:   proto.Event.Content,
			UserID:    proto.Event.UserId,
			ChannelID: proto.Event.ChannelId,
			Metadata:  metadata,
		},
		Properties: props,
		Responses:  responses,
	}
}
