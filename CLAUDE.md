# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`go-poe` is a Go SDK for the [Poe bot protocol](https://creator.poe.com/docs/), a port of Python's `fastapi_poe`. It provides both a **server** (host bots on Poe) and a **client** (call other Poe bots). Zero external dependencies — uses only the Go standard library. Protocol version: **1.2**.

## Build & Test Commands

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./server/...
go test ./client/...
go test ./types/...
go test ./sse/...

# Run a single test
go test ./server/... -run TestHandlerQuery

# Verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Run benchmarks (SSE package)
go test -bench=. ./sse/...

# Format code
gofmt -w .

# Run an example bot
POE_ACCESS_KEY=<key> POE_BOT_NAME=<name> go run ./examples/echo_bot/
```

## Architecture

### Package Structure

Five packages with a strict dependency hierarchy (no cycles):

```
types  ←── sse  ←── client
  ↑
  └──── server ────────┘ (server uses sse + types; client uses sse + types)

models  (standalone — no internal deps, only standard library)
```

- **`types/`** — Protocol types, constants, and the `BotEvent` interface. All request/response structs, tool definitions, attachment types, UI parameter controls (discriminated unions via `BaseControl`/`FullControl`), and JSON (un)marshaling.
- **`sse/`** — Minimal SSE implementation: `Reader` (parses SSE streams from `io.Reader`), `Writer` (writes SSE events to `http.ResponseWriter` with flush), and `Event` struct.
- **`server/`** — Bot hosting framework. `PoeBot` interface + `BasePoeBot` default implementation. `MakeApp()` creates an `http.Handler` for one or more bots. Handles auth, request routing by type, SSE streaming of `BotEvent` channels, attachment processing, message merging, cost API, and settings sync on startup.
- **`client/`** — Bot Query API client. `StreamRequest()` returns `<-chan *types.PartialResponse`. Supports SSE streaming, retry logic, OpenAI-compatible tool calling (two-pass: aggregate deltas → execute → send results), file upload (multipart + URL modes), and `SyncBotSettings()`.
- **`models/`** — Model catalog client. `Fetch()` retrieves available Poe models from the public API (`https://api.poe.com/v1/models`). Returns structured types with pricing, context window, architecture, reasoning config, and parameters. No authentication required.

### Key Patterns

- **Channel-based streaming**: Both server and client use `<-chan` for streaming. Server bots return `<-chan types.BotEvent` from `GetResponse()`. Client returns `<-chan *types.PartialResponse` from `StreamRequest()`.
- **Interface embedding**: Custom bots embed `*server.BasePoeBot` and override only needed methods (typically `GetResponse` and `GetSettings`).
- **BotEvent discriminated union**: `types.BotEvent` is a marker interface (`isBotEvent()`) implemented by `PartialResponse`, `ErrorResponse`, `MetaResponse`, and `DataResponse`. Server's `handleQuery` uses type switches to emit the correct SSE event type.
- **Pointer fields for optional values**: Optional JSON fields use pointer types (`*string`, `*bool`, `*int`) throughout the types package.
- **Controls discriminated unions**: `BaseControl` and `FullControl` wrap concrete UI control types. Use `NewBaseControl()`/`NewFullControl()` to wrap, `.Underlying()` to unwrap. Custom JSON marshaling in `controls_json.go`.

### Environment Variables

- `POE_ACCESS_KEY` — Bot access key (checked by `server.FindAccessKey()`)
- `POE_BOT_NAME` — Bot name as shown on Poe (used in examples)

### Server Request Flow

1. HTTP POST → `botHandler` → auth check → `ParseRawRequest` (extracts `type` field)
2. Switch on request type → unmarshal into specific request struct
3. For `query`: optionally `InsertAttachmentMessages()` → call `bot.GetResponse()` → consume `BotEvent` channel → write SSE events via `sse.Writer` → emit `done`

### Client Tool Calling Flow (two-pass)

1. First pass: send request with tool definitions → collect and aggregate `ToolCallDefinitionDelta` chunks by index
2. Execute tool functions via `ToolExecutable` map
3. Second pass: send original request + tool calls + tool results → stream final response
