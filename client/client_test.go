package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/n0madic/go-poe/types"
)

// mockSSEServer creates a test server that responds with SSE events
func mockSSEServer(events []string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		for _, event := range events {
			fmt.Fprint(w, event)
			flusher.Flush()
		}
	}))
}

func TestStreamRequest_TextEvents(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"Hello\"}\n\n",
		"event: text\ndata: {\"text\": \" world\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query: []types.ProtocolMessage{
			{Role: "user", Content: "test"},
		},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var texts []string
	for msg := range ch {
		texts = append(texts, msg.Text)
	}

	expected := []string{"Hello", " world"}
	if len(texts) != len(expected) {
		t.Fatalf("Expected %d messages, got %d", len(expected), len(texts))
	}
	for i, text := range texts {
		if text != expected[i] {
			t.Errorf("Message %d: expected %q, got %q", i, expected[i], text)
		}
	}
}

func TestStreamRequest_ReplaceResponse(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"Old text\"}\n\n",
		"event: replace_response\ndata: {\"text\": \"New text\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var messages []*types.PartialResponse
	for msg := range ch {
		messages = append(messages, msg)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Text != "Old text" || messages[0].IsReplaceResponse {
		t.Errorf("First message: expected text='Old text', is_replace=false, got text=%q, is_replace=%v",
			messages[0].Text, messages[0].IsReplaceResponse)
	}

	if messages[1].Text != "New text" || !messages[1].IsReplaceResponse {
		t.Errorf("Second message: expected text='New text', is_replace=true, got text=%q, is_replace=%v",
			messages[1].Text, messages[1].IsReplaceResponse)
	}
}

func TestStreamRequest_SuggestedReply(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"Main response\"}\n\n",
		"event: suggested_reply\ndata: {\"text\": \"Try this\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var messages []*types.PartialResponse
	for msg := range ch {
		messages = append(messages, msg)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[1].Text != "Try this" || !messages[1].IsSuggestedReply {
		t.Errorf("Expected suggested reply, got text=%q, is_suggested=%v",
			messages[1].Text, messages[1].IsSuggestedReply)
	}
}

func TestStreamRequest_FileEvent(t *testing.T) {
	events := []string{
		"event: file\ndata: {\"url\": \"https://example.com/file.pdf\", \"content_type\": \"application/pdf\", \"name\": \"doc.pdf\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var messages []*types.PartialResponse
	for msg := range ch {
		messages = append(messages, msg)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	att := messages[0].Attachment
	if att == nil {
		t.Fatal("Expected attachment, got nil")
	}

	if att.URL != "https://example.com/file.pdf" {
		t.Errorf("Expected URL https://example.com/file.pdf, got %s", att.URL)
	}
	if att.ContentType != "application/pdf" {
		t.Errorf("Expected content_type application/pdf, got %s", att.ContentType)
	}
	if att.Name != "doc.pdf" {
		t.Errorf("Expected name doc.pdf, got %s", att.Name)
	}
}

func TestStreamRequest_ErrorEventRetry(t *testing.T) {
	events := []string{
		"event: error\ndata: {\"allow_retry\": true, \"text\": \"Server error\"}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:        server.URL + "/",
		HTTPClient:     &http.Client{Timeout: 5 * time.Second},
		NumTries:       1,
		RetrySleepTime: 10 * time.Millisecond,
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var count int
	for range ch {
		count++
	}

	// Should get no messages since error occurred
	if count != 0 {
		t.Errorf("Expected 0 messages after error, got %d", count)
	}
}

func TestStreamRequest_ErrorEventNoRetry(t *testing.T) {
	events := []string{
		"event: error\ndata: {\"allow_retry\": false, \"text\": \"Bad request\"}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:        server.URL + "/",
		HTTPClient:     &http.Client{Timeout: 5 * time.Second},
		NumTries:       3,
		RetrySleepTime: 10 * time.Millisecond,
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var count int
	for range ch {
		count++
	}

	// Should get no messages and no retries
	if count != 0 {
		t.Errorf("Expected 0 messages after no-retry error, got %d", count)
	}
}

func TestStreamRequest_MetaEvent(t *testing.T) {
	events := []string{
		"event: meta\ndata: {\"linkify\": true, \"suggested_replies\": false, \"content_type\": \"text/plain\"}\n\n",
		"event: text\ndata: {\"text\": \"Response text\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var messages []*types.PartialResponse
	for msg := range ch {
		messages = append(messages, msg)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	// First message should be meta
	if messages[0].RawResponse == nil {
		t.Fatal("Expected RawResponse to contain meta")
	}

	meta, ok := messages[0].RawResponse.(*types.MetaResponse)
	if !ok {
		t.Fatal("Expected RawResponse to be MetaResponse")
	}

	if !meta.Linkify {
		t.Error("Expected linkify=true")
	}
	if meta.SuggestedReplies {
		t.Error("Expected suggested_replies=false")
	}
	if meta.ContentType != "text/plain" {
		t.Errorf("Expected content_type=text/plain, got %s", meta.ContentType)
	}

	// Second message should be text
	if messages[1].Text != "Response text" {
		t.Errorf("Expected text='Response text', got %q", messages[1].Text)
	}
}

func TestStreamRequest_DoneEvent(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"Complete\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var messages []*types.PartialResponse
	for msg := range ch {
		messages = append(messages, msg)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message before done, got %d", len(messages))
	}

	if messages[0].Text != "Complete" {
		t.Errorf("Expected text='Complete', got %q", messages[0].Text)
	}
}

func TestGetFinalResponse_CollectsAllText(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"Hello\"}\n\n",
		"event: text\ndata: {\"text\": \" world\"}\n\n",
		"event: text\ndata: {\"text\": \"!\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	result, err := GetFinalResponse(context.Background(), req, "testbot", "", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Hello world!"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestGetFinalResponse_HandlesReplaceResponse(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"First\"}\n\n",
		"event: text\ndata: {\"text\": \" response\"}\n\n",
		"event: replace_response\ndata: {\"text\": \"Replaced\"}\n\n",
		"event: text\ndata: {\"text\": \" text\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	result, err := GetFinalResponse(context.Background(), req, "testbot", "", opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "Replaced text"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestToolCallDeltaAggregation(t *testing.T) {
	// Simulate tool call deltas
	events := []string{
		`event: json
data: {"choices": [{"delta": {"tool_calls": [{"index": 0, "id": "call_1", "type": "function", "function": {"name": "get_weather", "arguments": ""}}]}, "finish_reason": null}]}

`,
		`event: json
data: {"choices": [{"delta": {"tool_calls": [{"index": 0, "function": {"arguments": "{\"location\":"}}]}, "finish_reason": null}]}

`,
		`event: json
data: {"choices": [{"delta": {"tool_calls": [{"index": 0, "function": {"arguments": " \"Paris\"}"}}]}, "finish_reason": null}]}

`,
		`event: json
data: {"choices": [{"finish_reason": "tool_calls"}]}

`,
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "What's the weather in Paris?"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	tools := []types.ToolDefinition{
		{
			Type: "function",
			Function: types.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get weather for a location",
				Parameters: types.ParametersDefinition{
					Type: "object",
					Properties: map[string]any{
						"location": map[string]any{"type": "string"},
					},
					Required: []string{"location"},
				},
			},
		},
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		Tools:      tools,
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var toolCalls []types.ToolCallDefinitionDelta
	for msg := range ch {
		if len(msg.ToolCalls) > 0 {
			toolCalls = append(toolCalls, msg.ToolCalls...)
		}
	}

	// We should receive the deltas
	if len(toolCalls) == 0 {
		t.Fatal("Expected tool call deltas, got none")
	}
}

func TestSyncBotSettings_WithSettings(t *testing.T) {
	receivedSettings := make(map[string]any)
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedSettings)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	settings := map[string]any{
		"introduction_message": "Hello!",
		"server_bot_dependencies": map[string]int{
			"GPT-4": 1,
		},
	}

	err := SyncBotSettings("testbot", "test-key", settings, server.URL+"/")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPath := fmt.Sprintf("/update_settings/testbot/test-key/%s", types.ProtocolVersion)
	if receivedPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, receivedPath)
	}

	if intro, ok := receivedSettings["introduction_message"].(string); !ok || intro != "Hello!" {
		t.Errorf("Expected introduction_message='Hello!', got %v", receivedSettings["introduction_message"])
	}
}

func TestSyncBotSettings_WithoutSettings(t *testing.T) {
	var receivedPath string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := SyncBotSettings("testbot", "test-key", nil, server.URL+"/")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedPath := fmt.Sprintf("/fetch_settings/testbot/test-key/%s", types.ProtocolVersion)
	if receivedPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, receivedPath)
	}
}

func TestStreamRequest_Index(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"First\", \"index\": 0}\n\n",
		"event: text\ndata: {\"text\": \"Second\", \"index\": 1}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var messages []*types.PartialResponse
	for msg := range ch {
		messages = append(messages, msg)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].Index == nil || *messages[0].Index != 0 {
		t.Errorf("Expected index 0, got %v", messages[0].Index)
	}

	if messages[1].Index == nil || *messages[1].Index != 1 {
		t.Errorf("Expected index 1, got %v", messages[1].Index)
	}
}

func TestStreamRequest_JsonEvent(t *testing.T) {
	jsonData := map[string]any{
		"key":   "value",
		"count": float64(42),
	}
	jsonBytes, _ := json.Marshal(jsonData)

	events := []string{
		fmt.Sprintf("event: json\ndata: %s\n\n", string(jsonBytes)),
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          []types.ProtocolMessage{{Role: "user", Content: "test"}},
		UserID:         "test-user",
		ConversationID: "test-conv",
		MessageID:      "test-msg",
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := StreamRequest(context.Background(), req, "testbot", opts)

	var messages []*types.PartialResponse
	for msg := range ch {
		messages = append(messages, msg)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].Data == nil {
		t.Fatal("Expected Data field to be populated")
	}

	if messages[0].Data["key"] != "value" {
		t.Errorf("Expected Data['key']='value', got %v", messages[0].Data["key"])
	}

	if messages[0].Data["count"] != float64(42) {
		t.Errorf("Expected Data['count']=42, got %v", messages[0].Data["count"])
	}
}

func TestGetBotResponse(t *testing.T) {
	events := []string{
		"event: text\ndata: {\"text\": \"Response\"}\n\n",
		"event: done\ndata: {}\n\n",
	}

	server := mockSSEServer(events)
	defer server.Close()

	messages := []types.ProtocolMessage{
		{Role: "user", Content: "Hello"},
	}

	opts := &StreamRequestOptions{
		BaseURL:    server.URL + "/",
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	}

	ch := GetBotResponse(context.Background(), messages, "testbot", "test-key", opts)

	var count int
	for range ch {
		count++
	}

	if count != 1 {
		t.Errorf("Expected 1 message, got %d", count)
	}
}

func TestBotErrorNoRetry_Type(t *testing.T) {
	err := &BotErrorNoRetry{BotError{Message: "test error"}}

	if !IsBotErrorNoRetry(err) {
		t.Error("Expected IsBotErrorNoRetry to return true")
	}

	regularErr := &BotError{Message: "regular error"}
	if IsBotErrorNoRetry(regularErr) {
		t.Error("Expected IsBotErrorNoRetry to return false for regular BotError")
	}
}

func TestUploadFile_RequiresAPIKey(t *testing.T) {
	opts := &UploadFileOptions{
		FileURL:  "https://example.com/file.txt",
		FileName: "test.txt",
		APIKey:   "",
	}

	_, err := UploadFile(context.Background(), opts)
	if err == nil {
		t.Fatal("Expected error when APIKey is missing")
	}

	if !strings.Contains(err.Error(), "api_key is required") {
		t.Errorf("Expected 'api_key is required' error, got: %v", err)
	}
}

func TestUploadFile_RequiresFileOrURL(t *testing.T) {
	opts := &UploadFileOptions{
		APIKey: "test-key",
	}

	_, err := UploadFile(context.Background(), opts)
	if err == nil {
		t.Fatal("Expected error when neither File nor FileURL is provided")
	}

	if !strings.Contains(err.Error(), "provide either File or FileURL") {
		t.Errorf("Expected 'provide either File or FileURL' error, got: %v", err)
	}
}

func TestUploadFile_NotBoth(t *testing.T) {
	opts := &UploadFileOptions{
		APIKey:   "test-key",
		File:     strings.NewReader("content"),
		FileURL:  "https://example.com/file.txt",
		FileName: "test.txt",
	}

	_, err := UploadFile(context.Background(), opts)
	if err == nil {
		t.Fatal("Expected error when both File and FileURL are provided")
	}

	if !strings.Contains(err.Error(), "not both") {
		t.Errorf("Expected 'not both' error, got: %v", err)
	}
}
