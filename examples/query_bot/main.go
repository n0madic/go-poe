package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/n0madic/go-poe/client"
	"github.com/n0madic/go-poe/types"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("POE_API_KEY")
	if apiKey == "" {
		log.Fatal("POE_API_KEY environment variable is required")
	}

	// Create a simple query
	messages := []types.ProtocolMessage{
		{
			Role:    "user",
			Content: "What is the capital of France?",
		},
	}

	// Stream the response
	fmt.Println("Streaming response from GPT-4o:")
	ctx := context.Background()

	ch := client.GetBotResponse(ctx, messages, "GPT-4o", apiKey, nil)

	for response := range ch {
		// Skip meta responses
		if response.RawResponse != nil {
			if meta, ok := response.RawResponse.(*types.MetaResponse); ok {
				fmt.Printf("[Meta: content_type=%s, linkify=%v]\n",
					meta.ContentType, meta.Linkify)
				continue
			}
		}

		// Skip suggested replies in streaming
		if response.IsSuggestedReply {
			fmt.Printf("\n[Suggested: %s]\n", response.Text)
			continue
		}

		// Handle replace response
		if response.IsReplaceResponse {
			fmt.Print("\r[Replace] ")
		}

		// Print the text
		if response.Text != "" {
			fmt.Print(response.Text)
		}

		// Handle attachments
		if response.Attachment != nil {
			fmt.Printf("\n[Attachment: %s (%s)]\n",
				response.Attachment.Name,
				response.Attachment.URL)
		}
	}

	fmt.Println("\n\n--- Using GetFinalResponse ---")

	// Or get the final response directly
	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query: messages,
	}

	finalResponse, err := client.GetFinalResponse(ctx, req, "GPT-4o", apiKey, nil)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("Final response: %s\n", finalResponse)
}
