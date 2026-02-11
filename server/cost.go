package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/n0madic/go-poe/sse"
	"github.com/n0madic/go-poe/types"
)

// CostRequestError is returned when a cost request fails
type CostRequestError struct {
	Message string
}

func (e *CostRequestError) Error() string { return e.Message }

// InsufficientFundError is returned when the user doesn't have enough funds
type InsufficientFundError struct{}

func (e *InsufficientFundError) Error() string { return "insufficient funds" }

// CaptureCost captures variable costs for monetized bot creators
func CaptureCost(ctx context.Context, accessKey, botQueryID string, amounts []types.CostItem, baseURL string) error {
	if baseURL == "" {
		baseURL = "https://api.poe.com/"
	}
	url := fmt.Sprintf("%sbot/cost/%s/capture", baseURL, botQueryID)
	return costRequestInner(ctx, accessKey, url, amounts)
}

// AuthorizeCost authorizes a cost for monetized bot creators
func AuthorizeCost(ctx context.Context, accessKey, botQueryID string, amounts []types.CostItem, baseURL string) error {
	if baseURL == "" {
		baseURL = "https://api.poe.com/"
	}
	url := fmt.Sprintf("%sbot/cost/%s/authorize", baseURL, botQueryID)
	return costRequestInner(ctx, accessKey, url, amounts)
}

func costRequestInner(ctx context.Context, accessKey, url string, amounts []types.CostItem) error {
	data := map[string]any{
		"amounts":    amounts,
		"access_key": accessKey,
	}
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cost request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create cost request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return &CostRequestError{Message: fmt.Sprintf("HTTP error during cost request: %v", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return &CostRequestError{
			Message: fmt.Sprintf("%d %s: %s", resp.StatusCode, resp.Status, string(respBody)),
		}
	}

	reader := sse.NewReader(resp.Body)
	for {
		event, err := reader.ReadEvent()
		if err != nil {
			break
		}
		if event.Event == "result" {
			var eventData map[string]any
			if err := json.Unmarshal([]byte(event.Data), &eventData); err == nil {
				if status, ok := eventData["status"].(string); ok && status == "success" {
					return nil
				}
			}
		}
	}

	return &InsufficientFundError{}
}
