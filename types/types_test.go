package types

import (
	"encoding/json"
	"testing"
)

// TestQueryRequestRoundTrip tests JSON marshaling and unmarshaling of QueryRequest
func TestQueryRequestRoundTrip(t *testing.T) {
	temp := 0.7
	qct := int64(1234567890)

	original := QueryRequest{
		BaseRequest: BaseRequest{
			Version: ProtocolVersion,
			Type:    RequestTypeQuery,
		},
		Query: []ProtocolMessage{
			{
				Role:    "user",
				Content: "Hello, bot!",
			},
		},
		UserID:            "user123",
		ConversationID:    "conv456",
		MessageID:         "msg789",
		Temperature:       &temp,
		QueryCreationTime: &qct,
		ExtraParams: map[string]any{
			"custom_field": "custom_value",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal QueryRequest: %v", err)
	}

	var decoded QueryRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal QueryRequest: %v", err)
	}

	if decoded.UserID != original.UserID {
		t.Errorf("UserID mismatch: got %q, want %q", decoded.UserID, original.UserID)
	}
	if decoded.ConversationID != original.ConversationID {
		t.Errorf("ConversationID mismatch: got %q, want %q", decoded.ConversationID, original.ConversationID)
	}
	if decoded.MessageID != original.MessageID {
		t.Errorf("MessageID mismatch: got %q, want %q", decoded.MessageID, original.MessageID)
	}
	if decoded.Temperature == nil || *decoded.Temperature != *original.Temperature {
		t.Errorf("Temperature mismatch: got %v, want %v", decoded.Temperature, original.Temperature)
	}
	if decoded.QueryCreationTime == nil || *decoded.QueryCreationTime != *original.QueryCreationTime {
		t.Errorf("QueryCreationTime mismatch: got %v, want %v", decoded.QueryCreationTime, original.QueryCreationTime)
	}
	if decoded.ExtraParams["custom_field"] != "custom_value" {
		t.Errorf("ExtraParams mismatch: got %v, want %v", decoded.ExtraParams, original.ExtraParams)
	}
}

// TestPartialResponseRoundTrip tests JSON marshaling and unmarshaling of PartialResponse
func TestPartialResponseRoundTrip(t *testing.T) {
	idx := 0
	original := PartialResponse{
		Text:              "This is a partial response",
		IsSuggestedReply:  true,
		IsReplaceResponse: false,
		Index:             &idx,
		Data: map[string]any{
			"key": "value",
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal PartialResponse: %v", err)
	}

	var decoded PartialResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal PartialResponse: %v", err)
	}

	if decoded.Text != original.Text {
		t.Errorf("Text mismatch: got %q, want %q", decoded.Text, original.Text)
	}
	if decoded.IsSuggestedReply != original.IsSuggestedReply {
		t.Errorf("IsSuggestedReply mismatch: got %v, want %v", decoded.IsSuggestedReply, original.IsSuggestedReply)
	}
	if decoded.Index == nil || *decoded.Index != *original.Index {
		t.Errorf("Index mismatch: got %v, want %v", decoded.Index, original.Index)
	}
	if decoded.Data["key"] != "value" {
		t.Errorf("Data mismatch: got %v, want %v", decoded.Data, original.Data)
	}
}

// TestSettingsResponseRoundTrip tests JSON marshaling and unmarshaling of SettingsResponse
func TestSettingsResponseRoundTrip(t *testing.T) {
	introMsg := "Welcome to the bot!"
	allowAttach := true
	original := SettingsResponse{
		ResponseVersion:     ptr(2),
		IntroductionMessage: &introMsg,
		AllowAttachments:    &allowAttach,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal SettingsResponse: %v", err)
	}

	var decoded SettingsResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal SettingsResponse: %v", err)
	}

	if decoded.ResponseVersion == nil || *decoded.ResponseVersion != 2 {
		t.Errorf("ResponseVersion mismatch: got %v, want 2", decoded.ResponseVersion)
	}
	if decoded.IntroductionMessage == nil || *decoded.IntroductionMessage != introMsg {
		t.Errorf("IntroductionMessage mismatch: got %v, want %q", decoded.IntroductionMessage, introMsg)
	}
	if decoded.AllowAttachments == nil || *decoded.AllowAttachments != allowAttach {
		t.Errorf("AllowAttachments mismatch: got %v, want %v", decoded.AllowAttachments, allowAttach)
	}
}

// TestToolDefinitionRoundTrip tests JSON marshaling and unmarshaling of ToolDefinition
func TestToolDefinitionRoundTrip(t *testing.T) {
	original := ToolDefinition{
		Type: "function",
		Function: FunctionDefinition{
			Name:        "get_weather",
			Description: "Get the current weather",
			Parameters: ParametersDefinition{
				Type: "object",
				Properties: map[string]any{
					"location": map[string]string{
						"type":        "string",
						"description": "The city name",
					},
				},
				Required: []string{"location"},
			},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal ToolDefinition: %v", err)
	}

	var decoded ToolDefinition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ToolDefinition: %v", err)
	}

	if decoded.Type != original.Type {
		t.Errorf("Type mismatch: got %q, want %q", decoded.Type, original.Type)
	}
	if decoded.Function.Name != original.Function.Name {
		t.Errorf("Function.Name mismatch: got %q, want %q", decoded.Function.Name, original.Function.Name)
	}
	if decoded.Function.Parameters.Type != "object" {
		t.Errorf("Parameters.Type mismatch: got %q, want %q", decoded.Function.Parameters.Type, "object")
	}
	if len(decoded.Function.Parameters.Required) != 1 || decoded.Function.Parameters.Required[0] != "location" {
		t.Errorf("Parameters.Required mismatch: got %v, want [\"location\"]", decoded.Function.Parameters.Required)
	}
}

// TestToolCallDefinitionDeltaRoundTrip tests JSON marshaling and unmarshaling of ToolCallDefinitionDelta
func TestToolCallDefinitionDeltaRoundTrip(t *testing.T) {
	id := "call_123"
	toolType := "function"
	funcName := "get_weather"

	original := ToolCallDefinitionDelta{
		Index: 0,
		ID:    &id,
		Type:  &toolType,
		Function: FunctionCallDefinitionDelta{
			Name:      &funcName,
			Arguments: `{"location":"San Francisco"}`,
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal ToolCallDefinitionDelta: %v", err)
	}

	var decoded ToolCallDefinitionDelta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal ToolCallDefinitionDelta: %v", err)
	}

	if decoded.Index != original.Index {
		t.Errorf("Index mismatch: got %d, want %d", decoded.Index, original.Index)
	}
	if decoded.ID == nil || *decoded.ID != id {
		t.Errorf("ID mismatch: got %v, want %q", decoded.ID, id)
	}
	if decoded.Type == nil || *decoded.Type != toolType {
		t.Errorf("Type mismatch: got %v, want %q", decoded.Type, toolType)
	}
	if decoded.Function.Name == nil || *decoded.Function.Name != funcName {
		t.Errorf("Function.Name mismatch: got %v, want %q", decoded.Function.Name, funcName)
	}
	if decoded.Function.Arguments != original.Function.Arguments {
		t.Errorf("Function.Arguments mismatch: got %q, want %q", decoded.Function.Arguments, original.Function.Arguments)
	}
}

// TestCostItemFloatCeiling tests that CostItem rounds floats up using math.Ceil
func TestCostItemFloatCeiling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "float 1.2 rounds to 2",
			input:    `{"amount_usd_milli_cents": 1.2}`,
			expected: 2,
		},
		{
			name:     "float 3.0 stays 3",
			input:    `{"amount_usd_milli_cents": 3.0}`,
			expected: 3,
		},
		{
			name:     "integer stays exact",
			input:    `{"amount_usd_milli_cents": 5}`,
			expected: 5,
		},
		{
			name:     "float 2.9 rounds to 3",
			input:    `{"amount_usd_milli_cents": 2.9}`,
			expected: 3,
		},
		{
			name:     "negative float -1.5 rounds to -1",
			input:    `{"amount_usd_milli_cents": -1.5}`,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var item CostItem
			if err := json.Unmarshal([]byte(tt.input), &item); err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			if item.AmountUSDMilliCents != tt.expected {
				t.Errorf("AmountUSDMilliCents = %d, want %d", item.AmountUSDMilliCents, tt.expected)
			}
		})
	}
}

// TestBaseControlUnionMarshalUnmarshal tests BaseControl discriminated union
func TestBaseControlUnionMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		control any
	}{
		{
			name:    "Divider",
			control: Divider{Control: "divider"},
		},
		{
			name: "TextField",
			control: TextField{
				Control:       "text_field",
				Label:         "Username",
				ParameterName: "username",
			},
		},
		{
			name: "TextArea",
			control: TextArea{
				Control:       "text_area",
				Label:         "Description",
				ParameterName: "desc",
			},
		},
		{
			name: "DropDown",
			control: DropDown{
				Control:       "drop_down",
				Label:         "Size",
				ParameterName: "size",
				Options: []ValueNamePair{
					{Value: "s", Name: "Small"},
					{Value: "m", Name: "Medium"},
				},
			},
		},
		{
			name: "ToggleSwitch",
			control: ToggleSwitch{
				Control:       "toggle_switch",
				Label:         "Enable feature",
				ParameterName: "enabled",
			},
		},
		{
			name: "Slider",
			control: Slider{
				Control:       "slider",
				Label:         "Volume",
				ParameterName: "volume",
				MinValue:      0,
				MaxValue:      100,
				Step:          1,
			},
		},
		{
			name: "AspectRatio",
			control: AspectRatio{
				Control:       "aspect_ratio",
				Label:         "Ratio",
				ParameterName: "ratio",
				Options: []AspectRatioOption{
					{Width: 16, Height: 9},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := NewBaseControl(tt.control)

			// Marshal
			data, err := json.Marshal(bc)
			if err != nil {
				t.Fatalf("Failed to marshal BaseControl: %v", err)
			}

			// Unmarshal
			var decoded BaseControl
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal BaseControl: %v", err)
			}

			// Check type matches
			originalData, _ := json.Marshal(tt.control)
			decodedData, _ := json.Marshal(decoded.Underlying())

			if string(originalData) != string(decodedData) {
				t.Errorf("Mismatch:\noriginal: %s\ndecoded:  %s", originalData, decodedData)
			}
		})
	}
}

// TestFullControlUnionMarshalUnmarshal tests FullControl discriminated union
func TestFullControlUnionMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		control any
	}{
		{
			name: "TextField (BaseControl type)",
			control: TextField{
				Control:       "text_field",
				Label:         "Name",
				ParameterName: "name",
			},
		},
		{
			name: "ConditionallyRenderControls",
			control: ConditionallyRenderControls{
				Control: "condition",
				Condition: ComparatorCondition{
					Comparator: "eq",
					Left:       LiteralValue{Literal: "test"},
					Right:      ParameterValue{ParameterName: "param1"},
				},
				Controls: []BaseControl{
					NewBaseControl(Divider{Control: "divider"}),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := NewFullControl(tt.control)

			// Marshal
			data, err := json.Marshal(fc)
			if err != nil {
				t.Fatalf("Failed to marshal FullControl: %v", err)
			}

			// Unmarshal
			var decoded FullControl
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal FullControl: %v", err)
			}

			// Check type matches
			originalData, _ := json.Marshal(tt.control)
			decodedData, _ := json.Marshal(decoded.Underlying())

			if string(originalData) != string(decodedData) {
				t.Errorf("Mismatch:\noriginal: %s\ndecoded:  %s", originalData, decodedData)
			}
		})
	}
}

// TestSettingsResponseDefaultVersion tests that NewSettingsResponse sets version to 2
func TestSettingsResponseDefaultVersion(t *testing.T) {
	sr := NewSettingsResponse()
	if sr.ResponseVersion == nil {
		t.Fatal("ResponseVersion should not be nil")
	}
	if *sr.ResponseVersion != 2 {
		t.Errorf("ResponseVersion = %d, want 2", *sr.ResponseVersion)
	}
}

// TestErrorResponseDefaultAllowRetry tests that NewErrorResponse sets AllowRetry to true
func TestErrorResponseDefaultAllowRetry(t *testing.T) {
	er := NewErrorResponse("test error")
	if !er.AllowRetry {
		t.Error("AllowRetry should be true by default")
	}
	if er.PartialResponse.Text != "test error" {
		t.Errorf("Text = %q, want %q", er.PartialResponse.Text, "test error")
	}
}

// TestMetaResponseDefaults tests NewMetaResponse default values
func TestMetaResponseDefaults(t *testing.T) {
	mr := NewMetaResponse()
	if !mr.Linkify {
		t.Error("Linkify should be true by default")
	}
	if !mr.SuggestedReplies {
		t.Error("SuggestedReplies should be true by default")
	}
	if mr.ContentType != ContentTypeMarkdown {
		t.Errorf("ContentType = %q, want %q", mr.ContentType, ContentTypeMarkdown)
	}
}

// TestBotEventInterface tests that response types implement BotEvent
func TestBotEventInterface(t *testing.T) {
	var _ BotEvent = &PartialResponse{}
	var _ BotEvent = &ErrorResponse{}
	var _ BotEvent = &MetaResponse{}
	var _ BotEvent = &DataResponse{}
}

// TestParseRawRequest tests ParseRawRequest function
func TestParseRawRequest(t *testing.T) {
	input := `{"version":"1.2","type":"query","query":[],"user_id":"u1","conversation_id":"c1","message_id":"m1"}`

	reqType, rawMsg, err := ParseRawRequest([]byte(input))
	if err != nil {
		t.Fatalf("ParseRawRequest failed: %v", err)
	}

	if reqType != RequestTypeQuery {
		t.Errorf("reqType = %q, want %q", reqType, RequestTypeQuery)
	}

	var qr QueryRequest
	if err := json.Unmarshal(rawMsg, &qr); err != nil {
		t.Fatalf("Failed to unmarshal raw message: %v", err)
	}

	if qr.UserID != "u1" {
		t.Errorf("UserID = %q, want %q", qr.UserID, "u1")
	}
}

// ptr is a helper to create a pointer to a value
func ptr(i int) *int {
	return &i
}
