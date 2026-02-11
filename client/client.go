package client

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/n0madic/go-poe/types"
)

const (
	defaultBaseURL       = "https://api.poe.com/bot/"
	defaultNumTries      = 2
	defaultRetrySleep    = 500 * time.Millisecond
	defaultClientTimeout = 600 * time.Second
)

// ToolExecutable represents a tool function that can be called
type ToolExecutable struct {
	Name    string
	Execute func(ctx context.Context, args string) (string, error)
}

// StreamRequestOptions configures a stream request
type StreamRequestOptions struct {
	APIKey          string
	Tools           []types.ToolDefinition
	ToolExecutables []ToolExecutable
	NumTries        int
	RetrySleepTime  time.Duration
	BaseURL         string
	ExtraHeaders    map[string]string
	HTTPClient      *http.Client
}

func (o *StreamRequestOptions) defaults() {
	if o.NumTries <= 0 {
		o.NumTries = defaultNumTries
	}
	if o.RetrySleepTime <= 0 {
		o.RetrySleepTime = defaultRetrySleep
	}
	if o.BaseURL == "" {
		o.BaseURL = defaultBaseURL
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: defaultClientTimeout}
	}
}

func (o *StreamRequestOptions) headers() map[string]string {
	headers := make(map[string]string)
	if o.APIKey != "" {
		headers["Authorization"] = "Bearer " + o.APIKey
	}
	for k, v := range o.ExtraHeaders {
		headers[k] = v
	}
	return headers
}

// StreamRequest is the main entry point for calling other Poe bots.
// If Tools are provided, it uses the tools path.
func StreamRequest(ctx context.Context, req *types.QueryRequest, botName string, opts *StreamRequestOptions) <-chan *types.PartialResponse {
	ch := make(chan *types.PartialResponse, 64)
	if opts == nil {
		opts = &StreamRequestOptions{}
	}
	opts.defaults()

	go func() {
		defer close(ch)
		if len(opts.Tools) > 0 {
			streamRequestWithTools(ctx, req, botName, opts, ch)
		} else {
			streamRequestBase(ctx, req, botName, opts, ch)
		}
	}()
	return ch
}

// streamRequestBase handles retries and calls performQueryRequest
func streamRequestBase(ctx context.Context, req *types.QueryRequest, botName string, opts *StreamRequestOptions, ch chan<- *types.PartialResponse) {
	url := strings.TrimRight(opts.BaseURL, "/") + "/" + botName
	headers := opts.headers()

	payload := buildPayload(req, nil, nil, nil)

	for i := 0; i < opts.NumTries; i++ {
		err := performQueryRequest(ctx, opts.HTTPClient, url, payload, headers, ch)
		if err == nil {
			return
		}

		if IsBotErrorNoRetry(err) {
			log.Printf("Bot request to %s failed (no retry): %v", botName, err)
			return
		}

		log.Printf("Bot request to %s failed on try %d: %v", botName, i, err)

		if i == opts.NumTries-1 {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(opts.RetrySleepTime):
		}
	}
}

// streamRequestBaseWithPayload handles retries with a custom payload
func streamRequestBaseWithPayload(ctx context.Context, botName string, opts *StreamRequestOptions, payload map[string]any, ch chan<- *types.PartialResponse) {
	url := strings.TrimRight(opts.BaseURL, "/") + "/" + botName
	headers := opts.headers()

	for i := 0; i < opts.NumTries; i++ {
		err := performQueryRequest(ctx, opts.HTTPClient, url, payload, headers, ch)
		if err == nil {
			return
		}

		if IsBotErrorNoRetry(err) {
			log.Printf("Bot request to %s failed (no retry): %v", botName, err)
			return
		}

		log.Printf("Bot request to %s failed on try %d: %v", botName, i, err)

		if i == opts.NumTries-1 {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(opts.RetrySleepTime):
		}
	}
}

func buildPayload(req *types.QueryRequest, tools []types.ToolDefinition, toolCalls []types.ToolCallDefinition, toolResults []types.ToolResultDefinition) map[string]any {
	// Marshal the request to get a map
	data, _ := json.Marshal(req)
	var payload map[string]any
	json.Unmarshal(data, &payload)

	if tools != nil {
		payload["tools"] = tools
	}
	if toolCalls != nil {
		payload["tool_calls"] = toolCalls
	}
	if toolResults != nil {
		payload["tool_results"] = toolResults
	}
	return payload
}

// GetBotResponse constructs a QueryRequest and calls StreamRequest
func GetBotResponse(ctx context.Context, messages []types.ProtocolMessage, botName, apiKey string, opts *StreamRequestOptions) <-chan *types.PartialResponse {
	if opts == nil {
		opts = &StreamRequestOptions{}
	}
	opts.APIKey = apiKey

	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:          messages,
		UserID:         "",
		ConversationID: "",
		MessageID:      "",
	}

	return StreamRequest(ctx, req, botName, opts)
}

// GetFinalResponse collects the full response text
func GetFinalResponse(ctx context.Context, req *types.QueryRequest, botName, apiKey string, opts *StreamRequestOptions) (string, error) {
	if opts == nil {
		opts = &StreamRequestOptions{}
	}
	if apiKey != "" {
		opts.APIKey = apiKey
	}

	ch := StreamRequest(ctx, req, botName, opts)
	var chunks []string

	for msg := range ch {
		// Skip meta responses
		if msg.RawResponse != nil {
			if _, ok := msg.RawResponse.(*types.MetaResponse); ok {
				continue
			}
		}
		if msg.IsSuggestedReply {
			continue
		}
		if msg.IsReplaceResponse {
			chunks = nil
		}
		chunks = append(chunks, msg.Text)
	}

	if len(chunks) == 0 {
		return "", &BotError{Message: "Bot " + botName + " sent no response"}
	}
	return strings.Join(chunks, ""), nil
}
