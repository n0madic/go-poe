package sse

// Event represents a Server-Sent Event
type Event struct {
	Event string // The event type (e.g., "text", "done", "meta")
	Data  string // The event data
	ID    string // Optional event ID
}
