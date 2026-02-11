package server

import (
	"context"

	"github.com/n0madic/go-poe/types"
)

// PoeBot defines the interface that bot developers implement
type PoeBot interface {
	// Path returns the URL path where this bot is served
	Path() string
	// AccessKey returns the access key for authentication
	AccessKey() string
	// BotName returns the name of the bot as it appears on Poe
	BotName() string
	// ShouldInsertAttachmentMessages returns whether to auto-parse attachments
	ShouldInsertAttachmentMessages() bool
	// GetResponse returns a channel of BotEvents in response to a query
	GetResponse(ctx context.Context, req *types.QueryRequest) <-chan types.BotEvent
	// GetSettings returns the bot's settings
	GetSettings(ctx context.Context, req *types.SettingsRequest) (*types.SettingsResponse, error)
	// OnFeedback is called when user provides feedback
	OnFeedback(ctx context.Context, req *types.ReportFeedbackRequest) error
	// OnReaction is called when user reacts to a message
	OnReaction(ctx context.Context, req *types.ReportReactionRequest) error
	// OnError is called when Poe server reports an error
	OnError(ctx context.Context, req *types.ReportErrorRequest) error
}

// BasePoeBot provides a default implementation of PoeBot that can be embedded
type BasePoeBot struct {
	path                           string
	accessKey                      string
	botName                        string
	shouldInsertAttachmentMessages bool
}

// NewBasePoeBot creates a new BasePoeBot with the given configuration
func NewBasePoeBot(path, accessKey, botName string) *BasePoeBot {
	return &BasePoeBot{
		path:                           path,
		accessKey:                      accessKey,
		botName:                        botName,
		shouldInsertAttachmentMessages: true,
	}
}

func (b *BasePoeBot) Path() string                         { return b.path }
func (b *BasePoeBot) AccessKey() string                    { return b.accessKey }
func (b *BasePoeBot) BotName() string                      { return b.botName }
func (b *BasePoeBot) ShouldInsertAttachmentMessages() bool { return b.shouldInsertAttachmentMessages }

// SetAccessKey sets the access key (used during app setup)
func (b *BasePoeBot) SetAccessKey(key string) { b.accessKey = key }

// SetBotName sets the bot name (used during app setup)
func (b *BasePoeBot) SetBotName(name string) { b.botName = name }

// GetResponse default implementation yields "hello"
func (b *BasePoeBot) GetResponse(ctx context.Context, req *types.QueryRequest) <-chan types.BotEvent {
	ch := make(chan types.BotEvent, 1)
	go func() {
		defer close(ch)
		ch <- &types.PartialResponse{Text: "hello"}
	}()
	return ch
}

// GetSettings default returns a SettingsResponse with default version=2
func (b *BasePoeBot) GetSettings(ctx context.Context, req *types.SettingsRequest) (*types.SettingsResponse, error) {
	return types.NewSettingsResponse(), nil
}

// OnFeedback default is a no-op
func (b *BasePoeBot) OnFeedback(ctx context.Context, req *types.ReportFeedbackRequest) error {
	return nil
}

// OnReaction default is a no-op
func (b *BasePoeBot) OnReaction(ctx context.Context, req *types.ReportReactionRequest) error {
	return nil
}

// OnError default logs the error (just returns nil here)
func (b *BasePoeBot) OnError(ctx context.Context, req *types.ReportErrorRequest) error {
	return nil
}
