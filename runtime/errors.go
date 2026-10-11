package runtime

import "fmt"

type ErrorCode string

const (
	ErrSyntax         ErrorCode = "syntax_error"    // Parse error in script source text (invalid syntax, block mismatches)
	ErrRuntime        ErrorCode = "runtime_error"   // Execution error (undefined variable, type mismatch, etc)
	ErrUnknownCommand ErrorCode = "unknown_command" // Host command not found
	ErrBudgetExceeded ErrorCode = "budget_exceeded" // Execution step limit reached
)

type ScriptError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Line    int       `json:"line,omitempty"`
}

func (e *ScriptError) Error() string {
	// Include line number if available, otherwise just show code and message
	if e.Line > 0 {
		return fmt.Sprintf("%s on line %d: %s", e.Code, e.Line, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func newRuntimeError(line int, msg string) error {
	return &ScriptError{Code: ErrRuntime, Message: msg, Line: line}
}
