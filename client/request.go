package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/n0madic/go-poe/sse"
	"github.com/n0madic/go-poe/types"
)

// performQueryRequest sends a query and parses SSE responses into the channel
func performQueryRequest(
	ctx context.Context,
	httpClient *http.Client,
	url string,
	payload map[string]any,
	headers map[string]string,
	ch chan<- *types.PartialResponse,
) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return &BotError{Message: fmt.Sprintf("failed to marshal request: %v", err)}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return &BotError{Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	// Set mandatory headers after custom headers to prevent override
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := httpClient.Do(req)
	if err != nil {
		return &BotError{Message: fmt.Sprintf("HTTP request failed: %v", err), Cause: err}
	}
	defer resp.Body.Close()

	reader := sse.NewReader(resp.Body)
	var chunks []string
	eventCount := 0
	errorReported := false
	hasTools := payload["tools"] != nil

	for {
		event, err := reader.ReadEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &BotError{Message: fmt.Sprintf("SSE read error: %v", err), Cause: err}
		}

		eventCount++

		// Parse index from data if present
		var index *int
		if event.Data != "" {
			var dataMap map[string]any
			if json.Unmarshal([]byte(event.Data), &dataMap) == nil {
				if idx, ok := dataMap["index"]; ok {
					if idxFloat, ok := idx.(float64); ok {
						idxInt := int(idxFloat)
						index = &idxInt
					}
				}
			}
		}

		switch event.Event {
		case "done":
			if len(chunks) == 0 && !errorReported && !hasTools {
				log.Printf("Bot returned no text in response")
			}
			return nil

		case "text":
			text, err := getJSONStringField(event.Data, "text")
			if err != nil {
				return err
			}
			chunks = append(chunks, text)
			ch <- &types.PartialResponse{Text: text, Index: index}

		case "replace_response":
			text, err := getJSONStringField(event.Data, "text")
			if err != nil {
				return err
			}
			chunks = nil
			ch <- &types.PartialResponse{Text: text, IsReplaceResponse: true, Index: index}

		case "file":
			var dataMap map[string]any
			if err := json.Unmarshal([]byte(event.Data), &dataMap); err != nil {
				return &BotErrorNoRetry{BotError{Message: "Invalid JSON in file event"}}
			}
			fileURL, _ := dataMap["url"].(string)
			contentType, _ := dataMap["content_type"].(string)
			name, _ := dataMap["name"].(string)
			var inlineRef *string
			if ref, ok := dataMap["inline_ref"].(string); ok {
				inlineRef = &ref
			}
			ch <- &types.PartialResponse{
				Text: "",
				Attachment: &types.Attachment{
					URL:         fileURL,
					ContentType: contentType,
					Name:        name,
					InlineRef:   inlineRef,
				},
				Index: index,
			}

		case "suggested_reply":
			text, err := getJSONStringField(event.Data, "text")
			if err != nil {
				return err
			}
			ch <- &types.PartialResponse{
				Text:             text,
				IsSuggestedReply: true,
				Index:            index,
			}

		case "json":
			var data map[string]any
			if err := json.Unmarshal([]byte(event.Data), &data); err != nil {
				return &BotErrorNoRetry{BotError{Message: "Invalid JSON in json event"}}
			}
			ch <- &types.PartialResponse{Text: "", Data: data, Index: index}

		case "meta":
			if eventCount != 1 {
				// meta event that is not the first event is ignored per spec
				continue
			}
			var dataMap map[string]any
			if err := json.Unmarshal([]byte(event.Data), &dataMap); err != nil {
				errorReported = true
				continue
			}
			linkify, _ := dataMap["linkify"].(bool)
			suggestedReplies, _ := dataMap["suggested_replies"].(bool)
			contentType := "text/markdown"
			if ct, ok := dataMap["content_type"].(string); ok {
				contentType = ct
			}
			meta := &types.MetaResponse{
				PartialResponse:  types.PartialResponse{Text: ""},
				Linkify:          linkify,
				SuggestedReplies: suggestedReplies,
				ContentType:      types.ContentType(contentType),
			}
			// Send meta as a PartialResponse with RawResponse carrying the meta info
			ch <- &types.PartialResponse{
				Text:        "",
				RawResponse: meta,
				Index:       index,
			}

		case "error":
			var dataMap map[string]any
			if err := json.Unmarshal([]byte(event.Data), &dataMap); err != nil {
				return &BotError{Message: event.Data}
			}
			allowRetry := true
			if ar, ok := dataMap["allow_retry"].(bool); ok {
				allowRetry = ar
			}
			if allowRetry {
				return &BotError{Message: event.Data}
			}
			return &BotErrorNoRetry{BotError{Message: event.Data}}

		case "ping":
			continue

		default:
			log.Printf("Unknown event type: %s", event.Event)
			errorReported = true
			continue
		}
	}

	log.Printf("Bot exited without sending 'done' event")
	return nil
}

func getJSONStringField(data, field string) (string, error) {
	var dataMap map[string]any
	if err := json.Unmarshal([]byte(data), &dataMap); err != nil {
		return "", &BotErrorNoRetry{BotError{Message: fmt.Sprintf("Invalid JSON in event: %s", data)}}
	}
	text, ok := dataMap[field].(string)
	if !ok {
		return "", &BotErrorNoRetry{BotError{Message: fmt.Sprintf("Expected string in '%s' field", field)}}
	}
	return text, nil
}
