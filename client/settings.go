package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/n0madic/go-poe/types"
)

// SyncBotSettings syncs bot settings with the Poe API
func SyncBotSettings(botName, accessKey string, settings map[string]any, baseURL string) error {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	var syncURL string
	var body io.Reader
	var contentType string

	escapedName := url.PathEscape(botName)
	escapedKey := url.PathEscape(accessKey)
	if settings != nil {
		syncURL = fmt.Sprintf("%supdate_settings/%s/%s/%s", baseURL, escapedName, escapedKey, types.ProtocolVersion)
		data, err := json.Marshal(settings)
		if err != nil {
			return &BotError{Message: fmt.Sprintf("failed to marshal settings: %v", err)}
		}
		body = bytes.NewReader(data)
		contentType = "application/json"
	} else {
		syncURL = fmt.Sprintf("%sfetch_settings/%s/%s/%s", baseURL, escapedName, escapedKey, types.ProtocolVersion)
		body = http.NoBody
		contentType = ""
	}

	req, err := http.NewRequest(http.MethodPost, syncURL, body)
	if err != nil {
		return &BotError{Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		msg := fmt.Sprintf("timeout syncing settings for bot %s", botName)
		if settings == nil {
			msg += ". Check that the bot server is running."
		}
		return &BotError{Message: msg, Cause: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return &BotError{Message: fmt.Sprintf("error syncing settings for bot %s: %s", botName, string(respBody))}
	}

	return nil
}
