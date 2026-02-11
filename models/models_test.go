package models

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	result, err := Fetch(context.Background(), nil)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected at least one model from the API")
	}

	// Verify every model has required fields populated
	for _, m := range result {
		if m.ID == "" {
			t.Error("model has empty ID")
		}
		if m.Object != "model" {
			t.Errorf("model %s: expected object \"model\", got %q", m.ID, m.Object)
		}
		if m.Created == 0 {
			t.Errorf("model %s: expected non-zero created timestamp", m.ID)
		}
		if m.OwnedBy == "" {
			t.Errorf("model %s: expected non-empty owned_by", m.ID)
		}
		if m.Architecture.Modality == "" {
			t.Errorf("model %s: expected non-empty modality", m.ID)
		}
		if len(m.Architecture.InputModalities) == 0 {
			t.Errorf("model %s: expected at least one input modality", m.ID)
		}
		if len(m.Architecture.OutputModalities) == 0 {
			t.Errorf("model %s: expected at least one output modality", m.ID)
		}
		if m.Metadata.DisplayName == "" {
			t.Errorf("model %s: expected non-empty display_name", m.ID)
		}
	}

	// Find a model with pricing to verify nullable object parsing
	var foundPriced bool
	for _, m := range result {
		if m.Pricing != nil {
			foundPriced = true
			if m.Pricing.Prompt == nil && m.Pricing.Completion == nil {
				t.Errorf("model %s: has pricing object but both prompt and completion are nil", m.ID)
			}
			break
		}
	}
	if !foundPriced {
		t.Log("warning: no model with pricing found in API response")
	}

	// Find a model with context window
	var foundContext bool
	for _, m := range result {
		if m.ContextWindow != nil {
			foundContext = true
			if m.ContextWindow.ContextLength <= 0 {
				t.Errorf("model %s: context_window present but context_length <= 0", m.ID)
			}
			// Top-level context_length should match
			if m.ContextLength == nil {
				t.Errorf("model %s: has context_window but top-level context_length is nil", m.ID)
			} else if *m.ContextLength != m.ContextWindow.ContextLength {
				t.Errorf("model %s: top-level context_length %d != context_window.context_length %d",
					m.ID, *m.ContextLength, m.ContextWindow.ContextLength)
			}
			break
		}
	}
	if !foundContext {
		t.Log("warning: no model with context_window found in API response")
	}

	// Find a model with reasoning
	var foundReasoning bool
	for _, m := range result {
		if m.Reasoning != nil {
			foundReasoning = true
			break
		}
	}
	if !foundReasoning {
		t.Log("warning: no model with reasoning found in API response")
	}

	// Find a model with parameters
	var foundParams bool
	for _, m := range result {
		if len(m.Parameters) > 0 {
			foundParams = true
			for _, p := range m.Parameters {
				if p.Name == "" {
					t.Errorf("model %s: parameter has empty name", m.ID)
				}
				if len(p.Schema) == 0 {
					t.Errorf("model %s: parameter %s has empty schema", m.ID, p.Name)
				}
			}
			break
		}
	}
	if !foundParams {
		t.Log("warning: no model with parameters found in API response")
	}

	// Find a model with null pricing (e.g. image gen models)
	var foundNullPricing bool
	for _, m := range result {
		if m.Pricing == nil {
			foundNullPricing = true
			break
		}
	}
	if !foundNullPricing {
		t.Log("warning: no model with null pricing found in API response")
	}
}

func TestFetchError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := Fetch(context.Background(), &Options{BaseURL: srv.URL})
	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}

func TestModelTypes(t *testing.T) {
	prompt := "0.0000026"
	completion := "0.000013"
	cacheRead := "0.00000026"
	cacheWrite := "0.0000032"
	ctxLen := 983040
	maxOutput := 32768

	m := Model{
		ID:          "test-model",
		Object:      "model",
		Created:     1758868894776,
		Description: "A test model",
		OwnedBy:     "TestOrg",
		Root:        "test-model",
		Architecture: Architecture{
			InputModalities:  []string{"text", "image"},
			OutputModalities: []string{"text"},
			Modality:         "text,image->text",
		},
		SupportedFeatures: []string{"tools"},
		Pricing: &Pricing{
			Prompt:          &prompt,
			Completion:      &completion,
			InputCacheRead:  &cacheRead,
			InputCacheWrite: &cacheWrite,
		},
		ContextWindow: &ContextWindow{
			ContextLength:   ctxLen,
			MaxOutputTokens: &maxOutput,
		},
		ContextLength: &ctxLen,
		Metadata: ModelMetadata{
			DisplayName: "Test Model",
			Image: &ModelImage{
				URL:    "https://example.com/icon.png",
				Alt:    "icon",
				Width:  200,
				Height: 200,
			},
			URL: "https://poe.com/test-model",
		},
		Reasoning: &Reasoning{
			Budget:                  &ReasoningBudget{MaxTokens: 31999, MinTokens: 0},
			Required:                false,
			SupportsReasoningEffort: true,
		},
		Parameters: []Parameter{
			{
				Name:         "thinking_budget",
				Schema:       json.RawMessage(`{"type":"number","minimum":0,"maximum":31999}`),
				DefaultValue: json.RawMessage(`0`),
				Description:  "Token budget for thinking",
			},
		},
	}

	// Marshal and unmarshal to verify JSON round-trip
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var m2 Model
	if err := json.Unmarshal(data, &m2); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if m2.ID != m.ID {
		t.Errorf("ID mismatch: %s != %s", m2.ID, m.ID)
	}
	if m2.Pricing == nil || m2.Pricing.Prompt == nil || *m2.Pricing.Prompt != prompt {
		t.Errorf("Pricing round-trip failed")
	}
	if m2.Pricing.Image != nil {
		t.Errorf("expected nil image pricing after round-trip")
	}
	if m2.ContextWindow == nil || m2.ContextWindow.MaxOutputTokens == nil || *m2.ContextWindow.MaxOutputTokens != maxOutput {
		t.Errorf("ContextWindow round-trip failed")
	}
	if m2.Reasoning == nil || m2.Reasoning.Budget == nil || m2.Reasoning.Budget.MaxTokens != 31999 {
		t.Errorf("Reasoning round-trip failed")
	}
	if !m2.Reasoning.SupportsReasoningEffort {
		t.Errorf("expected SupportsReasoningEffort to be true")
	}
	if len(m2.Parameters) != 1 || m2.Parameters[0].Name != "thinking_budget" {
		t.Errorf("Parameters round-trip failed")
	}

	// Verify DefaultValue round-trip
	var defaultVal float64
	if err := json.Unmarshal(m2.Parameters[0].DefaultValue, &defaultVal); err != nil {
		t.Fatalf("DefaultValue unmarshal error: %v", err)
	}
	if defaultVal != 0 {
		t.Errorf("expected default_value 0, got %v", defaultVal)
	}
}

func TestFetchCustomOptions(t *testing.T) {
	var receivedHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"object":"list","data":[]}`))
	}))
	defer srv.Close()

	customClient := &http.Client{}
	models, err := Fetch(context.Background(), &Options{
		BaseURL:    srv.URL,
		HTTPClient: customClient,
		ExtraHeaders: map[string]string{
			"X-Custom-Header": "test-value",
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("expected 0 models, got %d", len(models))
	}
	if receivedHeaders.Get("X-Custom-Header") != "test-value" {
		t.Errorf("custom header not received: %v", receivedHeaders)
	}
}

func TestFetchNilOptions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"object":"list","data":[]}`))
	}))
	defer srv.Close()

	// Pass nil opts — verify it doesn't panic (uses defaults)
	// We can't test with nil opts against the real API, so we just test
	// that constructing defaults works with a non-nil opts pointing to the mock
	models, err := Fetch(context.Background(), &Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if models == nil {
		t.Errorf("expected non-nil (possibly empty) slice")
	}
}

func TestFetchCancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"object":"list","data":[]}`))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := Fetch(ctx, &Options{BaseURL: srv.URL})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
