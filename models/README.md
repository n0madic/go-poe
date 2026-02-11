# models

Go package for fetching the [Poe model catalog](https://api.poe.com/v1/models). Returns structured types with model properties including pricing, context window, architecture, reasoning config, and parameters.

No authentication required — the endpoint is public.

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/n0madic/go-poe/models"
)

func main() {
	list, err := models.Fetch(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range list {
		fmt.Printf("%-30s %-15s %s\n", m.ID, m.OwnedBy, m.Architecture.Modality)
	}
}
```

## Options

```go
list, err := models.Fetch(ctx, &models.Options{
	BaseURL:      "https://api.poe.com/v1/models", // default
	HTTPClient:   customClient,                     // default: 30s timeout
	ExtraHeaders: map[string]string{"X-Custom": "value"},
})
```

## Types

| Type | Description |
|------|-------------|
| `Model` | A single model with all properties |
| `Architecture` | Input/output modalities |
| `Pricing` | Per-token pricing (decimal strings) |
| `ContextWindow` | Context length and max output tokens |
| `ModelMetadata` | Display name, image, URL |
| `Reasoning` | Reasoning/thinking budget and capabilities |
| `Parameter` | Configurable model parameter with JSON schema |
