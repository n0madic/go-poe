# Server Package

The `server` package provides a Go implementation of the Poe bot protocol server. It allows you to create custom bots that can be hosted and integrated with the Poe platform.

## Features

- **PoeBot Interface**: Define custom bot behavior by implementing the `PoeBot` interface
- **BasePoeBot**: Convenient base implementation with sensible defaults
- **HTTP Server**: Built-in HTTP server with authentication and request routing
- **SSE Streaming**: Server-Sent Events streaming for real-time bot responses
- **Attachment Processing**: Automatic parsing and insertion of attachment messages
- **Multi-bot Support**: Host multiple bots on a single server
- **Cost API**: Support for monetized bots with cost authorization and capture
- **Standard Library Only**: No external dependencies beyond the go-poe packages

## Quick Start

### Creating a Simple Bot

```go
package main

import (
    "context"
    "github.com/n0madic/go-poe/server"
    "github.com/n0madic/go-poe/types"
)

type EchoBot struct {
    *server.BasePoeBot
}

func NewEchoBot() *EchoBot {
    return &EchoBot{
        BasePoeBot: server.NewBasePoeBot("/", "", "EchoBot"),
    }
}

func (b *EchoBot) GetResponse(ctx context.Context, req *types.QueryRequest) <-chan types.BotEvent {
    ch := make(chan types.BotEvent, 1)
    go func() {
        defer close(ch)
        lastMsg := req.Query[len(req.Query)-1]
        ch <- &types.PartialResponse{Text: "You said: " + lastMsg.Content}
    }()
    return ch
}

func main() {
    bot := NewEchoBot()
    server.Run(bot)
}
```

### Environment Variables

- `POE_ACCESS_KEY` — Your bot's access key (from the bot's edit page on Poe)
- `POE_BOT_NAME` — Your bot's name (must match exactly as shown on Poe)

### Running the Server

```bash
# Default port 8080
go run main.go

# Custom port
go run main.go -port 3000
```

### Deployment

The bot server must be accessible from the internet over **HTTPS on port 443**.
Poe's servers only connect to standard HTTPS port — non-standard ports (8443, 3000, etc.)
will not work.

A typical production setup uses a reverse proxy (nginx, caddy) to handle TLS
and forward traffic to the Go server:

```nginx
server {
    listen 443 ssl;
    server_name bot.example.com;

    ssl_certificate     /etc/letsencrypt/live/bot.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/bot.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Connection '';
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding off;
    }
}
```

Key nginx settings for SSE streaming: `proxy_buffering off`, `proxy_cache off`,
and `Connection ''`.

Set the bot server URL in the Poe dashboard to `https://bot.example.com/`.

## PoeBot Interface

The `PoeBot` interface defines the methods a bot must implement:

```go
type PoeBot interface {
    Path() string
    AccessKey() string
    BotName() string
    ShouldInsertAttachmentMessages() bool
    GetResponse(ctx context.Context, req *types.QueryRequest) <-chan types.BotEvent
    GetSettings(ctx context.Context, req *types.SettingsRequest) (*types.SettingsResponse, error)
    OnFeedback(ctx context.Context, req *types.ReportFeedbackRequest) error
    OnReaction(ctx context.Context, req *types.ReportReactionRequest) error
    OnError(ctx context.Context, req *types.ReportErrorRequest) error
}
```

### Key Methods

- **GetResponse**: Returns a channel of `BotEvent` items (PartialResponse, ErrorResponse, MetaResponse, DataResponse) that are streamed to the client as SSE events
- **GetSettings**: Returns bot configuration like introduction message, attachment support, etc.
- **OnFeedback/OnReaction/OnError**: Handle user feedback, reactions, and error reports

## BasePoeBot

The `BasePoeBot` struct provides default implementations for all `PoeBot` methods:

```go
bot := server.NewBasePoeBot("/path", "access_key", "bot_name")
```

You can embed `BasePoeBot` in your custom bot and override only the methods you need.

## Response Types

### PartialResponse

Stream text responses to the user:

```go
ch <- &types.PartialResponse{Text: "Hello!"}
ch <- &types.PartialResponse{Text: " World"}
```

### MetaResponse

Control response metadata:

```go
meta := types.NewMetaResponse()
meta.ContentType = types.ContentTypeMarkdown
meta.RefetchSettings = true
ch <- meta
```

### ErrorResponse

Send error messages:

```go
err := types.NewErrorResponse("Something went wrong")
err.AllowRetry = true
ch <- err
```

### DataResponse

Attach arbitrary metadata:

```go
ch <- &types.DataResponse{Metadata: `{"key": "value"}`}
```

## Attachment Handling

By default, `ShouldInsertAttachmentMessages()` returns `true`, which automatically processes attachments and inserts their content as separate messages before the user's message.

Supported attachment types:
- **Text files** (`text/plain`, etc.): Content is inserted with a text template
- **HTML** (`text/html`): Treated as web content with URL template
- **Images** (`image/*`): Vision descriptions are inserted with image template
- **PDFs** (`application/pdf`): Extracted text is inserted

To disable automatic attachment processing:

```go
type MyBot struct {
    *server.BasePoeBot
}

func (b *MyBot) ShouldInsertAttachmentMessages() bool {
    return false
}
```

## Multi-Bot Hosting

Host multiple bots on different paths:

```go
bot1 := NewBot1("/bot1", "key1", "Bot1")
bot2 := NewBot2("/bot2", "key2", "Bot2")

app := server.MakeApp(bot1, bot2)
http.ListenAndServe(":8080", app)
```

## Authentication

Set the access key in three ways (in order of priority):

1. Pass directly to `NewBasePoeBot(path, accessKey, botName)`
2. Use `server.FindAccessKey("")` which checks the `POE_ACCESS_KEY` environment variable
3. Leave empty for no authentication (development only)

```go
accessKey := server.FindAccessKey("")  // Checks POE_ACCESS_KEY env
bot := server.NewBasePoeBot("/", accessKey, "MyBot")
```

## Settings Sync

When a bot has both `BotName()` and `AccessKey()` set, the server automatically syncs the bot's settings with the Poe API on startup. This ensures your bot's configuration on Poe matches your code.

## Cost API (Monetization)

For monetized bots, use the cost API to authorize or capture charges:

```go
import "github.com/n0madic/go-poe/server"
import "github.com/n0madic/go-poe/types"

// Authorize a cost before processing
amounts := []types.CostItem{
    {AmountUSDMilliCents: 1000},  // $0.01
}
err := server.AuthorizeCost(ctx, accessKey, req.BotQueryID, amounts, "")

// Capture cost after processing
err = server.CaptureCost(ctx, accessKey, req.BotQueryID, amounts, "")
```

### Error Handling

```go
if err != nil {
    if _, ok := err.(*server.InsufficientFundError); ok {
        // User doesn't have enough funds
    } else if costErr, ok := err.(*server.CostRequestError); ok {
        // Other cost API error
        log.Printf("Cost error: %s", costErr.Message)
    }
}
```

## Utility Functions

### MakePromptAuthorRoleAlternated

Merge consecutive messages with the same role and deduplicate attachments:

```go
messages := []types.ProtocolMessage{
    {Role: "user", Content: "First"},
    {Role: "user", Content: "Second"},
}
merged := server.MakePromptAuthorRoleAlternated(messages)
// Result: [{Role: "user", Content: "First\n\nSecond"}]
```

## Testing

The package includes comprehensive tests covering:

- HTTP handler behavior (GET, POST, authentication)
- SSE streaming
- Attachment processing (text, HTML, images, PDFs)
- Message merging and deduplication
- Multi-bot hosting
- Error handling

Run tests:

```bash
go test ./server/... -v
```

## Example Bots

- **[echo_bot](../examples/echo_bot)**: Simple bot that echoes user messages back

See also the [client examples](../client/README.md#examples) for querying bots from Go code.

## License

Part of the go-poe project. See repository root for license information.
