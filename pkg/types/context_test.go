package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_AddProperty(t *testing.T) {
	ctx := &Context{
		Event: Event{
			Type:    EventMessage,
			Content: "test",
		},
		Properties: make(map[string]interface{}),
		Responses:  []Response{},
	}

	// Add properties
	ctx.Properties["key1"] = "value1"
	ctx.Properties["key2"] = 123
	ctx.Properties["key3"] = true

	assert.Equal(t, "value1", ctx.Properties["key1"])
	assert.Equal(t, 123, ctx.Properties["key2"])
	assert.Equal(t, true, ctx.Properties["key3"])
}

func TestContext_AddResponse(t *testing.T) {
	ctx := &Context{
		Event: Event{
			Type:    EventMessage,
			Content: "test",
		},
		Properties: make(map[string]interface{}),
		Responses:  []Response{},
	}

	// Add responses
	resp1 := Response{
		PluginName: "plugin1",
		Content:    "response1",
		Type:       "text",
		Data:       map[string]interface{}{"foo": "bar"},
	}

	resp2 := Response{
		PluginName: "plugin2",
		Content:    "response2",
		Type:       "file",
		Data:       map[string]interface{}{"path": "/tmp/file"},
	}

	ctx.Responses = append(ctx.Responses, resp1, resp2)

	assert.Len(t, ctx.Responses, 2)
	assert.Equal(t, "plugin1", ctx.Responses[0].PluginName)
	assert.Equal(t, "plugin2", ctx.Responses[1].PluginName)
}

func TestExecutionDecision(t *testing.T) {
	tests := []struct {
		name     string
		decision ExecutionDecision
		want     bool
	}{
		{
			name: "should execute",
			decision: ExecutionDecision{
				ShouldExecute: true,
				Reason:        "All conditions met",
			},
			want: true,
		},
		{
			name: "should not execute",
			decision: ExecutionDecision{
				ShouldExecute: false,
				Reason:        "Missing required property",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.decision.ShouldExecute)
			assert.NotEmpty(t, tt.decision.Reason)
		})
	}
}

func TestResponse_GetDataValue(t *testing.T) {
	resp := Response{
		PluginName: "test-plugin",
		Content:    "Test response",
		Type:       "data",
		Data: map[string]interface{}{
			"string_val": "hello",
			"int_val":    42,
			"bool_val":   true,
			"nested": map[string]interface{}{
				"key": "value",
			},
		},
	}

	// Test string value
	if val, ok := resp.Data["string_val"].(string); ok {
		assert.Equal(t, "hello", val)
	} else {
		t.Error("Failed to get string value")
	}

	// Test int value
	if val, ok := resp.Data["int_val"].(int); ok {
		assert.Equal(t, 42, val)
	} else {
		t.Error("Failed to get int value")
	}

	// Test bool value
	if val, ok := resp.Data["bool_val"].(bool); ok {
		assert.True(t, val)
	} else {
		t.Error("Failed to get bool value")
	}

	// Test nested value
	if nested, ok := resp.Data["nested"].(map[string]interface{}); ok {
		assert.Equal(t, "value", nested["key"])
	} else {
		t.Error("Failed to get nested value")
	}

	// Test missing key
	_, exists := resp.Data["missing"]
	assert.False(t, exists)
}
