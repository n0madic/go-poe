package types

import (
	"encoding/json"
	"fmt"
	"math"
)

// CostItem represents a cost item for authorization and charge requests
type CostItem struct {
	AmountUSDMilliCents int     `json:"amount_usd_milli_cents"`
	Description         *string `json:"description,omitempty"`
}

// costItemJSON is the JSON representation for custom unmarshaling
type costItemJSON struct {
	AmountUSDMilliCents json.Number `json:"amount_usd_milli_cents"`
	Description         *string     `json:"description,omitempty"`
}

// UnmarshalJSON implements custom JSON unmarshaling for CostItem.
// If amount_usd_milli_cents is a float, it is rounded up using math.Ceil.
func (c *CostItem) UnmarshalJSON(data []byte) error {
	var raw costItemJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Try int first
	if i, err := raw.AmountUSDMilliCents.Int64(); err == nil {
		c.AmountUSDMilliCents = int(i)
	} else if f, err := raw.AmountUSDMilliCents.Float64(); err == nil {
		c.AmountUSDMilliCents = int(math.Ceil(f))
	} else {
		return fmt.Errorf(
			"invalid amount: expected a number for amount_usd_milli_cents, got %s",
			raw.AmountUSDMilliCents.String(),
		)
	}
	c.Description = raw.Description
	return nil
}
