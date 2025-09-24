package structs

import "time"

// Event is the canonical log record for NEON.
type Event struct {
	Timestamp int64             `json:"timestamp"`
	Type      string            `json:"type"`
	Source    string            `json:"source"`
	Payload   map[string]any    `json:"payload,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
}

// NewEvent helper
func NewEvent(etype, source string, payload map[string]any) Event {
	return Event{
		Timestamp: time.Now().Unix(),
		Type:      etype,
		Source:    source,
		Payload:   payload,
	}
}
