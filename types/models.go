package types

import "encoding/json"

// BaseRequest contains common data for all requests
type BaseRequest struct {
	Version string      `json:"version"`
	Type    RequestType `json:"type"`
}

// MessageFeedback represents feedback for a message
type MessageFeedback struct {
	Type   FeedbackType `json:"type"`
	Reason *string      `json:"reason,omitempty"`
}

// Sender of a message
type Sender struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// User in a chat
type User struct {
	ID   Identifier `json:"id"`
	Name *string    `json:"name,omitempty"`
}

// MessageReaction represents a reaction to a message
type MessageReaction struct {
	UserID   Identifier `json:"user_id"`
	Reaction string     `json:"reaction"`
}

// ProtocolMessage is a message in the Poe protocol
type ProtocolMessage struct {
	Role              string            `json:"role"` // "system", "user", "bot", "tool"
	MessageType       *string           `json:"message_type,omitempty"`
	SenderID          *string           `json:"sender_id,omitempty"`
	Sender            *Sender           `json:"sender,omitempty"`
	Content           string            `json:"content"`
	Parameters        map[string]any    `json:"parameters,omitempty"`
	ContentType       ContentType       `json:"content_type,omitempty"`
	Timestamp         int64             `json:"timestamp,omitempty"`
	MessageID         string            `json:"message_id,omitempty"`
	Feedback          []MessageFeedback `json:"feedback,omitempty"`
	Attachments       []Attachment      `json:"attachments,omitempty"`
	Metadata          *string           `json:"metadata,omitempty"`
	ReferencedMessage *ProtocolMessage  `json:"referenced_message,omitempty"`
	Reactions         []MessageReaction `json:"reactions,omitempty"`
}

// QueryRequest is the request for a query
type QueryRequest struct {
	BaseRequest
	Query               []ProtocolMessage      `json:"query"`
	UserID              Identifier             `json:"user_id"`
	ConversationID      Identifier             `json:"conversation_id"`
	MessageID           Identifier             `json:"message_id"`
	Metadata            string                 `json:"metadata,omitempty"`
	AccessKey           string                 `json:"access_key,omitempty"`
	Temperature         *float64               `json:"temperature,omitempty"`
	SkipSystemPrompt    bool                   `json:"skip_system_prompt,omitempty"`
	LogitBias           map[string]float64     `json:"logit_bias,omitempty"`
	StopSequences       []string               `json:"stop_sequences,omitempty"`
	LanguageCode        string                 `json:"language_code,omitempty"`
	AdoptCurrentBotName *bool                  `json:"adopt_current_bot_name,omitempty"`
	BotQueryID          Identifier             `json:"bot_query_id,omitempty"`
	Users               []User                 `json:"users,omitempty"`
	Tools               []ToolDefinition       `json:"tools,omitempty"`
	ToolCalls           []ToolCallDefinition   `json:"tool_calls,omitempty"`
	ToolResults         []ToolResultDefinition `json:"tool_results,omitempty"`
	QueryCreationTime   *int64                 `json:"query_creation_time,omitempty"`
	ExtraParams         map[string]any         `json:"extra_params,omitempty"`
}

// SettingsRequest is the request for settings
type SettingsRequest struct {
	BaseRequest
}

// ReportFeedbackRequest is the request for reporting feedback
type ReportFeedbackRequest struct {
	BaseRequest
	MessageID      Identifier   `json:"message_id"`
	UserID         Identifier   `json:"user_id"`
	ConversationID Identifier   `json:"conversation_id"`
	FeedbackType   FeedbackType `json:"feedback_type"`
}

// ReportReactionRequest is the request for reporting reactions
type ReportReactionRequest struct {
	BaseRequest
	MessageID      Identifier `json:"message_id"`
	UserID         Identifier `json:"user_id"`
	ConversationID Identifier `json:"conversation_id"`
	Reaction       string     `json:"reaction"`
}

// ReportErrorRequest is the request for reporting errors
type ReportErrorRequest struct {
	BaseRequest
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata"`
}

// ParseRawRequest parses a raw JSON request and returns the type field
func ParseRawRequest(data []byte) (RequestType, json.RawMessage, error) {
	var base BaseRequest
	if err := json.Unmarshal(data, &base); err != nil {
		return "", nil, err
	}
	return base.Type, json.RawMessage(data), nil
}
