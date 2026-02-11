// Package models provides access to the Poe model catalog API.
// It fetches available models from https://api.poe.com/v1/models
// and returns structured Go types with model properties including
// pricing, context window, architecture, reasoning config, and parameters.
package models

import "encoding/json"

// ModelsResponse is the top-level response from the models API.
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Model represents a single model in the Poe catalog.
type Model struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int64          `json:"created"`
	Description       string         `json:"description"`
	OwnedBy           string         `json:"owned_by"`
	Root              string         `json:"root"`
	Architecture      Architecture   `json:"architecture"`
	SupportedFeatures []string       `json:"supported_features"`
	Pricing           *Pricing       `json:"pricing"`
	ContextWindow     *ContextWindow `json:"context_window"`
	ContextLength     *int           `json:"context_length"`
	Metadata          ModelMetadata  `json:"metadata"`
	Reasoning         *Reasoning     `json:"reasoning"`
	Parameters        []Parameter    `json:"parameters"`
}

// Architecture describes the model's input/output modalities.
type Architecture struct {
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Modality         string   `json:"modality"`
}

// Pricing contains per-token pricing as decimal strings.
// Null values indicate the pricing component is not applicable.
type Pricing struct {
	Prompt          *string `json:"prompt"`
	Completion      *string `json:"completion"`
	Image           *string `json:"image"`
	Request         *string `json:"request"`
	InputCacheRead  *string `json:"input_cache_read"`
	InputCacheWrite *string `json:"input_cache_write"`
}

// ContextWindow describes the model's context and output token limits.
type ContextWindow struct {
	ContextLength   int  `json:"context_length"`
	MaxOutputTokens *int `json:"max_output_tokens"`
}

// ModelMetadata contains display information for the model.
type ModelMetadata struct {
	DisplayName string      `json:"display_name"`
	Image       *ModelImage `json:"image"`
	URL         string      `json:"url"`
}

// ModelImage represents the model's icon or avatar.
type ModelImage struct {
	URL    string `json:"url"`
	Alt    string `json:"alt"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Reasoning describes the model's reasoning/thinking capabilities.
type Reasoning struct {
	Budget                  *ReasoningBudget `json:"budget"`
	Required                bool             `json:"required"`
	SupportsReasoningEffort bool             `json:"supports_reasoning_effort"`
}

// ReasoningBudget defines token limits for the reasoning budget.
type ReasoningBudget struct {
	MaxTokens int `json:"max_tokens"`
	MinTokens int `json:"min_tokens"`
}

// Parameter describes a configurable parameter for the model.
type Parameter struct {
	Name         string          `json:"name"`
	Schema       json.RawMessage `json:"schema"`
	DefaultValue json.RawMessage `json:"default_value"`
	Description  string          `json:"description,omitempty"`
}
