# types

Go types package for the Poe bot protocol. This is a port of the Python `fastapi_poe` types to idiomatic Go.

## Features

- Complete type definitions for the Poe protocol v1.2
- Request types: `QueryRequest`, `SettingsRequest`, `ReportFeedbackRequest`, etc.
- Response types: `PartialResponse`, `ErrorResponse`, `MetaResponse`, `SettingsResponse`
- Tool calling support: `ToolDefinition`, `ToolCallDefinition`, `ToolResultDefinition`
- UI parameter controls: `TextField`, `DropDown`, `Slider`, `AspectRatio`, etc.
- Discriminated union types for flexible control structures
- Custom JSON unmarshaling for `CostItem` with ceiling behavior for floats
- Message attachment types with templates for LLM consumption

## Usage

```go
import "github.com/n0madic/go-poe/types"

// Create a query request
req := types.QueryRequest{
    BaseRequest: types.BaseRequest{
        Version: types.ProtocolVersion,
        Type:    types.RequestTypeQuery,
    },
    Query: []types.ProtocolMessage{
        {
            Role:    "user",
            Content: "Hello, bot!",
        },
    },
    UserID:         "user123",
    ConversationID: "conv456",
    MessageID:      "msg789",
}

// Create a partial response
resp := types.PartialResponse{
    Text: "Hello! How can I help you?",
}

// Create settings response with defaults
settings := types.NewSettingsResponse()
settings.IntroductionMessage = ptr("Welcome to my bot!")

// Create error response with retry enabled
errResp := types.NewErrorResponse("Something went wrong")
```

## Type Aliases

The package provides type aliases for common types:
- `Identifier` = `string`
- `FeedbackType` = `string`
- `ContentType` = `string`
- `MessageType` = `string`
- `ErrorType` = `string`
- `RequestType` = `string`
- `Number` = `float64`

## Constants

Protocol version and common values:
```go
types.ProtocolVersion           // "1.2"
types.FeedbackLike              // "like"
types.ContentTypeMarkdown       // "text/markdown"
types.RequestTypeQuery          // "query"
types.ErrorUserMessageTooLong   // "user_message_too_long"
```

## Discriminated Unions

The package provides two discriminated union types for UI controls:

- `BaseControl`: Union of basic control types (Divider, TextField, TextArea, DropDown, ToggleSwitch, Slider, AspectRatio)
- `FullControl`: All BaseControl types plus ConditionallyRenderControls

Use `NewBaseControl()` and `NewFullControl()` to wrap concrete types, and `.Underlying()` to retrieve them.

## Templates

String templates for formatting attachment content:
```go
types.TextAttachmentTemplate        // For text files
types.URLAttachmentTemplate         // For URL content
types.ImageVisionAttachmentTemplate // For image analysis
```

Use with `fmt.Sprintf()` to format attachment messages.

## Testing

Run tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover
```

## Standards

- Uses only Go standard library
- Idiomatic Go naming and style
- JSON tags for all serializable fields
- Pointer fields for optional values
- Comprehensive test coverage
