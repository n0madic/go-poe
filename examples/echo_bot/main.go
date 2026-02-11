package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/n0madic/go-poe/server"
	"github.com/n0madic/go-poe/types"
)

// EchoBot echoes back the user's message
type EchoBot struct {
	*server.BasePoeBot
}

func NewEchoBot() *EchoBot {
	accessKey := server.FindAccessKey("")
	botName := os.Getenv("POE_BOT_NAME")
	if botName == "" {
		log.Fatal("POE_BOT_NAME environment variable is required (must match your bot name on Poe)")
	}
	return &EchoBot{
		BasePoeBot: server.NewBasePoeBot("/", accessKey, botName),
	}
}

// GetResponse echoes the last message from the user
func (b *EchoBot) GetResponse(ctx context.Context, req *types.QueryRequest) <-chan types.BotEvent {
	ch := make(chan types.BotEvent, 2)

	go func() {
		defer close(ch)

		// Get the last message
		if len(req.Query) == 0 {
			ch <- &types.PartialResponse{Text: "No message received"}
			return
		}

		lastMessage := req.Query[len(req.Query)-1]
		response := fmt.Sprintf("You said: %s", lastMessage.Content)

		// Send the response
		ch <- &types.PartialResponse{Text: response}
	}()

	return ch
}

// GetSettings returns bot settings
func (b *EchoBot) GetSettings(ctx context.Context, req *types.SettingsRequest) (*types.SettingsResponse, error) {
	settings := types.NewSettingsResponse()
	intro := "Hello! I'm EchoBot. I will echo back whatever you say to me."
	settings.IntroductionMessage = &intro
	return settings, nil
}

func main() {
	bot := NewEchoBot()
	server.Run(bot)
}
