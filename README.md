# go-poe

Go SDK for the [Poe bot protocol](https://creator.poe.com/docs/) — a port of the Python [`fastapi_poe`](https://github.com/poe-platform/fastapi_poe) library to idiomatic Go.

Build custom bots for the [Poe](https://poe.com) platform and call other Poe bots via the Bot Query API. Zero external dependencies — uses only the Go standard library.

## Features

- **Server** — host custom bots on Poe with SSE streaming, attachment handling, multi-bot support, and cost API
- **Client** — query other Poe bots with SSE streaming, tool calling, file uploads, and retry logic
- **Models** — fetch the Poe model catalog with pricing, context window, reasoning config, and parameters
- **Types** — complete type definitions for the Poe protocol v1.2
- **SSE** — lightweight Server-Sent Events reader and writer

## Installation

```bash
go get github.com/n0madic/go-poe
```

## Quick Start

### Create a bot (server)

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
    bot := &EchoBot{
        BasePoeBot: server.NewBasePoeBot("/", server.FindAccessKey(""), "EchoBot"),
    }
    server.Run(bot)
}
```

```bash
POE_ACCESS_KEY=<key> go run main.go
```

### Query a bot (client)

```go
package main

import (
    "context"
    "fmt"
    "github.com/n0madic/go-poe/client"
    "github.com/n0madic/go-poe/types"
)

func main() {
    messages := []types.ProtocolMessage{
        {Role: "user", Content: "What is the capital of France?"},
    }

    ch := client.GetBotResponse(context.Background(), messages, "GPT-4o", "your-api-key", nil)
    for response := range ch {
        fmt.Print(response.Text)
    }
}
```

## Packages

| Package | Description | Docs |
|---------|-------------|------|
| [`types`](./types) | Protocol types, constants, and JSON serialization | [README](./types/README.md) |
| [`sse`](./sse) | SSE reader and writer | — |
| [`server`](./server) | Bot hosting framework with HTTP server | [README](./server/README.md) |
| [`client`](./client) | Bot Query API client for calling other bots | [README](./client/README.md) |
| [`models`](./models) | Poe model catalog (pricing, context, reasoning) | [README](./models/README.md) |

## Examples

| Example | Description | Run |
|---------|-------------|-----|
| [`echo_bot`](./examples/echo_bot) | Server bot that echoes user messages | `POE_ACCESS_KEY=<key> POE_BOT_NAME=<name> go run ./examples/echo_bot/` |
| [`query_bot`](./examples/query_bot) | Client querying a bot with streaming | `POE_API_KEY=<key> go run ./examples/query_bot/` |
| [`tool_bot`](./examples/tool_bot) | Client with OpenAI-compatible tool calling | `POE_API_KEY=<key> go run ./examples/tool_bot/` |

## Testing

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Benchmarks (SSE package)
go test -bench=. ./sse/...
```
