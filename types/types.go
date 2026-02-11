package types

// Type aliases
type Identifier = string
type FeedbackType = string
type ContentType = string
type MessageType = string
type ErrorType = string
type RequestType = string
type Number = float64

// FeedbackType constants
const (
	FeedbackLike    FeedbackType = "like"
	FeedbackDislike FeedbackType = "dislike"
)

// ContentType constants
const (
	ContentTypeMarkdown ContentType = "text/markdown"
	ContentTypePlain    ContentType = "text/plain"
)

// MessageType constants
const (
	MessageTypeFunctionCall MessageType = "function_call"
)

// ErrorType constants
const (
	ErrorUserMessageTooLong        ErrorType = "user_message_too_long"
	ErrorInsufficientFund          ErrorType = "insufficient_fund"
	ErrorUserCausedError           ErrorType = "user_caused_error"
	ErrorPrivacyAuthorizationError ErrorType = "privacy_authorization_error"
)

// RequestType constants
const (
	RequestTypeQuery          RequestType = "query"
	RequestTypeSettings       RequestType = "settings"
	RequestTypeReportFeedback RequestType = "report_feedback"
	RequestTypeReportReaction RequestType = "report_reaction"
	RequestTypeReportError    RequestType = "report_error"
)

// ProtocolVersion is the current protocol version
const ProtocolVersion = "1.2"
