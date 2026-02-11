package client

import (
	"context"
	"encoding/json"
	"log"

	"github.com/n0madic/go-poe/types"
)

// streamRequestWithTools handles the two-pass tool execution flow
func streamRequestWithTools(ctx context.Context, req *types.QueryRequest, botName string, opts *StreamRequestOptions, ch chan<- *types.PartialResponse) {
	// First pass: collect tool call deltas
	firstPassCh := make(chan *types.PartialResponse, 64)
	aggregatedToolCalls := make(map[int]*types.ToolCallDefinition)

	payload := buildPayload(req, opts.Tools, nil, nil)

	go func() {
		defer close(firstPassCh)
		streamRequestBaseWithPayload(ctx, botName, opts, payload, firstPassCh)
	}()

	for msg := range firstPassCh {
		if msg.Data == nil || msg.Data["choices"] == nil {
			ch <- msg
			continue
		}

		choices, ok := msg.Data["choices"].([]any)
		if !ok || len(choices) == 0 {
			ch <- msg
			continue
		}

		choice, ok := choices[0].(map[string]any)
		if !ok {
			ch <- msg
			continue
		}

		// Check finish reason
		if choice["finish_reason"] != nil {
			continue
		}

		delta, ok := choice["delta"].(map[string]any)
		if !ok {
			ch <- msg
			continue
		}

		if toolCallsRaw, ok := delta["tool_calls"]; ok {
			toolCallsList, ok := toolCallsRaw.([]any)
			if !ok {
				continue
			}

			// Parse tool call deltas
			var deltas []types.ToolCallDefinitionDelta
			for _, tcRaw := range toolCallsList {
				tcBytes, _ := json.Marshal(tcRaw)
				var delta types.ToolCallDefinitionDelta
				if json.Unmarshal(tcBytes, &delta) == nil {
					deltas = append(deltas, delta)
				}
			}

			// If no executables, yield raw deltas
			if len(opts.ToolExecutables) == 0 {
				ch <- &types.PartialResponse{
					Text:      "",
					ToolCalls: deltas,
					Index:     msg.Index,
				}
				continue
			}

			// Aggregate tool calls
			for _, delta := range deltas {
				if _, exists := aggregatedToolCalls[delta.Index]; !exists {
					if delta.ID == nil || delta.Type == nil || delta.Function.Name == nil {
						continue
					}
					aggregatedToolCalls[delta.Index] = &types.ToolCallDefinition{
						ID:   *delta.ID,
						Type: *delta.Type,
						Function: types.FunctionCallDefinition{
							Name:      *delta.Function.Name,
							Arguments: delta.Function.Arguments,
						},
					}
				} else {
					aggregatedToolCalls[delta.Index].Function.Arguments += delta.Function.Arguments
				}
			}
		} else if content, ok := delta["content"]; ok {
			if contentStr, ok := content.(string); ok {
				ch <- &types.PartialResponse{Text: contentStr, Index: msg.Index}
			}
		}
	}

	// If no tool executables, exit early
	if len(opts.ToolExecutables) == 0 {
		return
	}

	// Execute tools
	toolCalls := make([]types.ToolCallDefinition, 0, len(aggregatedToolCalls))
	for _, tc := range aggregatedToolCalls {
		toolCalls = append(toolCalls, *tc)
	}

	if len(toolCalls) == 0 {
		return
	}

	toolResults, err := executeTools(ctx, opts.ToolExecutables, toolCalls)
	if err != nil {
		log.Printf("Error executing tools: %v", err)
		return
	}

	// Second pass: send tool results back to LLM
	secondPayload := buildPayload(req, opts.Tools, toolCalls, toolResults)
	streamRequestBaseWithPayload(ctx, botName, opts, secondPayload, ch)
}

// executeTools runs tool functions and collects results
func executeTools(ctx context.Context, executables []ToolExecutable, toolCalls []types.ToolCallDefinition) ([]types.ToolResultDefinition, error) {
	execMap := make(map[string]ToolExecutable)
	for _, exec := range executables {
		execMap[exec.Name] = exec
	}

	var results []types.ToolResultDefinition
	for _, tc := range toolCalls {
		exec, ok := execMap[tc.Function.Name]
		if !ok {
			log.Printf("Tool executable not found: %s", tc.Function.Name)
			continue
		}

		content, err := exec.Execute(ctx, tc.Function.Arguments)
		if err != nil {
			log.Printf("Tool execution error for %s: %v", tc.Function.Name, err)
			content = err.Error()
		}

		results = append(results, types.ToolResultDefinition{
			Role:       "tool",
			ToolCallID: tc.ID,
			Name:       tc.Function.Name,
			Content:    content,
		})
	}
	return results, nil
}
