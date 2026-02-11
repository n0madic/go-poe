package types

// ParametersDefinition defines parameters for function calling
type ParametersDefinition struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required,omitempty"`
}

// FunctionDefinition defines a function for OpenAI function calling
type FunctionDefinition struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Parameters  ParametersDefinition `json:"parameters"`
}

// ToolDefinition represents a tool definition for OpenAI function calling
type ToolDefinition struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

// FunctionCallDefinition represents a function call
type FunctionCallDefinition struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolCallDefinition represents a tool call returned by the model
type ToolCallDefinition struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCallDefinition `json:"function"`
}

// ToolResultDefinition represents a function result
type ToolResultDefinition struct {
	Role       string `json:"role"`
	Name       string `json:"name"`
	ToolCallID string `json:"tool_call_id"`
	Content    string `json:"content"`
}

// FunctionCallDefinitionDelta is a streaming function call delta
type FunctionCallDefinitionDelta struct {
	Name      *string `json:"name,omitempty"`
	Arguments string  `json:"arguments"`
}

// ToolCallDefinitionDelta is a streaming tool call chunk
type ToolCallDefinitionDelta struct {
	Index    int                         `json:"index"`
	ID       *string                     `json:"id,omitempty"`
	Type     *string                     `json:"type,omitempty"`
	Function FunctionCallDefinitionDelta `json:"function"`
}

// CustomToolDefinition is for OpenAI-compatible custom tools
type CustomToolDefinition struct {
	Name        string         `json:"name"`
	Description *string        `json:"description,omitempty"`
	Format      map[string]any `json:"format,omitempty"`
}

// CustomCallDefinition is a custom tool call in model response
type CustomCallDefinition struct {
	Name  string `json:"name"`
	Input string `json:"input"`
}
