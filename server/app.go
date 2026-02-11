package server

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/n0madic/go-poe/types"
)

// FindAccessKey checks param, then POE_ACCESS_KEY env
func FindAccessKey(accessKey string) string {
	if accessKey != "" {
		return strings.TrimSpace(accessKey)
	}
	if envKey := os.Getenv("POE_ACCESS_KEY"); envKey != "" {
		return strings.TrimSpace(envKey)
	}
	return ""
}

// syncBotSettings syncs bot settings with the Poe API
func syncBotSettings(botName, accessKey string, settings map[string]any, baseURL string) error {
	if baseURL == "" {
		baseURL = "https://api.poe.com/bot/"
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
			return fmt.Errorf("failed to marshal settings: %w", err)
		}
		body = bytes.NewReader(data)
		contentType = "application/json"
	} else {
		syncURL = fmt.Sprintf("%sfetch_settings/%s/%s/%s", baseURL, escapedName, escapedKey, types.ProtocolVersion)
		body = nil
		contentType = ""
	}

	req, err := http.NewRequest(http.MethodPost, syncURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("timeout syncing settings for bot %s: %w", botName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error syncing settings for bot %s: %s", botName, string(respBody))
	}
	return nil
}

// MakeApp creates an http.Handler that serves one or more PoeBot instances
func MakeApp(bots ...PoeBot) http.Handler {
	mux := http.NewServeMux()

	// Validate unique paths
	paths := make(map[string]bool)
	for _, bot := range bots {
		if paths[bot.Path()] {
			panic(fmt.Sprintf("Multiple bots are trying to use the same path: %s", bot.Path()))
		}
		paths[bot.Path()] = true
	}

	for _, bot := range bots {
		handler := botHandler(bot)
		mux.Handle(bot.Path(), handler)

		// Sync settings on startup if bot has name and access key
		if bot.BotName() != "" && bot.AccessKey() != "" {
			go func(b PoeBot) {
				settings, err := b.GetSettings(context.Background(), &types.SettingsRequest{
					BaseRequest: types.BaseRequest{
						Version: types.ProtocolVersion,
						Type:    types.RequestTypeSettings,
					},
				})
				if err != nil {
					log.Printf("Error getting settings for %s: %v", b.BotName(), err)
					return
				}
				settingsMap := make(map[string]any)
				data, _ := json.Marshal(settings)
				json.Unmarshal(data, &settingsMap)
				if err := syncBotSettings(b.BotName(), b.AccessKey(), settingsMap, ""); err != nil {
					log.Printf("Error syncing settings for %s: %v", b.BotName(), err)
				}
			}(bot)
		} else {
			log.Printf("Warning: Bot name or access key not set. Settings will NOT be synced automatically.")
		}
	}

	return mux
}

// Run creates the app and starts an HTTP server
func Run(bots ...PoeBot) {
	port := flag.Int("p", 8080, "port to listen on")
	flag.IntVar(port, "port", 8080, "port to listen on")
	flag.Parse()

	handler := MakeApp(bots...)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting Poe bot server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
