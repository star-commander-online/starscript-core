package command

import (
	"context"

	"github.com/star-commander-online/starscript-core/value"
)

// Call describes one command invocation requested by a script.
type Call struct {
	RawLine string        `json:"rawLine"` // The original command line from source
	Args    []value.Value `json:"args"`    // Pre-evaluated arguments (for convenience)
}

// Invoker is implemented by host integrations and is owned by the embedding app.
type Invoker interface {
	Invoke(ctx context.Context, call Call) (value.Value, error)
}
