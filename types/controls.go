package types

// Divider control
type Divider struct {
	Control string `json:"control"`
}

// TextField control
type TextField struct {
	Control       string  `json:"control"`
	Label         string  `json:"label"`
	Description   *string `json:"description,omitempty"`
	ParameterName string  `json:"parameter_name"`
	DefaultValue  *string `json:"default_value,omitempty"`
	Placeholder   *string `json:"placeholder,omitempty"`
}

// TextArea control
type TextArea struct {
	Control       string  `json:"control"`
	Label         string  `json:"label"`
	Description   *string `json:"description,omitempty"`
	ParameterName string  `json:"parameter_name"`
	DefaultValue  *string `json:"default_value,omitempty"`
	Placeholder   *string `json:"placeholder,omitempty"`
}

// ValueNamePair for dropdown options
type ValueNamePair struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// DropDown control
type DropDown struct {
	Control       string          `json:"control"`
	Label         string          `json:"label"`
	Description   *string         `json:"description,omitempty"`
	ParameterName string          `json:"parameter_name"`
	DefaultValue  *string         `json:"default_value,omitempty"`
	Options       []ValueNamePair `json:"options"`
}

// ToggleSwitch control
type ToggleSwitch struct {
	Control       string  `json:"control"`
	Label         string  `json:"label"`
	Description   *string `json:"description,omitempty"`
	ParameterName string  `json:"parameter_name"`
	DefaultValue  *bool   `json:"default_value,omitempty"`
}

// Slider control
type Slider struct {
	Control       string  `json:"control"`
	Label         string  `json:"label"`
	Description   *string `json:"description,omitempty"`
	ParameterName string  `json:"parameter_name"`
	DefaultValue  *Number `json:"default_value,omitempty"`
	MinValue      Number  `json:"min_value"`
	MaxValue      Number  `json:"max_value"`
	Step          Number  `json:"step"`
}

// AspectRatioOption for aspect ratio control
type AspectRatioOption struct {
	Value  *string `json:"value,omitempty"`
	Width  Number  `json:"width"`
	Height Number  `json:"height"`
}

// AspectRatio control
type AspectRatio struct {
	Control       string              `json:"control"`
	Label         string              `json:"label"`
	Description   *string             `json:"description,omitempty"`
	ParameterName string              `json:"parameter_name"`
	DefaultValue  *string             `json:"default_value,omitempty"`
	Options       []AspectRatioOption `json:"options"`
}

// LiteralValue holds a literal value for conditions
type LiteralValue struct {
	Literal any `json:"literal"` // string, float64, int, bool
}

// ParameterValue references a parameter by name
type ParameterValue struct {
	ParameterName string `json:"parameter_name"`
}

// ComparatorCondition is a condition with a comparator
type ComparatorCondition struct {
	Comparator string `json:"comparator"` // "eq", "ne", "gt", "ge", "lt", "le"
	Left       any    `json:"left"`       // LiteralValue or ParameterValue
	Right      any    `json:"right"`      // LiteralValue or ParameterValue
}

// ConditionallyRenderControls conditionally renders controls
type ConditionallyRenderControls struct {
	Control   string              `json:"control"`
	Condition ComparatorCondition `json:"condition"`
	Controls  []BaseControl       `json:"controls"`
}

// Tab groups controls
type Tab struct {
	Name     *string       `json:"name,omitempty"`
	Controls []FullControl `json:"controls"`
}

// Section groups controls or tabs
type Section struct {
	Name               *string       `json:"name,omitempty"`
	Controls           []FullControl `json:"controls,omitempty"`
	Tabs               []Tab         `json:"tabs,omitempty"`
	CollapsedByDefault *bool         `json:"collapsed_by_default,omitempty"`
}

// ParameterControls is the top-level parameter controls structure
type ParameterControls struct {
	APIVersion string    `json:"api_version"`
	Sections   []Section `json:"sections"`
}
