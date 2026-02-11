package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.poe.com/v1/models"
	defaultTimeout = 30 * time.Second
)

// Options configures the Fetch request.
type Options struct {
	// BaseURL overrides the default API endpoint.
	BaseURL string
	// HTTPClient overrides the default HTTP client.
	HTTPClient *http.Client
	// ExtraHeaders are added to the request.
	ExtraHeaders map[string]string
}

func (o *Options) defaults() {
	if o.BaseURL == "" {
		o.BaseURL = defaultBaseURL
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: defaultTimeout}
	}
}

// Fetch retrieves the list of available models from the Poe API.
func Fetch(ctx context.Context, opts *Options) ([]Model, error) {
	if opts == nil {
		opts = &Options{}
	}
	opts.defaults()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, opts.BaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("models: create request: %w", err)
	}

	for k, v := range opts.ExtraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := opts.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("models: fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models: unexpected status %d", resp.StatusCode)
	}

	var result ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("models: decode response: %w", err)
	}

	return result.Data, nil
}
