package types

// Context carries data through the plugin pipeline
type Context struct {
	Event      Event                  `json:"event"`
	Properties map[string]interface{} `json:"properties"` // Shared data between plugins
	Responses  []Response             `json:"responses"`  // Accumulated responses from plugins
}

// Response represents what a plugin wants to send back
type Response struct {
	PluginName string                 `json:"plugin_name"`
	Content    string                 `json:"content"`
	Type       string                 `json:"type"` // "text", "file", "embed", etc.
	Data       map[string]interface{} `json:"data"`
}

// ExecutionDecision tells whether a plugin should run
type ExecutionDecision struct {
	ShouldExecute bool   `json:"should_execute"`
	Reason        string `json:"reason"`
}
