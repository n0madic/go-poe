package types

import (
	"encoding/json"
	"fmt"
)

// BaseControl is a discriminated union of control types (excluding ConditionallyRenderControls)
type BaseControl struct {
	value any // Divider, TextField, TextArea, DropDown, ToggleSwitch, Slider, or AspectRatio
}

// NewBaseControl wraps a control struct into a BaseControl
func NewBaseControl(v any) BaseControl {
	return BaseControl{value: v}
}

// Underlying returns the underlying concrete type
func (bc BaseControl) Underlying() any {
	return bc.value
}

func (bc BaseControl) MarshalJSON() ([]byte, error) {
	if bc.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(bc.value)
}

func (bc *BaseControl) UnmarshalJSON(data []byte) error {
	var peek struct {
		Control string `json:"control"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}
	switch peek.Control {
	case "divider":
		var v Divider
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		bc.value = v
	case "text_field":
		var v TextField
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		bc.value = v
	case "text_area":
		var v TextArea
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		bc.value = v
	case "drop_down":
		var v DropDown
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		bc.value = v
	case "toggle_switch":
		var v ToggleSwitch
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		bc.value = v
	case "slider":
		var v Slider
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		bc.value = v
	case "aspect_ratio":
		var v AspectRatio
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		bc.value = v
	default:
		return fmt.Errorf("unknown base control type: %q", peek.Control)
	}
	return nil
}

// FullControl is a discriminated union that includes all BaseControl types plus ConditionallyRenderControls
type FullControl struct {
	value any
}

// NewFullControl wraps a control struct into a FullControl
func NewFullControl(v any) FullControl {
	return FullControl{value: v}
}

// Underlying returns the underlying concrete type
func (fc FullControl) Underlying() any {
	return fc.value
}

func (fc FullControl) MarshalJSON() ([]byte, error) {
	if fc.value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(fc.value)
}

func (fc *FullControl) UnmarshalJSON(data []byte) error {
	var peek struct {
		Control string `json:"control"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return err
	}
	switch peek.Control {
	case "condition":
		var v ConditionallyRenderControls
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		fc.value = v
	default:
		// Try as BaseControl
		var bc BaseControl
		if err := json.Unmarshal(data, &bc); err != nil {
			return err
		}
		fc.value = bc.Underlying()
	}
	return nil
}
