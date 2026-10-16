package runtime

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/star-commander-online/starscript-core/ast"
	"github.com/star-commander-online/starscript-core/command"
	"github.com/star-commander-online/starscript-core/parser"
	"github.com/star-commander-online/starscript-core/value"
)

var variablePattern = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)

type Options struct {
	MaxSteps       int
	MaxSourceBytes int
}

type ExecutionResult struct {
	Value     value.Value            `json:"value"`
	Variables map[string]value.Value `json:"variables"`
	Steps     int                    `json:"steps"`
}

type Engine struct {
	vars    map[string]value.Value
	invoker command.Invoker
	opts    Options
	steps   int
}

func NewEngine(invoker command.Invoker, opts Options) *Engine {
	// Set default options if not provided.
	if opts.MaxSteps <= 0 {
		opts.MaxSteps = 10_000 // default maximum execution steps
	}
	if opts.MaxSourceBytes <= 0 {
		opts.MaxSourceBytes = 256 * 1024 // 256 KB default limit
	}

	// Return the newly created engine instance with initialized state.
	return &Engine{
		vars:    map[string]value.Value{},
		invoker: invoker,
		opts:    opts,
	}
}

func (e *Engine) Execute(ctx context.Context, source string) (ExecutionResult, error) {
	// Enforce source size limit to prevent memory exhaustion
	if len(source) > e.opts.MaxSourceBytes {
		return ExecutionResult{}, &ScriptError{Code: ErrRuntime, Message: "Source exceeds maximum size"}
	}

	// Parse source text into statement tree
	program, err := parser.Parse(source)
	if err != nil {
		return ExecutionResult{}, &ScriptError{Code: ErrSyntax, Message: err.Error()}
	}

	// Initialize execution state
	e.steps = 0

	// Execute all statements and return result
	last, _, execErr := e.execBlock(ctx, program.Statements)
	if execErr != nil {
		return ExecutionResult{}, execErr
	}

	return ExecutionResult{Value: last, Variables: e.cloneVars(), Steps: e.steps}, nil
}

func (e *Engine) execBlock(ctx context.Context, statements []ast.Statement) (value.Value, bool, error) {
	last := value.Null()
	// Iterate over each statement in the block
	for _, stmt := range statements {
		current, returned, err := e.execStatement(ctx, stmt)
		if err != nil {
			return value.Value{}, false, err
		}
		last = current
		if returned {
			return last, true, nil
		}
	}

	return last, false, nil
}

func (e *Engine) execStatement(ctx context.Context, stmt ast.Statement) (value.Value, bool, error) {
	// Check max execution steps (budget)
	if err := e.step(ctx, stmt.Line()); err != nil {
		return value.Value{}, false, err
	}

	// Dispatch by statement type
	switch node := stmt.(type) {
	// Assignment: name = expression
	case *ast.AssignmentStmt:
		v, err := e.evalExpression(node.Expression, node.Line())
		if err != nil {
			return value.Value{}, false, err
		}
		e.vars[node.Name] = v
		return v, false, nil

	// Command: command arg1 arg2 ...
	case *ast.CommandStmt:
		v, err := e.invokeCommand(ctx, node)
		if err != nil {
			return value.Value{}, false, err
		}
		return v, false, nil

	// Return: return expression
	case *ast.ReturnStmt:
		v, err := e.evalExpression(node.Expression, node.Line())
		if err != nil {
			return value.Value{}, false, err
		}
		return v, true, nil

	// If block: if condition ... [else ...] fi
	case *ast.IfStmt:
		ok, err := e.evalCondition(node.Condition, node.Line())
		if err != nil {
			return value.Value{}, false, err
		}
		if ok {
			return e.execBlock(ctx, node.Then)
		}
		return e.execBlock(ctx, node.Else)

	// While block: while condition ... done
	case *ast.WhileStmt:
		last := value.Null()
		for { // loop forever until something is returned or the condition fails
			ok, err := e.evalCondition(node.Condition, node.Line())
			if err != nil {
				return value.Value{}, false, err
			}
			if !ok {
				return last, false, nil
			}

			// If condition is true, execute the block
			blockVal, returned, blockErr := e.execBlock(ctx, node.Body)
			if blockErr != nil {
				return value.Value{}, false, blockErr
			}
			last = blockVal
			if returned {
				return last, true, nil
			}
		}

	// For range: for var in [start..end]; do ... done
	case *ast.ForStmt:
		start, end, err := e.evalForBounds(node)
		if err != nil {
			return value.Value{}, false, err
		}
		last := value.Null()
		for i := start; i <= end; i++ {
			// Set loop variable for this iteration
			e.vars[node.VarName] = value.Number(float64(i))
			blockVal, returned, blockErr := e.execBlock(ctx, node.Body)
			if blockErr != nil {
				return value.Value{}, false, blockErr
			}
			last = blockVal
			if returned {
				// Propagate return up
				return last, true, nil
			}
		}
		return last, false, nil

	default:
		return value.Value{}, false, newRuntimeError(stmt.Line(), "Unsupported statement")
	}
}

func (e *Engine) step(ctx context.Context, line int) error {
	// Increment execution step counter
	e.steps++

	// Check if we've exceeded the maximum allowed steps
	if e.steps > e.opts.MaxSteps {
		return &ScriptError{Code: ErrBudgetExceeded, Message: "Maximum steps exceeded", Line: line}
	}

	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

func (e *Engine) cloneVars() map[string]value.Value {
	// Create a shallow copy of the variable map
	out := make(map[string]value.Value, len(e.vars))
	for k, v := range e.vars {
		out[k] = v
	}

	return out
}

func (e *Engine) invokeCommand(ctx context.Context, stmt *ast.CommandStmt) (value.Value, error) {
	// Prepare the arguments
	args := make([]value.Value, 0, len(stmt.Args))
	for _, raw := range stmt.Args {
		// Evaluate each command argument
		arg, err := e.evalCommandArg(raw, stmt.Line())
		if err != nil {
			return value.Value{}, err
		}
		args = append(args, arg)
	}

	// Invoke the command
	out, err := e.invoker.Invoke(ctx, command.Call{Name: stmt.Name, Args: args, RawLine: stmt.RawLine})
	if err != nil {
		// Handle command invocation errors
		if command.IsUnknown(err) {
			return value.Value{}, &ScriptError{Code: ErrUnknownCommand, Message: err.Error(), Line: stmt.Line()}
		}
		return value.Value{}, &ScriptError{Code: ErrRuntime, Message: err.Error(), Line: stmt.Line()}
	}

	return out, nil
}

func (e *Engine) evalExpression(raw string, line int) (value.Value, error) {
	// Interpolate the raw expression string and evaluate it
	interpolated, err := e.interpolate(raw, line)
	if err != nil {
		return value.Value{}, err
	}

	// Evaluate the interpolated expression
	v, evalErr := evalInterpolatedExpression(interpolated)
	if evalErr != nil {
		return value.String(interpolated), nil
	}

	return v, nil
}

func (e *Engine) evalCondition(raw string, line int) (bool, error) {
	// Interpolate the raw condition string and evaluate it as an expression
	interpolated, err := e.interpolate(raw, line)
	if err != nil {
		return false, err
	}

	// Evaluate the interpolated condition expression
	v, evalErr := evalInterpolatedExpression(interpolated)
	if evalErr != nil {
		return false, newRuntimeError(line, fmt.Sprintf("Invalid condition %q: %v", interpolated, evalErr))
	}

	// Convert the evaluated value to a boolean
	return isTruthy(v), nil
}

func (e *Engine) evalForBounds(stmt *ast.ForStmt) (int, int, error) {
	// Evaluate the start and end expressions for the for loop range
	startVal, err := e.evalExpression(stmt.StartExpr, stmt.Line())
	if err != nil {
		return 0, 0, err
	}
	endVal, err := e.evalExpression(stmt.EndExpr, stmt.Line())
	if err != nil {
		return 0, 0, err
	}

	// Convert the evaluated start and end values to non-negative integers
	start, err := valueToNonNegativeInt(startVal, stmt.Line())
	if err != nil {
		return 0, 0, err
	}
	end, err := valueToNonNegativeInt(endVal, stmt.Line())
	if err != nil {
		return 0, 0, err
	}

	// Ensure the end of the range is not less than the start
	if end < start {
		return 0, 0, newRuntimeError(stmt.Line(), "for range end must be >= start")
	}

	return start, end, nil
}

func (e *Engine) evalCommandArg(raw string, line int) (value.Value, error) {
	// Interpolate the raw command argument
	interpolated, err := e.interpolate(raw, line)
	if err != nil {
		return value.Value{}, err
	}

	// Evaluate the interpolated command argument as an expression
	v, evalErr := evalInterpolatedExpression(interpolated)
	if evalErr != nil {
		// Fallback to using interpolated string
		return value.String(interpolated), nil
	}

	return v, nil
}

func (e *Engine) interpolate(raw string, line int) (string, error) {
	// Replace all variable placeholders with their corresponding values
	out := variablePattern.ReplaceAllStringFunc(raw, func(part string) string {
		// Extract the variable name from the matched placeholder.
		matches := variablePattern.FindStringSubmatch(part)
		if len(matches) != 2 {
			return part
		}
		name := matches[1]
		v, ok := e.vars[name]
		if !ok {
			// Keep sentinel marker; validated immediately after replacement
			return "__UNRESOLVED__" + name
		}
		return asExpressionLiteral(v)
	})

	// Check if there are any unresolved variable placeholders after interpolation
	if strings.Contains(out, "__UNRESOLVED__") {
		idx := strings.Index(out, "__UNRESOLVED__")
		name := out[idx+len("__UNRESOLVED__"):]
		// Extract the variable name from the unresolved placeholder
		if cut := strings.IndexAny(name, " +-*/()&|!=<>\t\r\n"); cut >= 0 {
			name = name[:cut]
		}
		// Return error indicating unknown variable
		return "", newRuntimeError(line, fmt.Sprintf("Unknown variable %q", name))
	}

	return out, nil
}

func asExpressionLiteral(v value.Value) string {
	// Return the value as a literal suitable for use in an expression
	switch v.Kind {
	case value.KindNull, value.KindBool:
		return v.Stringify()
	case value.KindNumber:
		return strconv.FormatFloat(v.Number, 'f', -1, 64)
	case value.KindString:
		return strconv.Quote(v.String)
	default:
		// Arrays/objects are intentionally out-of-scope for this simplified mode
		return strconv.Quote(v.Stringify())
	}
}

func valueToNonNegativeInt(v value.Value, line int) (int, error) {
	if v.Kind != value.KindNumber {
		return 0, newRuntimeError(line, "for bounds must evaluate to numbers")
	}
	if v.Number < 0 {
		return 0, newRuntimeError(line, "for bounds must be non-negative")
	}
	if math.Trunc(v.Number) != v.Number {
		return 0, newRuntimeError(line, "for bounds must be integers")
	}
	return int(v.Number), nil
}

func isTruthy(v value.Value) bool {
	switch v.Kind {
	case value.KindNull:
		return false
	case value.KindBool:
		return v.Bool
	case value.KindNumber:
		return v.Number != 0 && !math.IsNaN(v.Number)
	case value.KindString:
		return v.String != ""
	default:
		return false
	}
}

// TODO CHECK
// AsScriptError is a small helper used by tests and host integrations
// to safely downcast runtime errors.
func AsScriptError(err error, out **ScriptError) bool {
	v, ok := err.(*ScriptError)
	if !ok {
		return false
	}
	*out = v
	return true
}
