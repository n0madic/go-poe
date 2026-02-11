// Package client provides a Go client for the Poe bot protocol.
//
// This package implements the Bot Query API for calling other Poe bots
// with support for SSE streaming, OpenAI-compatible tool calling, file uploads,
// and settings synchronization.
//
// # Basic Usage
//
// Query a bot and stream the response:
//
//	messages := []types.ProtocolMessage{
//	    {Role: "user", Content: "Hello!"},
//	}
//	ch := client.GetBotResponse(ctx, messages, "GPT-4o", apiKey, nil)
//	for response := range ch {
//	    fmt.Print(response.Text)
//	}
//
// Get the final response without streaming:
//
//	req := &types.QueryRequest{
//	    BaseRequest: types.BaseRequest{
//	        Version: types.ProtocolVersion,
//	        Type:    types.RequestTypeQuery,
//	    },
//	    Query:          messages,
//	    UserID:         "user-123",
//	    ConversationID: "conv-123",
//	    MessageID:      "msg-123",
//	}
//	finalText, err := client.GetFinalResponse(ctx, req, "GPT-4o", apiKey, nil)
//
// # Tool Calling
//
// Define tools and executables for OpenAI-compatible function calling:
//
//	tools := []types.ToolDefinition{
//	    {
//	        Type: "function",
//	        Function: types.FunctionDefinition{
//	            Name:        "get_weather",
//	            Description: "Get weather for a location",
//	            Parameters: types.ParametersDefinition{
//	                Type: "object",
//	                Properties: map[string]any{
//	                    "location": map[string]any{"type": "string"},
//	                },
//	                Required: []string{"location"},
//	            },
//	        },
//	    },
//	}
//
//	executables := []client.ToolExecutable{
//	    {
//	        Name: "get_weather",
//	        Execute: func(ctx context.Context, args string) (string, error) {
//	            // Parse args and return result
//	            return "Sunny, 22°C", nil
//	        },
//	    },
//	}
//
//	opts := &client.StreamRequestOptions{
//	    APIKey:          apiKey,
//	    Tools:           tools,
//	    ToolExecutables: executables,
//	}
//	ch := client.StreamRequest(ctx, req, "GPT-4o", opts)
//
// When tools are provided, the client automatically:
// 1. Sends the request with tool definitions
// 2. Aggregates tool call deltas from the streaming response
// 3. Executes the tool functions
// 4. Sends tool results back to the LLM
// 5. Streams the final response
//
// # File Upload
//
// Upload a file from disk:
//
//	file, _ := os.Open("document.pdf")
//	defer file.Close()
//	opts := &client.UploadFileOptions{
//	    File:     file,
//	    FileName: "document.pdf",
//	    APIKey:   apiKey,
//	}
//	attachment, err := client.UploadFile(ctx, opts)
//
// Upload a file by URL:
//
//	opts := &client.UploadFileOptions{
//	    FileURL:  "https://example.com/doc.pdf",
//	    FileName: "doc.pdf",
//	    APIKey:   apiKey,
//	}
//	attachment, err := client.UploadFile(ctx, opts)
//
// # Settings Sync
//
// Sync bot settings with the Poe API:
//
//	settings := map[string]any{
//	    "introduction_message": "Hello!",
//	    "server_bot_dependencies": map[string]int{"GPT-4": 1},
//	}
//	err := client.SyncBotSettings("mybot", "access-key", settings, "")
//
// Fetch settings (no update):
//
//	err := client.SyncBotSettings("mybot", "access-key", nil, "")
//
// # Error Handling
//
// The package defines three error types:
//
// BotError - retryable errors (network issues, temporary failures):
//
//	err := &client.BotError{Message: "timeout", Cause: originalErr}
//
// BotErrorNoRetry - permanent errors (bad request, invalid auth):
//
//	if client.IsBotErrorNoRetry(err) {
//	    log.Fatal("Permanent error:", err)
//	}
//
// AttachmentUploadError - file upload failures:
//
//	err := &client.AttachmentUploadError{Message: "upload failed"}
//
// # Retries and Timeouts
//
// Configure retries and timeouts via StreamRequestOptions:
//
//	opts := &client.StreamRequestOptions{
//	    NumTries:       3,                    // Retry up to 3 times
//	    RetrySleepTime: 1 * time.Second,      // Wait 1s between retries
//	    HTTPClient:     &http.Client{         // Custom client with timeout
//	        Timeout: 5 * time.Minute,
//	    },
//	}
//
// # SSE Event Types
//
// The client handles these Server-Sent Event types:
//   - text: Regular response text
//   - replace_response: Replace previous response
//   - suggested_reply: Suggested reply button
//   - file: File attachment
//   - json: Arbitrary JSON data
//   - meta: Metadata (linkify, content type, etc.)
//   - error: Error with optional retry flag
//   - done: End of stream
//   - ping: Keepalive (ignored)
//
// # Standard Library Only
//
// This package uses ONLY the Go standard library with no external dependencies
// (except sibling packages in this repo: types and sse).
package client
