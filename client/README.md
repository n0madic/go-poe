# Client Package

The `client` package provides a Go client for the Poe bot protocol. It allows you to query other Poe bots via the Bot Query API with support for SSE streaming, tool calling, and file uploads.

## Features

- **SSE Streaming**: Stream bot responses in real-time using Server-Sent Events
- **Tool Calling**: OpenAI-compatible function calling with automatic execution
- **File Upload**: Upload files to Poe with retry logic
- **Settings Sync**: Synchronize bot settings with the Poe API
- **Retry Logic**: Automatic retries with configurable sleep time
- **Context Support**: Full context.Context support for cancellation

## Installation

```bash
go get github.com/n0madic/go-poe
```

## Usage

### Basic Query

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/n0madic/go-poe/client"
    "github.com/n0madic/go-poe/types"
)

func main() {
    apiKey := "your-poe-api-key"

    messages := []types.ProtocolMessage{
        {Role: "user", Content: "What is the capital of France?"},
    }

    ctx := context.Background()
    ch := client.GetBotResponse(ctx, messages, "GPT-4o", apiKey, nil)

    for response := range ch {
        if response.Text != "" {
            fmt.Print(response.Text)
        }
    }
}
```

### Get Final Response

```go
req := &types.QueryRequest{
    BaseRequest: types.BaseRequest{
        Version: types.ProtocolVersion,
        Type:    types.RequestTypeQuery,
    },
    Query:          messages,
    UserID:         "user-123",
    ConversationID: "conv-123",
    MessageID:      "msg-123",
}

opts := &client.StreamRequestOptions{
    APIKey: apiKey,
}

finalResponse, err := client.GetFinalResponse(ctx, req, "GPT-4o", "", opts)
if err != nil {
    log.Fatal(err)
}

fmt.Println(finalResponse)
```

### Tool Calling

```go
// Define a tool
tools := []types.ToolDefinition{
    {
        Type: "function",
        Function: types.FunctionDefinition{
            Name:        "get_weather",
            Description: "Get current weather",
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

// Define the executable function
executables := []client.ToolExecutable{
    {
        Name: "get_weather",
        Execute: func(ctx context.Context, args string) (string, error) {
            var params struct {
                Location string `json:"location"`
            }
            json.Unmarshal([]byte(args), &params)
            return fmt.Sprintf("Weather in %s: Sunny", params.Location), nil
        },
    },
}

// Query with tools
opts := &client.StreamRequestOptions{
    APIKey:          apiKey,
    Tools:           tools,
    ToolExecutables: executables,
}

ch := client.StreamRequest(ctx, req, "GPT-4o", opts)
for response := range ch {
    fmt.Print(response.Text)
}
```

### Upload File

```go
file, _ := os.Open("document.pdf")
defer file.Close()

opts := &client.UploadFileOptions{
    File:     file,
    FileName: "document.pdf",
    APIKey:   apiKey,
}

attachment, err := client.UploadFile(ctx, opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Uploaded: %s (%s)\n", attachment.Name, attachment.URL)
```

### Upload File by URL

```go
opts := &client.UploadFileOptions{
    FileURL:  "https://example.com/document.pdf",
    FileName: "document.pdf",
    APIKey:   apiKey,
}

attachment, err := client.UploadFile(ctx, opts)
```

### Sync Bot Settings

```go
settings := map[string]any{
    "introduction_message": "Hello! I'm a helpful bot.",
    "server_bot_dependencies": map[string]int{
        "GPT-4": 1,
    },
}

err := client.SyncBotSettings("mybot", "access-key", settings, "")
if err != nil {
    log.Fatal(err)
}
```

## Configuration

### StreamRequestOptions

```go
type StreamRequestOptions struct {
    APIKey          string                    // Poe API key (Bearer token)
    Tools           []types.ToolDefinition    // Tools for function calling
    ToolExecutables []ToolExecutable          // Executable functions
    NumTries        int                       // Number of retry attempts (default: 2)
    RetrySleepTime  time.Duration            // Sleep between retries (default: 500ms)
    BaseURL         string                    // API base URL (default: https://api.poe.com/bot/)
    ExtraHeaders    map[string]string        // Additional HTTP headers
    HTTPClient      *http.Client             // Custom HTTP client
}
```

### UploadFileOptions

```go
type UploadFileOptions struct {
    File           io.Reader         // File reader (for file upload)
    FileURL        string            // File URL (for URL upload)
    FileName       string            // Name of the file
    APIKey         string            // Poe API key (raw, not Bearer)
    NumTries       int               // Number of retry attempts
    RetrySleepTime time.Duration     // Sleep between retries
    BaseURL        string            // API base URL
    ExtraHeaders   map[string]string // Additional HTTP headers
    HTTPClient     *http.Client      // Custom HTTP client
}
```

## Response Types

### PartialResponse

The main response type yielded during streaming:

```go
type PartialResponse struct {
    Text              string                    // Response text
    Data              map[string]any            // Arbitrary data (json event)
    RawResponse       any                       // MetaResponse, etc.
    FullPrompt        *string                   // Full prompt used
    RequestID         *string                   // Request ID
    IsSuggestedReply  bool                      // Is this a suggested reply
    IsReplaceResponse bool                      // Replace previous response
    Attachment        *Attachment               // File attachment
    ToolCalls         []ToolCallDefinitionDelta // Tool call deltas
    Index             *int                      // Response index
}
```

## Error Handling

### BotError

Standard error that can be retried:

```go
err := &client.BotError{
    Message: "Connection failed",
    Cause:   originalError,
}
```

### BotErrorNoRetry

Error that should not be retried (e.g., bad request):

```go
if client.IsBotErrorNoRetry(err) {
    log.Fatal("Permanent error:", err)
}
```

### AttachmentUploadError

Specific error for file upload failures:

```go
err := &client.AttachmentUploadError{
    Message: "Upload failed",
}
```

## Examples

See the `/examples` directory for complete working examples:

- `query_bot` - Basic bot querying with streaming
- `tool_bot` - Tool calling with custom functions

## Testing

Run the test suite:

```bash
go test ./client/...
```

Run with verbose output:

```bash
go test -v ./client/...
```

## Standard Library Only

This package uses ONLY the Go standard library with no external dependencies (except the sibling packages in this repo):

- `github.com/n0madic/go-poe/types` - Protocol types
- `github.com/n0madic/go-poe/sse` - SSE reader

## Protocol Version

Current protocol version: **1.2**

The client automatically uses the correct protocol version defined in `types.ProtocolVersion`.

## License

See the main repository LICENSE file.
