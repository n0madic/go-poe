package sse

import (
	"fmt"
	"net/http"
)

// Writer writes Server-Sent Events to an http.ResponseWriter
type Writer struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewWriter creates a new SSE Writer and sets appropriate headers.
// Sets Content-Type: text/event-stream, Cache-Control: no-cache, Connection: keep-alive
func NewWriter(w http.ResponseWriter) *Writer {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)
	return &Writer{w: w, flusher: flusher}
}

// WriteEvent writes a single SSE event and flushes
func (sw *Writer) WriteEvent(e Event) error {
	if e.ID != "" {
		if _, err := fmt.Fprintf(sw.w, "id: %s\n", e.ID); err != nil {
			return err
		}
	}
	if e.Event != "" {
		if _, err := fmt.Fprintf(sw.w, "event: %s\n", e.Event); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(sw.w, "data: %s\n\n", e.Data); err != nil {
		return err
	}
	if sw.flusher != nil {
		sw.flusher.Flush()
	}
	return nil
}
