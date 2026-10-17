package command

import (
	"context"

	"github.com/star-commander-online/starscript-core/value"
)

// Call describes one command invocation requested by a script.
// CommandLine has all variables interpolated ($name becomes actual value).
type Call struct {
	CommandLine string        `json:"commandLine"` // The interpolated command line with variables resolved
	Args        []value.Value `json:"args"`        // Pre-evaluated arguments (for convenience)
	LineNumber  int           `json:"lineNumber"`  // Source line number for error reporting
}

// Invoker is implemented by host integrations and is owned by the embedding app.
type Invoker interface {
	Invoke(ctx context.Context, call Call) (value.Value, error)
}
