package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/n0madic/go-poe/sse"
	"github.com/n0madic/go-poe/types"
)

// testBot is a simple bot for testing
type testBot struct {
	*BasePoeBot
	responseText string
}

func newTestBot(path, accessKey, botName, responseText string) *testBot {
	return &testBot{
		BasePoeBot:   NewBasePoeBot(path, accessKey, botName),
		responseText: responseText,
	}
}

func (b *testBot) GetResponse(ctx context.Context, req *types.QueryRequest) <-chan types.BotEvent {
	ch := make(chan types.BotEvent, 1)
	go func() {
		defer close(ch)
		ch <- &types.PartialResponse{Text: b.responseText}
	}()
	return ch
}

func TestHandlerReturnsHTMLOnGET(t *testing.T) {
	bot := newTestBot("/", "", "", "test")
	handler := botHandler(bot)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Errorf("Expected HTML in response, got: %s", body)
	}
	if !strings.Contains(body, "Go Poe bot server") {
		t.Errorf("Expected 'Go Poe bot server' in response, got: %s", body)
	}
}

func TestHandlerReturns401OnBadAuth(t *testing.T) {
	bot := newTestBot("/", "secret123", "", "test")
	handler := botHandler(bot)

	reqBody := `{"version":"1.2","type":"query","query":[{"role":"user","content":"hi"}],"user_id":"u1","conversation_id":"c1","message_id":"m1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrongkey")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Invalid access key") {
		t.Errorf("Expected 'Invalid access key' in response, got: %s", body)
	}
}

func TestHandlerReturns200OnValidSettingsRequest(t *testing.T) {
	bot := newTestBot("/", "secret123", "testbot", "test")
	handler := botHandler(bot)

	reqBody := `{"version":"1.2","type":"settings"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if version, ok := response["response_version"].(float64); !ok || version != 2 {
		t.Errorf("Expected response_version=2, got: %v", response["response_version"])
	}
}

func TestHandlerStreamsSSEForQueryRequest(t *testing.T) {
	bot := newTestBot("/", "secret123", "testbot", "Hello world")
	handler := botHandler(bot)

	reqBody := `{"version":"1.2","type":"query","query":[{"role":"user","content":"hi"}],"user_id":"u1","conversation_id":"c1","message_id":"m1"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Errorf("Expected Content-Type 'text/event-stream', got '%s'", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "event: text") {
		t.Errorf("Expected 'event: text' in SSE stream, got: %s", body)
	}
	if !strings.Contains(body, "Hello world") {
		t.Errorf("Expected 'Hello world' in SSE stream, got: %s", body)
	}
	if !strings.Contains(body, "event: done") {
		t.Errorf("Expected 'event: done' in SSE stream, got: %s", body)
	}
}

func TestInsertAttachmentMessagesWithTextAttachment(t *testing.T) {
	parsedContent := "This is the content of the text file."
	req := &types.QueryRequest{
		Query: []types.ProtocolMessage{
			{Role: "user", Content: "Check this file"},
			{
				Role:    "user",
				Content: "Process this",
				Attachments: []types.Attachment{
					{
						Name:          "file.txt",
						ContentType:   "text/plain",
						ParsedContent: &parsedContent,
					},
				},
			},
		},
	}

	result := InsertAttachmentMessages(req)

	// Should have: original first message + attachment message + last message
	if len(result.Query) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(result.Query))
	}

	attachmentMsg := result.Query[1]
	if !strings.Contains(attachmentMsg.Content, "file.txt") {
		t.Errorf("Expected attachment message to contain filename, got: %s", attachmentMsg.Content)
	}
	if !strings.Contains(attachmentMsg.Content, parsedContent) {
		t.Errorf("Expected attachment message to contain parsed content, got: %s", attachmentMsg.Content)
	}
}

func TestInsertAttachmentMessagesWithHTMLAttachment(t *testing.T) {
	parsedContent := "<html><body>Web content</body></html>"
	req := &types.QueryRequest{
		Query: []types.ProtocolMessage{
			{
				Role:    "user",
				Content: "Process this webpage",
				Attachments: []types.Attachment{
					{
						Name:          "page.html",
						ContentType:   "text/html",
						ParsedContent: &parsedContent,
					},
				},
			},
		},
	}

	result := InsertAttachmentMessages(req)

	// Should have: attachment message + last message
	if len(result.Query) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(result.Query))
	}

	attachmentMsg := result.Query[0]
	if !strings.Contains(attachmentMsg.Content, "Assume you can access the external URL") {
		t.Errorf("Expected URL attachment template, got: %s", attachmentMsg.Content)
	}
	if !strings.Contains(attachmentMsg.Content, parsedContent) {
		t.Errorf("Expected attachment message to contain parsed content, got: %s", attachmentMsg.Content)
	}
}

func TestInsertAttachmentMessagesWithImageAttachment(t *testing.T) {
	parsedContent := "photo.jpg***A beautiful sunset over the ocean"
	req := &types.QueryRequest{
		Query: []types.ProtocolMessage{
			{
				Role:    "user",
				Content: "What's in this image?",
				Attachments: []types.Attachment{
					{
						Name:          "photo.jpg",
						ContentType:   "image/jpeg",
						ParsedContent: &parsedContent,
					},
				},
			},
		},
	}

	result := InsertAttachmentMessages(req)

	// Should have: attachment message + last message
	if len(result.Query) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(result.Query))
	}

	attachmentMsg := result.Query[0]
	if !strings.Contains(attachmentMsg.Content, "I have uploaded an image") {
		t.Errorf("Expected image attachment template, got: %s", attachmentMsg.Content)
	}
	if !strings.Contains(attachmentMsg.Content, "A beautiful sunset over the ocean") {
		t.Errorf("Expected image description in message, got: %s", attachmentMsg.Content)
	}
}

func TestInsertAttachmentMessagesWithPDFAttachment(t *testing.T) {
	parsedContent := "This is the extracted PDF text content."
	req := &types.QueryRequest{
		Query: []types.ProtocolMessage{
			{
				Role:    "user",
				Content: "Summarize this PDF",
				Attachments: []types.Attachment{
					{
						Name:          "document.pdf",
						ContentType:   "application/pdf",
						ParsedContent: &parsedContent,
					},
				},
			},
		},
	}

	result := InsertAttachmentMessages(req)

	// Should have: attachment message + last message
	if len(result.Query) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(result.Query))
	}

	attachmentMsg := result.Query[0]
	if !strings.Contains(attachmentMsg.Content, "document.pdf") {
		t.Errorf("Expected attachment message to contain filename, got: %s", attachmentMsg.Content)
	}
	if !strings.Contains(attachmentMsg.Content, parsedContent) {
		t.Errorf("Expected attachment message to contain parsed content, got: %s", attachmentMsg.Content)
	}
}

func TestMakePromptAuthorRoleAlternatedMergesConsecutiveSameRole(t *testing.T) {
	messages := []types.ProtocolMessage{
		{Role: "user", Content: "First message"},
		{Role: "user", Content: "Second message"},
		{Role: "bot", Content: "Bot response"},
		{Role: "user", Content: "Third message"},
	}

	result := MakePromptAuthorRoleAlternated(messages)

	if len(result) != 3 {
		t.Fatalf("Expected 3 messages after merging, got %d", len(result))
	}

	if result[0].Role != "user" {
		t.Errorf("Expected first message role to be 'user', got '%s'", result[0].Role)
	}
	if !strings.Contains(result[0].Content, "First message") {
		t.Errorf("Expected merged content to contain 'First message', got: %s", result[0].Content)
	}
	if !strings.Contains(result[0].Content, "Second message") {
		t.Errorf("Expected merged content to contain 'Second message', got: %s", result[0].Content)
	}

	if result[1].Role != "bot" {
		t.Errorf("Expected second message role to be 'bot', got '%s'", result[1].Role)
	}

	if result[2].Role != "user" {
		t.Errorf("Expected third message role to be 'user', got '%s'", result[2].Role)
	}
}

func TestMakePromptAuthorRoleAlternatedDeduplicatesAttachmentsByURL(t *testing.T) {
	messages := []types.ProtocolMessage{
		{
			Role:    "user",
			Content: "First",
			Attachments: []types.Attachment{
				{URL: "http://example.com/file1.txt", Name: "file1.txt"},
				{URL: "http://example.com/file2.txt", Name: "file2.txt"},
			},
		},
		{
			Role:    "user",
			Content: "Second",
			Attachments: []types.Attachment{
				{URL: "http://example.com/file1.txt", Name: "file1.txt"}, // duplicate
				{URL: "http://example.com/file3.txt", Name: "file3.txt"},
			},
		},
	}

	result := MakePromptAuthorRoleAlternated(messages)

	if len(result) != 1 {
		t.Fatalf("Expected 1 message after merging, got %d", len(result))
	}

	attachments := result[0].Attachments
	if len(attachments) != 3 {
		t.Fatalf("Expected 3 unique attachments, got %d", len(attachments))
	}

	// Check that we have all three unique URLs
	urls := make(map[string]bool)
	for _, att := range attachments {
		urls[att.URL] = true
	}

	expectedURLs := []string{
		"http://example.com/file1.txt",
		"http://example.com/file2.txt",
		"http://example.com/file3.txt",
	}
	for _, url := range expectedURLs {
		if !urls[url] {
			t.Errorf("Expected URL %s in attachments", url)
		}
	}
}

func TestHandlerReportFeedback(t *testing.T) {
	bot := newTestBot("/", "secret123", "testbot", "test")
	handler := botHandler(bot)

	reqBody := `{"version":"1.2","type":"report_feedback","message_id":"m1","user_id":"u1","conversation_id":"c1","feedback_type":"like"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "{}" {
		t.Errorf("Expected empty JSON object, got: %s", body)
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	bot := newTestBot("/", "", "", "test")
	handler := botHandler(bot)

	req := httptest.NewRequest(http.MethodPut, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestFindAccessKey(t *testing.T) {
	// Test with provided key
	key := FindAccessKey("mykey")
	if key != "mykey" {
		t.Errorf("Expected 'mykey', got '%s'", key)
	}

	// Test with env var
	os.Setenv("POE_ACCESS_KEY", "envkey")
	defer os.Unsetenv("POE_ACCESS_KEY")

	key = FindAccessKey("")
	if key != "envkey" {
		t.Errorf("Expected 'envkey' from env, got '%s'", key)
	}

	// Test priority (param over env)
	key = FindAccessKey("paramkey")
	if key != "paramkey" {
		t.Errorf("Expected 'paramkey' (param should override env), got '%s'", key)
	}
}

func TestMakeAppMultipleBots(t *testing.T) {
	bot1 := newTestBot("/bot1", "", "", "response1")
	bot2 := newTestBot("/bot2", "", "", "response2")

	app := MakeApp(bot1, bot2)

	// Test bot1
	reqBody := `{"version":"1.2","type":"query","query":[{"role":"user","content":"hi"}],"user_id":"u1","conversation_id":"c1","message_id":"m1"}`
	req := httptest.NewRequest(http.MethodPost, "/bot1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for bot1, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "response1") {
		t.Errorf("Expected 'response1' in bot1 response, got: %s", body)
	}

	// Test bot2
	req = httptest.NewRequest(http.MethodPost, "/bot2", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for bot2, got %d", w.Code)
	}

	body = w.Body.String()
	if !strings.Contains(body, "response2") {
		t.Errorf("Expected 'response2' in bot2 response, got: %s", body)
	}
}

func TestMakeAppPanicsOnDuplicatePaths(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for duplicate paths, but didn't panic")
		}
	}()

	bot1 := newTestBot("/same", "", "", "response1")
	bot2 := newTestBot("/same", "", "", "response2")

	MakeApp(bot1, bot2)
}

func TestBasePoeBot(t *testing.T) {
	bot := NewBasePoeBot("/test", "key123", "mybot")

	if bot.Path() != "/test" {
		t.Errorf("Expected path '/test', got '%s'", bot.Path())
	}
	if bot.AccessKey() != "key123" {
		t.Errorf("Expected access key 'key123', got '%s'", bot.AccessKey())
	}
	if bot.BotName() != "mybot" {
		t.Errorf("Expected bot name 'mybot', got '%s'", bot.BotName())
	}
	if !bot.ShouldInsertAttachmentMessages() {
		t.Error("Expected ShouldInsertAttachmentMessages to be true by default")
	}

	// Test SetAccessKey and SetBotName
	bot.SetAccessKey("newkey")
	bot.SetBotName("newname")

	if bot.AccessKey() != "newkey" {
		t.Errorf("Expected access key 'newkey' after SetAccessKey, got '%s'", bot.AccessKey())
	}
	if bot.BotName() != "newname" {
		t.Errorf("Expected bot name 'newname' after SetBotName, got '%s'", bot.BotName())
	}

	// Test default GetResponse
	ctx := context.Background()
	ch := bot.GetResponse(ctx, &types.QueryRequest{})
	response := <-ch
	if pr, ok := response.(*types.PartialResponse); !ok || pr.Text != "hello" {
		t.Errorf("Expected default response 'hello', got: %v", response)
	}

	// Test GetSettings
	settings, err := bot.GetSettings(ctx, &types.SettingsRequest{})
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}
	if settings.ResponseVersion == nil || *settings.ResponseVersion != 2 {
		t.Errorf("Expected default settings version 2")
	}

	// Test OnFeedback, OnReaction, OnError (should not error)
	if err := bot.OnFeedback(ctx, &types.ReportFeedbackRequest{}); err != nil {
		t.Errorf("OnFeedback returned error: %v", err)
	}
	if err := bot.OnReaction(ctx, &types.ReportReactionRequest{}); err != nil {
		t.Errorf("OnReaction returned error: %v", err)
	}
	if err := bot.OnError(ctx, &types.ReportErrorRequest{}); err != nil {
		t.Errorf("OnError returned error: %v", err)
	}
}

func TestWriteEventFunctions(t *testing.T) {
	// Test that write functions don't panic
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sseWriter := sse.NewWriter(w)

		writeTextEvent(sseWriter, "test text", nil)
		index := 1
		writeTextEvent(sseWriter, "indexed text", &index)
		writeReplaceResponseEvent(sseWriter, "replace")
		writeSuggestedReplyEvent(sseWriter, "suggestion")
		writeFileEvent(sseWriter, &types.Attachment{
			URL:         "http://example.com/file.txt",
			ContentType: "text/plain",
			Name:        "file.txt",
		})
		writeMetaEvent(sseWriter, types.NewMetaResponse())
		writeDataEvent(sseWriter, "metadata")
		errorType := "test_error"
		writeErrorEvent(sseWriter, "error text", true, &errorType)
		writeDoneEvent(sseWriter)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	expectedEvents := []string{
		"event: text",
		"test text",
		"indexed text",
		"event: replace_response",
		"event: suggested_reply",
		"event: file",
		"event: meta",
		"event: data",
		"event: error",
		"event: done",
	}

	for _, expected := range expectedEvents {
		if !strings.Contains(bodyStr, expected) {
			t.Errorf("Expected '%s' in SSE stream, got: %s", expected, bodyStr)
		}
	}
}
