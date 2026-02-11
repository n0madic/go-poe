package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/n0madic/go-poe/client"
	"github.com/n0madic/go-poe/types"
)

// getWeather is a mock function that returns weather for a location
func getWeather(ctx context.Context, args string) (string, error) {
	var params struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	// Mock weather data
	weather := map[string]string{
		"Paris":     "Sunny, 22°C",
		"London":    "Cloudy, 15°C",
		"New York":  "Rainy, 18°C",
		"Tokyo":     "Clear, 25°C",
		"Singapore": "Hot and humid, 32°C",
	}

	if w, ok := weather[params.Location]; ok {
		return fmt.Sprintf("The weather in %s is %s", params.Location, w), nil
	}

	return fmt.Sprintf("Weather data for %s is not available", params.Location), nil
}

func main() {
	apiKey := os.Getenv("POE_API_KEY")
	if apiKey == "" {
		log.Fatal("POE_API_KEY environment variable is required")
	}

	// Define the tool
	tools := []types.ToolDefinition{
		{
			Type: "function",
			Function: types.FunctionDefinition{
				Name:        "get_weather",
				Description: "Get the current weather for a location",
				Parameters: types.ParametersDefinition{
					Type: "object",
					Properties: map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "The city name",
						},
					},
					Required: []string{"location"},
				},
			},
		},
	}

	// Define the executable
	executables := []client.ToolExecutable{
		{
			Name:    "get_weather",
			Execute: getWeather,
		},
	}

	// Create the query
	messages := []types.ProtocolMessage{
		{
			Role:    "user",
			Content: "What's the weather like in Paris and London?",
		},
	}

	// Configure with tools
	opts := &client.StreamRequestOptions{
		APIKey:          apiKey,
		Tools:           tools,
		ToolExecutables: executables,
	}

	ctx := context.Background()
	fmt.Println("Querying bot with tool support...")

	ch := client.GetBotResponse(ctx, messages, "GPT-4o", apiKey, opts)

	for response := range ch {
		if response.Text != "" {
			fmt.Print(response.Text)
		}

		// Tool calls (if no executables were provided, you'd see these)
		if len(response.ToolCalls) > 0 {
			for _, tc := range response.ToolCalls {
				fmt.Printf("\n[Tool call delta: index=%d", tc.Index)
				if tc.ID != nil {
					fmt.Printf(", id=%s", *tc.ID)
				}
				if tc.Function.Name != nil {
					fmt.Printf(", name=%s", *tc.Function.Name)
				}
				if tc.Function.Arguments != "" {
					fmt.Printf(", args=%s", tc.Function.Arguments)
				}
				fmt.Println("]")
			}
		}
	}

	fmt.Println("\n\nDone!")
}
