package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		want      string
	}{
		{
			name:      "message event",
			eventType: EventMessage,
			want:      "message",
		},
		{
			name:      "command event",
			eventType: EventCommand,
			want:      "command",
		},
		{
			name:      "webhook event",
			eventType: EventWebhook,
			want:      "webhook",
		},
		{
			name:      "scheduled event",
			eventType: EventScheduled,
			want:      "scheduled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.eventType))
		})
	}
}

func TestEvent_Creation(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		check func(t *testing.T, e Event)
	}{
		{
			name: "discord message event",
			event: Event{
				Type:      EventMessage,
				Source:    "discord",
				Content:   "Hello, world!",
				UserID:    "user123",
				ChannelID: "general",
				Metadata: map[string]interface{}{
					"guild_id":    "guild456",
					"message_id":  "msg789",
					"is_bot":      false,
					"attachments": []string{"image.png"},
				},
			},
			check: func(t *testing.T, e Event) {
				assert.Equal(t, EventMessage, e.Type)
				assert.Equal(t, "discord", e.Source)
				assert.Equal(t, "Hello, world!", e.Content)
				assert.Equal(t, "user123", e.UserID)
				assert.Equal(t, "general", e.ChannelID)
				assert.NotNil(t, e.Metadata)
				assert.Equal(t, "guild456", e.Metadata["guild_id"])
				assert.False(t, e.Metadata["is_bot"].(bool))
			},
		},
		{
			name: "telegram command event",
			event: Event{
				Type:      EventCommand,
				Source:    "telegram",
				Content:   "/start",
				UserID:    "@john_doe",
				ChannelID: "private",
				Metadata: map[string]interface{}{
					"command":    "/start",
					"args":       []string{},
					"message_id": 12345,
				},
			},
			check: func(t *testing.T, e Event) {
				assert.Equal(t, EventCommand, e.Type)
				assert.Equal(t, "telegram", e.Source)
				assert.Equal(t, "/start", e.Content)
				assert.Equal(t, "@john_doe", e.UserID)
				assert.Equal(t, "private", e.ChannelID)
				assert.Equal(t, "/start", e.Metadata["command"])
			},
		},
		{
			name: "webhook event",
			event: Event{
				Type:      EventWebhook,
				Source:    "github",
				Content:   "push event",
				UserID:    "github-actions",
				ChannelID: "webhook",
				Metadata: map[string]interface{}{
					"repository": "owner/repo",
					"branch":     "main",
					"commit":     "abc123",
					"author":     "developer",
				},
			},
			check: func(t *testing.T, e Event) {
				assert.Equal(t, EventWebhook, e.Type)
				assert.Equal(t, "github", e.Source)
				assert.Equal(t, "owner/repo", e.Metadata["repository"])
				assert.Equal(t, "main", e.Metadata["branch"])
			},
		},
		{
			name: "scheduled event",
			event: Event{
				Type:      EventScheduled,
				Source:    "cron",
				Content:   "daily backup",
				UserID:    "system",
				ChannelID: "scheduler",
				Metadata: map[string]interface{}{
					"cron_expression": "0 2 * * *",
					"job_name":        "backup",
					"last_run":        "2024-01-01T02:00:00Z",
				},
			},
			check: func(t *testing.T, e Event) {
				assert.Equal(t, EventScheduled, e.Type)
				assert.Equal(t, "cron", e.Source)
				assert.Equal(t, "daily backup", e.Content)
				assert.Equal(t, "0 2 * * *", e.Metadata["cron_expression"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.event)
		})
	}
}

func TestEvent_MetadataOperations(t *testing.T) {
	event := Event{
		Type:     EventMessage,
		Source:   "test",
		Content:  "test content",
		Metadata: make(map[string]interface{}),
	}

	// Test adding metadata
	event.Metadata["key1"] = "value1"
	event.Metadata["key2"] = 123
	event.Metadata["key3"] = true
	event.Metadata["key4"] = []string{"a", "b", "c"}

	assert.Equal(t, "value1", event.Metadata["key1"])
	assert.Equal(t, 123, event.Metadata["key2"])
	assert.Equal(t, true, event.Metadata["key3"])
	assert.Len(t, event.Metadata["key4"].([]string), 3)

	// Test checking for existence
	_, exists := event.Metadata["key1"]
	assert.True(t, exists)

	_, exists = event.Metadata["nonexistent"]
	assert.False(t, exists)

	// Test updating metadata
	event.Metadata["key1"] = "updated_value"
	assert.Equal(t, "updated_value", event.Metadata["key1"])

	// Test deleting metadata
	delete(event.Metadata, "key1")
	_, exists = event.Metadata["key1"]
	assert.False(t, exists)
}

func TestEvent_EmptyMetadata(t *testing.T) {
	event := Event{
		Type:    EventMessage,
		Source:  "test",
		Content: "test",
		// Metadata not initialized
	}

	// Should be nil by default
	assert.Nil(t, event.Metadata)

	// Initialize it
	event.Metadata = make(map[string]interface{})
	assert.NotNil(t, event.Metadata)
	assert.Len(t, event.Metadata, 0)
}
