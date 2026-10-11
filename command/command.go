package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/star-commander-online/starscript-core/value"
)

// Call describes one command invocation requested by a script.
type Call struct {
	Name string        `json:"name"`
	Args []value.Value `json:"args"`
}

// Invoker is implemented by host integrations and is owned by the embedding app.
type Invoker interface {
	Invoke(ctx context.Context, call Call) (value.Value, error)
}

// UnknownCommandError allows host implementations to signal a missing command
// in a way that the runtime can map to ErrUnknownCommand.
type UnknownCommandError struct {
	Name string
}

func (e *UnknownCommandError) Error() string {
	return fmt.Sprintf("unknown command: %s", e.Name)
}

// UnknownCommand creates an UnknownCommandError for host adapters.
func UnknownCommand(name string) error {
	return &UnknownCommandError{Name: name}
}

// IsUnknown reports whether an error indicates a missing command.
func IsUnknown(err error) bool {
	var unknown *UnknownCommandError
	return errors.As(err, &unknown)
}
