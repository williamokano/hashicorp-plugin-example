package types

// EventType represents different types of events plugins can handle
type EventType string

const (
	EventMessage   EventType = "message"
	EventCommand   EventType = "command"
	EventWebhook   EventType = "webhook"
	EventScheduled EventType = "scheduled"
)

// Event represents an incoming event that plugins will process
type Event struct {
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"`  // e.g., "discord", "telegram", "slack"
	Content   string                 `json:"content"` // The actual message/command
	UserID    string                 `json:"user_id"`
	ChannelID string                 `json:"channel_id"`
	Metadata  map[string]interface{} `json:"metadata"` // Additional event-specific data
}
