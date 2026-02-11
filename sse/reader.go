package sse

import (
	"bufio"
	"io"
	"strings"
)

// Reader reads Server-Sent Events from an io.Reader
type Reader struct {
	scanner *bufio.Scanner
}

// NewReader creates a new SSE Reader
func NewReader(r io.Reader) *Reader {
	return &Reader{scanner: bufio.NewScanner(r)}
}

// ReadEvent reads the next SSE event from the stream.
// Returns io.EOF when the stream is exhausted.
func (r *Reader) ReadEvent() (Event, error) {
	var event Event
	var dataLines []string
	hasData := false

	for r.scanner.Scan() {
		line := r.scanner.Text()

		// Empty line means end of event
		if line == "" {
			if hasData || event.Event != "" || event.ID != "" {
				event.Data = strings.Join(dataLines, "\n")
				return event, nil
			}
			continue
		}

		// Comment lines start with ':'
		if strings.HasPrefix(line, ":") {
			continue
		}

		// Parse field
		field, value, _ := strings.Cut(line, ":")
		// Remove single leading space from value if present
		value = strings.TrimPrefix(value, " ")

		switch field {
		case "event":
			event.Event = value
		case "data":
			dataLines = append(dataLines, value)
			hasData = true
		case "id":
			event.ID = value
		}
	}

	if err := r.scanner.Err(); err != nil {
		return Event{}, err
	}

	// If we have accumulated data, return it
	if hasData || event.Event != "" || event.ID != "" {
		event.Data = strings.Join(dataLines, "\n")
		return event, nil
	}

	return Event{}, io.EOF
}
