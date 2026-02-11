package client

import "fmt"

// BotError is raised when there is an error communicating with the bot
type BotError struct {
	Message string
	Cause   error
}

func (e *BotError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *BotError) Unwrap() error { return e.Cause }

// BotErrorNoRetry is a BotError that should not be retried
type BotErrorNoRetry struct {
	BotError
}

// IsBotErrorNoRetry checks if the error is a BotErrorNoRetry
func IsBotErrorNoRetry(err error) bool {
	_, ok := err.(*BotErrorNoRetry)
	return ok
}

// AttachmentUploadError is raised when there is an error uploading an attachment
type AttachmentUploadError struct {
	Message string
}

func (e *AttachmentUploadError) Error() string { return e.Message }
