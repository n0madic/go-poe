package sse

import (
	"bytes"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func BenchmarkReaderSimpleEvent(b *testing.B) {
	input := `event: message
data: Hello, world!

`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := NewReader(strings.NewReader(input))
		_, err := reader.ReadEvent()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReaderMultiLineData(b *testing.B) {
	input := `event: multiline
data: Line 1
data: Line 2
data: Line 3
data: Line 4
data: Line 5

`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := NewReader(strings.NewReader(input))
		_, err := reader.ReadEvent()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReaderMultipleEvents(b *testing.B) {
	input := `event: first
data: First event

event: second
data: Second event

event: third
data: Third event

`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := NewReader(strings.NewReader(input))
		for {
			_, err := reader.ReadEvent()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkWriterSimpleEvent(b *testing.B) {
	event := Event{Event: "message", Data: "Hello, world!"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		writer := NewWriter(rec)
		err := writer.WriteEvent(event)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriterWithID(b *testing.B) {
	event := Event{ID: "123", Event: "message", Data: "Hello, world!"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		writer := NewWriter(rec)
		err := writer.WriteEvent(event)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWriterMultipleEvents(b *testing.B) {
	events := []Event{
		{Event: "first", Data: "First event"},
		{Event: "second", Data: "Second event"},
		{Event: "third", Data: "Third event"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		writer := NewWriter(rec)
		for _, e := range events {
			err := writer.WriteEvent(e)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	event := Event{ID: "123", Event: "message", Data: "Hello, world!"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		rec := httptest.NewRecorder()
		rec.Body = &buf

		writer := &Writer{w: rec}
		err := writer.WriteEvent(event)
		if err != nil {
			b.Fatal(err)
		}

		reader := NewReader(&buf)
		_, err = reader.ReadEvent()
		if err != nil {
			b.Fatal(err)
		}
	}
}
