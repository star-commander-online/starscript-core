//go:build js && wasm

package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"syscall/js"

	"github.com/star-commander-online/starscript-core/command"
	"github.com/star-commander-online/starscript-core/runtime"
	"github.com/star-commander-online/starscript-core/value"
)

// Global state for managing concurrent executions.
var (
	enginesMu   sync.Mutex
	engines     = map[string]*executionState{}
	runIDSeq    int
	commandsMu  sync.Mutex
	pendingCmds = map[string]chan commandResult{}
	cmdIDSeq    int
)

// executionState holds the context for one run.
type executionState struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// commandResult is what the invoker sends back after JS handles the command.
type commandResult struct {
	value value.Value
	err   error
}

func main() {
	// Register the public bridge API: execute, resolve, reject, cancel.
	// js.Func values are never released because they must remain callable
	// for the entire lifetime of the WASM instance. Cleanup is not needed.
	js.Global().Set("__starscript", js.ValueOf(map[string]any{
		"execute": js.FuncOf(jsExecute),
		"resolve": js.FuncOf(jsResolve),
		"reject":  js.FuncOf(jsReject),
		"cancel":  js.FuncOf(jsCancel),
	}))

	// Block forever so the runtime stays alive.
	select {}
}

// jsExecute(source, onCommand, onDone, opts?) runs a script asynchronously.
// Args: source (string), onCommand (function), onDone (function), opts (optional object with {maxSteps}).
// Returns immediately with a runId string.
// Calls onCommand(id, runId, rawLine) for each host command.
// Calls onDone(result) when execution finishes (result is {value, variables, steps} or {code, message, line}).
func jsExecute(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return nil
	}

	source := args[0].String()
	onCommand := args[1]
	onDone := args[2]

	// Parse optional opts (4th argument).
	opts := runtime.Options{}
	if len(args) > 3 && args[3].Type() == js.TypeObject {
		if maxSteps := args[3].Get("maxSteps"); maxSteps.Type() == js.TypeNumber {
			opts.MaxSteps = maxSteps.Int()
		}
	}

	// Allocate a unique run ID and create a cancellable context.
	enginesMu.Lock()
	runIDSeq++
	runID := fmt.Sprintf("run-%d", runIDSeq)
	ctx, cancel := context.WithCancel(context.Background())
	engines[runID] = &executionState{ctx: ctx, cancel: cancel}
	enginesMu.Unlock()

	// Run execution in a goroutine so jsExecute returns immediately.
	// JS event loop is never blocked inside a callback.
	go func() {
		// Ensure cancellation and cleanup even if execution panics.
		defer cancel()
		defer func() {
			enginesMu.Lock()
			delete(engines, runID)
			enginesMu.Unlock()
		}()

		// Catch any panic and report it as a runtime error.
		defer func() {
			if r := recover(); r != nil {
				onDone.Invoke(map[string]any{
					"code":    "runtime_error",
					"message": fmt.Sprintf("panic: %v", r),
					"line":    0,
				})
			}
		}()

		// Create a new engine with a bridge invoker.
		invoker := &wasmInvoker{
			runID:     runID,
			onCommand: onCommand,
		}
		engine := runtime.NewEngine(invoker, opts)

		// Execute the script with the cancellable context.
		result, err := engine.Execute(ctx, source)

		// Build result/error objects to send back to JS.
		var resultObj any
		if err != nil {
			resultObj = errorToJS(err)
		} else {
			// Convert ExecutionResult to JS object.
			vars := make(map[string]any)
			for k, v := range result.Variables {
				vars[k] = v.ToAny()
			}
			resultObj = map[string]any{
				"value":     result.Value.ToAny(),
				"variables": vars,
				"steps":     result.Steps,
			}
		}

		// Call the JS onDone callback with the result.
		onDone.Invoke(resultObj)
	}()

	return runID
}

// jsResolve(id, value) completes a pending command with a successful result.
// The waiting goroutine's Invoke() will return the value.
func jsResolve(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return nil
	}

	id := args[0].String()
	jsValue := args[1]

	// Convert JS value to Go Value (scalar only).
	val, err := jsToValue(jsValue)

	// Find the waiting goroutine and send the result.
	commandsMu.Lock()
	ch, ok := pendingCmds[id]
	delete(pendingCmds, id)
	commandsMu.Unlock()

	if ok {
		if err != nil {
			// Unsupported JS type; deliver as error to the script.
			ch <- commandResult{value: value.Null(), err: err}
		} else {
			ch <- commandResult{value: val, err: nil}
		}
	}

	return nil
}

// jsReject(id, code, message) completes a pending command with an error.
// code can be "unknown_command" or other error types.
// The waiting goroutine's Invoke() will return an error.
func jsReject(this js.Value, args []js.Value) any {
	if len(args) < 3 {
		return nil
	}

	id := args[0].String()
	code := args[1].String()
	message := args[2].String()

	// Build the appropriate error for the code.
	var err error
	err = fmt.Errorf("%s: %s", code, message)

	// Find the waiting goroutine and send the error.
	commandsMu.Lock()
	ch, ok := pendingCmds[id]
	delete(pendingCmds, id)
	commandsMu.Unlock()

	if ok {
		ch <- commandResult{value: value.Null(), err: err}
	}

	return nil
}

// jsCancel(runId) cancels an in-flight execution by cancelling its context.
func jsCancel(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return nil
	}

	runID := args[0].String()

	enginesMu.Lock()
	state, ok := engines[runID]
	enginesMu.Unlock()

	if ok && state.cancel != nil {
		state.cancel()
	}

	return nil
}

// wasmInvoker implements command.Invoker by forwarding all commands to JS.
type wasmInvoker struct {
	runID     string
	onCommand js.Value
}

// Invoke(ctx, call) blocks until the host JS side calls resolve or reject.
// This implements the async command bridge: allocation, channel blocking, and result unwrapping.
func (w *wasmInvoker) Invoke(ctx context.Context, call command.Call) (value.Value, error) {
	// Allocate a unique command ID.
	commandsMu.Lock()
	cmdIDSeq++
	cmdID := fmt.Sprintf("%s-cmd-%d", w.runID, cmdIDSeq)

	// Create a channel for this command and store it.
	// This channel will be closed when JS calls resolve/reject.
	resultCh := make(chan commandResult, 1)
	pendingCmds[cmdID] = resultCh
	commandsMu.Unlock()

	// Ensure the pendingCmds entry is removed even if we exit early via context cancellation.
	defer func() {
		commandsMu.Lock()
		delete(pendingCmds, cmdID)
		commandsMu.Unlock()
	}()

	// Call JS onCommand(id, runId, rawLine).
	// This returns immediately; JS handles the command asynchronously.
	// The host is responsible for parsing and executing the raw line.
	w.onCommand.Invoke(cmdID, w.runID, call.RawLine)

	// Block until JS calls resolve/reject, which sends on resultCh.
	// If the context is cancelled, return early.
	select {
	case <-ctx.Done():
		return value.Null(), ctx.Err()
	case result := <-resultCh:
		return result.value, result.err
	}
}

// jsToValue converts a JavaScript value to a Go Value (scalar types only).
func jsToValue(v js.Value) (value.Value, error) {
	switch v.Type() {
	case js.TypeNull, js.TypeUndefined:
		return value.Null(), nil
	case js.TypeBoolean:
		return value.Bool(v.Bool()), nil
	case js.TypeNumber:
		// Float64 directly maps to JS number. No BigInt conversion.
		// Integers > 2^53 will lose precision, but that is inherent to JS numbers.
		return value.Number(v.Float()), nil
	case js.TypeString:
		return value.String(v.String()), nil
	default:
		return value.Null(), fmt.Errorf("unsupported JS type: %s", v.Type().String())
	}
}

// errorToJS converts a Go error to a JS object with code, message, line fields.
func errorToJS(err error) map[string]any {
	result := map[string]any{
		"code":    "runtime_error",
		"message": err.Error(),
		"line":    0,
	}

	// Extract structured info if it's a ScriptError.
	var scriptErr *runtime.ScriptError
	if errors.As(err, &scriptErr) {
		result["code"] = string(scriptErr.Code)
		result["message"] = scriptErr.Message
		if scriptErr.Line > 0 {
			result["line"] = scriptErr.Line
		}
	}

	return result
}
