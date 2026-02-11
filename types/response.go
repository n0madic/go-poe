package types

// BotEvent is a marker interface for types that can be yielded from GetResponse
type BotEvent interface {
	isBotEvent()
}

// PartialResponse is the primary response type yielded during streaming
type PartialResponse struct {
	Text              string                    `json:"text"`
	Data              map[string]any            `json:"data,omitempty"`
	RawResponse       any                       `json:"raw_response,omitempty"`
	FullPrompt        *string                   `json:"full_prompt,omitempty"`
	RequestID         *string                   `json:"request_id,omitempty"`
	IsSuggestedReply  bool                      `json:"is_suggested_reply,omitempty"`
	IsReplaceResponse bool                      `json:"is_replace_response,omitempty"`
	Attachment        *Attachment               `json:"attachment,omitempty"`
	ToolCalls         []ToolCallDefinitionDelta `json:"tool_calls,omitempty"`
	Index             *int                      `json:"index,omitempty"`
}

func (r *PartialResponse) isBotEvent() {}

// ErrorResponse is similar to PartialResponse for communicating errors
type ErrorResponse struct {
	PartialResponse
	AllowRetry bool    `json:"allow_retry"`
	ErrorType  *string `json:"error_type,omitempty"`
}

func (r *ErrorResponse) isBotEvent() {}

// NewErrorResponse creates an ErrorResponse with default AllowRetry=true
func NewErrorResponse(text string) *ErrorResponse {
	return &ErrorResponse{
		PartialResponse: PartialResponse{Text: text},
		AllowRetry:      true,
	}
}

// MetaResponse carries meta event information
type MetaResponse struct {
	PartialResponse
	Linkify          bool        `json:"linkify"`
	SuggestedReplies bool        `json:"suggested_replies"`
	ContentType      ContentType `json:"content_type"`
	RefetchSettings  bool        `json:"refetch_settings,omitempty"`
}

func (r *MetaResponse) isBotEvent() {}

// NewMetaResponse creates a MetaResponse with sensible defaults
func NewMetaResponse() *MetaResponse {
	return &MetaResponse{
		Linkify:          true,
		SuggestedReplies: true,
		ContentType:      ContentTypeMarkdown,
	}
}

// DataResponse contains arbitrary data to attach to the bot response
type DataResponse struct {
	Metadata string `json:"metadata"`
}

func (r *DataResponse) isBotEvent() {}

// SettingsResponse is the bot's response to a settings request
type SettingsResponse struct {
	ResponseVersion              *int               `json:"response_version,omitempty"`
	ContextClearWindowSecs       *int               `json:"context_clear_window_secs,omitempty"`
	AllowUserContextClear        *bool              `json:"allow_user_context_clear,omitempty"`
	CustomRateCard               *string            `json:"custom_rate_card,omitempty"`
	ServerBotDependencies        map[string]int     `json:"server_bot_dependencies,omitempty"`
	AllowAttachments             *bool              `json:"allow_attachments,omitempty"`
	IntroductionMessage          *string            `json:"introduction_message,omitempty"`
	ExpandTextAttachments        *bool              `json:"expand_text_attachments,omitempty"`
	EnableImageComprehension     *bool              `json:"enable_image_comprehension,omitempty"`
	EnforceAuthorRoleAlternation *bool              `json:"enforce_author_role_alternation,omitempty"`
	EnableMultiBotChatPrompting  *bool              `json:"enable_multi_bot_chat_prompting,omitempty"`
	EnableMultiEntityPrompting   *bool              `json:"enable_multi_entity_prompting,omitempty"`
	RateCard                     *string            `json:"rate_card,omitempty"`
	CostLabel                    *string            `json:"cost_label,omitempty"`
	ParameterControls            *ParameterControls `json:"parameter_controls,omitempty"`
}

// NewSettingsResponse creates a SettingsResponse with default version=2
func NewSettingsResponse() *SettingsResponse {
	v := 2
	return &SettingsResponse{
		ResponseVersion: &v,
	}
}
