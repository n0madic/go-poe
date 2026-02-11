package sse

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Event
	}{
		{
			name: "simple event with type and data",
			input: `event: message
data: Hello, world!

`,
			expected: []Event{
				{Event: "message", Data: "Hello, world!"},
			},
		},
		{
			name: "event with no type (just data)",
			input: `data: Just data

`,
			expected: []Event{
				{Data: "Just data"},
			},
		},
		{
			name: "multi-line data",
			input: `event: multiline
data: Line 1
data: Line 2
data: Line 3

`,
			expected: []Event{
				{Event: "multiline", Data: "Line 1\nLine 2\nLine 3"},
			},
		},
		{
			name: "events with comments",
			input: `: This is a comment
event: test
data: Some data
: Another comment

`,
			expected: []Event{
				{Event: "test", Data: "Some data"},
			},
		},
		{
			name: "event with ID",
			input: `id: 123
event: message
data: Hello

`,
			expected: []Event{
				{ID: "123", Event: "message", Data: "Hello"},
			},
		},
		{
			name: "empty data field",
			input: `event: ping
data:

`,
			expected: []Event{
				{Event: "ping", Data: ""},
			},
		},
		{
			name: "multiple events in sequence",
			input: `event: first
data: First event

event: second
data: Second event

data: Third event without type

`,
			expected: []Event{
				{Event: "first", Data: "First event"},
				{Event: "second", Data: "Second event"},
				{Data: "Third event without type"},
			},
		},
		{
			name: "no trailing newline",
			input: `event: message
data: Hello

event: last
data: No trailing newline`,
			expected: []Event{
				{Event: "message", Data: "Hello"},
				{Event: "last", Data: "No trailing newline"},
			},
		},
		{
			name: "data with colon and space handling",
			input: `data: value:with:colons

data:value_no_space

`,
			expected: []Event{
				{Data: "value:with:colons"},
				{Data: "value_no_space"},
			},
		},
		{
			name:     "empty stream",
			input:    "",
			expected: []Event{},
		},
		{
			name: "only comments and empty lines",
			input: `: comment
: another comment


`,
			expected: []Event{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewReader(strings.NewReader(tt.input))
			var events []Event

			for {
				event, err := reader.ReadEvent()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				events = append(events, event)
			}

			if len(events) != len(tt.expected) {
				t.Fatalf("expected %d events, got %d", len(tt.expected), len(events))
			}

			for i, event := range events {
				if event.Event != tt.expected[i].Event {
					t.Errorf("event %d: expected Event=%q, got %q", i, tt.expected[i].Event, event.Event)
				}
				if event.Data != tt.expected[i].Data {
					t.Errorf("event %d: expected Data=%q, got %q", i, tt.expected[i].Data, event.Data)
				}
				if event.ID != tt.expected[i].ID {
					t.Errorf("event %d: expected ID=%q, got %q", i, tt.expected[i].ID, event.ID)
				}
			}
		})
	}
}

func TestWriter(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		expected string
	}{
		{
			name:     "simple event output format",
			event:    Event{Event: "message", Data: "Hello"},
			expected: "event: message\ndata: Hello\n\n",
		},
		{
			name:     "event with ID",
			event:    Event{ID: "123", Event: "message", Data: "Test"},
			expected: "id: 123\nevent: message\ndata: Test\n\n",
		},
		{
			name:     "event with no type (just data)",
			event:    Event{Data: "Just data"},
			expected: "data: Just data\n\n",
		},
		{
			name:     "event with ID but no type",
			event:    Event{ID: "456", Data: "Data only"},
			expected: "id: 456\ndata: Data only\n\n",
		},
		{
			name:     "empty data",
			event:    Event{Event: "ping", Data: ""},
			expected: "event: ping\ndata: \n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writer := NewWriter(rec)

			err := writer.WriteEvent(tt.event)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := rec.Body.String()
			if got != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, got)
			}
		})
	}
}

func TestWriterHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	_ = NewWriter(rec)

	headers := map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}

	for key, expected := range headers {
		got := rec.Header().Get(key)
		if got != expected {
			t.Errorf("header %q: expected %q, got %q", key, expected, got)
		}
	}
}

func TestWriterFlush(t *testing.T) {
	// Create a custom ResponseWriter that tracks flush calls
	flushed := false
	fw := &flushWriter{
		ResponseWriter: httptest.NewRecorder(),
		onFlush: func() {
			flushed = true
		},
	}

	writer := NewWriter(fw)
	err := writer.WriteEvent(Event{Data: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !flushed {
		t.Error("expected Flush to be called")
	}
}

// flushWriter is a helper type for testing flush behavior
type flushWriter struct {
	http.ResponseWriter
	onFlush func()
}

func (fw *flushWriter) Flush() {
	fw.onFlush()
}

func TestWriterError(t *testing.T) {
	// Test write error handling
	ew := &errorWriter{}
	writer := &Writer{w: ew}

	err := writer.WriteEvent(Event{Data: "test"})
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// errorWriter is a helper type that always returns an error
type errorWriter struct{}

func (ew *errorWriter) Header() http.Header {
	return http.Header{}
}

func (ew *errorWriter) Write([]byte) (int, error) {
	return 0, io.ErrClosedPipe
}

func (ew *errorWriter) WriteHeader(statusCode int) {}

func TestReaderWriterRoundTrip(t *testing.T) {
	// Test that we can write events and read them back
	var buf bytes.Buffer
	rec := httptest.NewRecorder()
	rec.Body = &buf

	writer := &Writer{w: rec}

	events := []Event{
		{Event: "message", Data: "Hello"},
		{ID: "123", Event: "update", Data: "Update message"},
		{Data: "No type"},
	}

	for _, e := range events {
		if err := writer.WriteEvent(e); err != nil {
			t.Fatalf("write error: %v", err)
		}
	}

	reader := NewReader(&buf)
	var readEvents []Event

	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read error: %v", err)
		}
		readEvents = append(readEvents, event)
	}

	if len(readEvents) != len(events) {
		t.Fatalf("expected %d events, got %d", len(events), len(readEvents))
	}

	for i := range events {
		if readEvents[i].Event != events[i].Event {
			t.Errorf("event %d: expected Event=%q, got %q", i, events[i].Event, readEvents[i].Event)
		}
		if readEvents[i].Data != events[i].Data {
			t.Errorf("event %d: expected Data=%q, got %q", i, events[i].Data, readEvents[i].Data)
		}
		if readEvents[i].ID != events[i].ID {
			t.Errorf("event %d: expected ID=%q, got %q", i, events[i].ID, readEvents[i].ID)
		}
	}
}
