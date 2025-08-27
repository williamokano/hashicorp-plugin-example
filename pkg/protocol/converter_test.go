package protocol

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/williamokano/hashicorp-plugin-example/pkg/types"
)

func TestContextToProto(t *testing.T) {
	tests := []struct {
		name    string
		context *types.Context
		verify  func(t *testing.T, proto *ContextProto)
	}{
		{
			name: "complete context with all fields",
			context: &types.Context{
				Event: types.Event{
					Type:      types.EventMessage,
					Source:    "discord",
					Content:   "test message",
					UserID:    "user123",
					ChannelID: "channel456",
					Metadata: map[string]interface{}{
						"key1": "value1",
						"key2": 42,
						"key3": true,
					},
				},
				Properties: map[string]interface{}{
					"prop1": "value1",
					"prop2": 123.45,
					"prop3": false,
				},
				Responses: []types.Response{
					{
						PluginName: "plugin1",
						Content:    "response1",
						Type:       "text",
						Data: map[string]interface{}{
							"data1": "value1",
						},
					},
					{
						PluginName: "plugin2",
						Content:    "response2",
						Type:       "file",
						Data: map[string]interface{}{
							"path": "/tmp/file",
						},
					},
				},
			},
			verify: func(t *testing.T, proto *ContextProto) {
				assert.NotNil(t, proto.Event)
				assert.Equal(t, "message", proto.Event.Type)
				assert.Equal(t, "discord", proto.Event.Source)
				assert.Equal(t, "test message", proto.Event.Content)
				assert.Equal(t, "user123", proto.Event.UserId)
				assert.Equal(t, "channel456", proto.Event.ChannelId)

				// Verify metadata JSON
				var metadata map[string]interface{}
				err := json.Unmarshal([]byte(proto.Event.MetadataJson), &metadata)
				require.NoError(t, err)
				assert.Equal(t, "value1", metadata["key1"])

				// Verify properties JSON
				var props map[string]interface{}
				err = json.Unmarshal([]byte(proto.PropertiesJson), &props)
				require.NoError(t, err)
				assert.Equal(t, "value1", props["prop1"])

				// Verify responses
				assert.Len(t, proto.Responses, 2)
				assert.Equal(t, "plugin1", proto.Responses[0].PluginName)
				assert.Equal(t, "plugin2", proto.Responses[1].PluginName)
			},
		},
		{
			name: "context with empty fields",
			context: &types.Context{
				Event: types.Event{
					Type:      types.EventCommand,
					Source:    "cli",
					Content:   "",
					UserID:    "",
					ChannelID: "",
					Metadata:  map[string]interface{}{},
				},
				Properties: map[string]interface{}{},
				Responses:  []types.Response{},
			},
			verify: func(t *testing.T, proto *ContextProto) {
				assert.NotNil(t, proto.Event)
				assert.Equal(t, "command", proto.Event.Type)
				assert.Equal(t, "cli", proto.Event.Source)
				assert.Empty(t, proto.Event.Content)
				assert.Empty(t, proto.Event.UserId)
				assert.Empty(t, proto.Event.ChannelId)
				assert.Equal(t, "{}", proto.Event.MetadataJson)
				assert.Equal(t, "{}", proto.PropertiesJson)
				assert.Empty(t, proto.Responses)
			},
		},
		{
			name: "context with nil maps",
			context: &types.Context{
				Event: types.Event{
					Type:     types.EventWebhook,
					Source:   "api",
					Content:  "webhook",
					Metadata: nil,
				},
				Properties: nil,
				Responses:  nil,
			},
			verify: func(t *testing.T, proto *ContextProto) {
				assert.NotNil(t, proto.Event)
				assert.Equal(t, "webhook", proto.Event.Type)
				assert.Equal(t, "null", proto.Event.MetadataJson)
				assert.Equal(t, "null", proto.PropertiesJson)
				assert.Empty(t, proto.Responses)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proto := ContextToProto(tt.context)
			tt.verify(t, proto)
		})
	}
}

func TestProtoToContext(t *testing.T) {
	tests := []struct {
		name   string
		proto  *ContextProto
		verify func(t *testing.T, ctx *types.Context)
	}{
		{
			name: "complete proto with all fields",
			proto: &ContextProto{
				Event: &EventProto{
					Type:         "message",
					Source:       "discord",
					Content:      "test message",
					UserId:       "user123",
					ChannelId:    "channel456",
					MetadataJson: `{"key1":"value1","key2":42}`,
				},
				PropertiesJson: `{"prop1":"value1","prop2":123.45}`,
				Responses: []*ResponseProto{
					{
						PluginName: "plugin1",
						Content:    "response1",
						Type:       "text",
						DataJson:   `{"data1":"value1"}`,
					},
				},
			},
			verify: func(t *testing.T, ctx *types.Context) {
				assert.Equal(t, types.EventMessage, ctx.Event.Type)
				assert.Equal(t, "discord", ctx.Event.Source)
				assert.Equal(t, "test message", ctx.Event.Content)
				assert.Equal(t, "user123", ctx.Event.UserID)
				assert.Equal(t, "channel456", ctx.Event.ChannelID)
				assert.Equal(t, "value1", ctx.Event.Metadata["key1"])
				assert.Equal(t, float64(42), ctx.Event.Metadata["key2"]) // JSON numbers decode as float64
				assert.Equal(t, "value1", ctx.Properties["prop1"])
				assert.Len(t, ctx.Responses, 1)
				assert.Equal(t, "plugin1", ctx.Responses[0].PluginName)
			},
		},
		{
			name: "proto with empty JSON fields",
			proto: &ContextProto{
				Event: &EventProto{
					Type:         "command",
					Source:       "cli",
					MetadataJson: "{}",
				},
				PropertiesJson: "{}",
				Responses:      []*ResponseProto{},
			},
			verify: func(t *testing.T, ctx *types.Context) {
				assert.Equal(t, types.EventCommand, ctx.Event.Type)
				assert.NotNil(t, ctx.Event.Metadata)
				assert.Len(t, ctx.Event.Metadata, 0)
				assert.NotNil(t, ctx.Properties)
				assert.Len(t, ctx.Properties, 0)
				assert.NotNil(t, ctx.Responses)
				assert.Len(t, ctx.Responses, 0)
			},
		},
		{
			name: "proto with invalid JSON",
			proto: &ContextProto{
				Event: &EventProto{
					Type:         "webhook",
					Source:       "api",
					MetadataJson: "invalid json",
				},
				PropertiesJson: "also invalid",
				Responses:      nil,
			},
			verify: func(t *testing.T, ctx *types.Context) {
				// Should handle invalid JSON gracefully
				assert.Equal(t, types.EventWebhook, ctx.Event.Type)
				assert.Nil(t, ctx.Event.Metadata) // Failed to parse
				assert.Nil(t, ctx.Properties)     // Failed to parse
				assert.NotNil(t, ctx.Responses)
				assert.Len(t, ctx.Responses, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ProtoToContext(tt.proto)
			tt.verify(t, ctx)
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Test that converting to proto and back preserves data
	original := &types.Context{
		Event: types.Event{
			Type:      types.EventMessage,
			Source:    "test",
			Content:   "round trip test",
			UserID:    "user",
			ChannelID: "channel",
			Metadata: map[string]interface{}{
				"string": "value",
				"number": float64(42), // Use float64 as JSON decodes to this
				"bool":   true,
				"array":  []interface{}{"a", "b", "c"},
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
		Properties: map[string]interface{}{
			"prop1": "value1",
			"prop2": float64(123),
		},
		Responses: []types.Response{
			{
				PluginName: "test-plugin",
				Content:    "test response",
				Type:       "test",
				Data: map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	// Convert to proto and back
	proto := ContextToProto(original)
	result := ProtoToContext(proto)

	// Verify the result matches original
	assert.Equal(t, original.Event.Type, result.Event.Type)
	assert.Equal(t, original.Event.Source, result.Event.Source)
	assert.Equal(t, original.Event.Content, result.Event.Content)
	assert.Equal(t, original.Event.UserID, result.Event.UserID)
	assert.Equal(t, original.Event.ChannelID, result.Event.ChannelID)

	// Check metadata (note: numbers will be float64 after JSON round-trip)
	assert.Equal(t, original.Event.Metadata["string"], result.Event.Metadata["string"])
	assert.Equal(t, original.Event.Metadata["number"], result.Event.Metadata["number"])
	assert.Equal(t, original.Event.Metadata["bool"], result.Event.Metadata["bool"])

	// Check properties
	assert.Equal(t, original.Properties["prop1"], result.Properties["prop1"])
	assert.Equal(t, original.Properties["prop2"], result.Properties["prop2"])

	// Check responses
	assert.Len(t, result.Responses, 1)
	assert.Equal(t, original.Responses[0].PluginName, result.Responses[0].PluginName)
	assert.Equal(t, original.Responses[0].Content, result.Responses[0].Content)
	assert.Equal(t, original.Responses[0].Type, result.Responses[0].Type)
}
