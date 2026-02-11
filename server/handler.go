package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/n0madic/go-poe/types"
)

// authenticate checks the Authorization: Bearer <key> header
func authenticate(r *http.Request, accessKey string) bool {
	if accessKey == "" {
		return true
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return false
	}
	return strings.TrimPrefix(auth, "Bearer ") == accessKey
}

// botHandler creates an http.Handler for a single bot
func botHandler(bot PoeBot) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request to %s", r.Method, r.URL.Path)

		if r.Method == http.MethodGet {
			handleIndex(w, r)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !authenticate(r, bot.AccessKey()) {
			log.Printf("Authentication failed for request to %s", r.URL.Path)
			http.Error(w, `{"detail":"Invalid access key"}`, http.StatusUnauthorized)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		reqType, rawMsg, err := types.ParseRawRequest(body)
		if err != nil {
			log.Printf("Invalid JSON in request: %v", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Processing request type: %s", reqType)

		ctx := r.Context()

		switch reqType {
		case types.RequestTypeQuery:
			var req types.QueryRequest
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				http.Error(w, "Invalid query request", http.StatusBadRequest)
				return
			}
			if bot.AccessKey() != "" {
				req.AccessKey = bot.AccessKey()
			}
			handleQuery(ctx, w, bot, &req)

		case types.RequestTypeSettings:
			var req types.SettingsRequest
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				http.Error(w, "Invalid settings request", http.StatusBadRequest)
				return
			}
			handleSettings(ctx, w, bot, &req)

		case types.RequestTypeReportFeedback:
			var req types.ReportFeedbackRequest
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				http.Error(w, "Invalid feedback request", http.StatusBadRequest)
				return
			}
			if err := bot.OnFeedback(ctx, &req); err != nil {
				log.Printf("Error handling feedback: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{}"))

		case types.RequestTypeReportReaction:
			var req types.ReportReactionRequest
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				http.Error(w, "Invalid reaction request", http.StatusBadRequest)
				return
			}
			if err := bot.OnReaction(ctx, &req); err != nil {
				log.Printf("Error handling reaction: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{}"))

		case types.RequestTypeReportError:
			var req types.ReportErrorRequest
			if err := json.Unmarshal(rawMsg, &req); err != nil {
				http.Error(w, "Invalid error request", http.StatusBadRequest)
				return
			}
			if err := bot.OnError(ctx, &req); err != nil {
				log.Printf("Error handling error report: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{}"))

		default:
			http.Error(w, "Unsupported request type", http.StatusNotImplemented)
		}
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	url := "https://poe.com/create_bot?server=1"
	w.Write([]byte(
		`<html><body><h1>Go Poe bot server</h1><p>Congratulations! Your server` +
			` is running. To connect it to Poe, create a bot at <a` +
			` href="` + url + `">` + url + `</a>.</p></body></html>`,
	))
}

func handleSettings(ctx context.Context, w http.ResponseWriter, bot PoeBot, req *types.SettingsRequest) {
	settings, err := bot.GetSettings(ctx, req)
	if err != nil {
		log.Printf("Error getting settings: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}
