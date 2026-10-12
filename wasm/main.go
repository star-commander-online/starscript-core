//go:build js && wasm

package main

import (
	"context"
	"fmt"
	"sync"
	"syscall/js"

	"github.com/star-commander-online/starscript-core/command"
	"github.com/star-commander-online/starscript-core/runtime"
	"github.com/star-commander-online/starscript-core/value"
)

type wasmEngine struct {
	engine *runtime.Engine
}

type wasmInvoker struct {
	engineID int
}

var (
	enginesMu    sync.Mutex
	nextEngineID = 1
	engines      = map[int]*wasmEngine{}
)

func main() {
	api := js.ValueOf(map[string]any{
		"createEngine":  js.FuncOf(jsCreateEngine),
		"disposeEngine": js.FuncOf(jsDisposeEngine),
		"execute":       js.FuncOf(jsExecute),
	})
	js.Global().Set("StarScriptWasm", api)
	select {}
}

func jsCreateEngine(this js.Value, args []js.Value) any {
	opts := runtime.Options{}
	if len(args) > 0 && args[0].Type() == js.TypeObject {
		if maxSteps := args[0].Get("maxSteps"); maxSteps.Truthy() {
			opts.MaxSteps = maxSteps.Int()
		}
		if maxSourceBytes := args[0].Get("maxSourceBytes"); maxSourceBytes.Truthy() {
			opts.MaxSourceBytes = maxSourceBytes.Int()
		}
	}

	enginesMu.Lock()
	id := nextEngineID
	nextEngineID++
	engines[id] = &wasmEngine{engine: runtime.NewEngine(wasmInvoker{engineID: id}, opts)}
	enginesMu.Unlock()

	return id
}

func jsDisposeEngine(this js.Value, args []js.Value) any {
	if len(args) < 1 {
		return nil
	}
	id := args[0].Int()
	enginesMu.Lock()
	delete(engines, id)
	enginesMu.Unlock()
	return nil
}

func jsExecute(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return rejectPromise(fmt.Errorf("execute requires engine id and source"))
	}

	engineID := args[0].Int()
	source := args[1].String()

	enginesMu.Lock()
	engineState, ok := engines[engineID]
	enginesMu.Unlock()
	if !ok {
		return rejectPromise(fmt.Errorf("unknown engine id %d", engineID))
	}

	execFn := js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		resolve := promiseArgs[0]
		reject := promiseArgs[1]

		go func() {
			result, err := engineState.engine.Execute(context.Background(), source)
			if err != nil {
				reject.Invoke(js.ValueOf(map[string]any{
					"message": err.Error(),
				}))
				return
			}

			serializedVars := map[string]any{}
			for k, v := range result.Variables {
				serializedVars[k] = v.ToAny()
			}

			resolve.Invoke(js.ValueOf(map[string]any{
				"value":     result.Value.ToAny(),
				"variables": serializedVars,
				"steps":     result.Steps,
			}))
		}()

		return nil
	})

	promise := js.Global().Get("Promise").New(execFn)
	execFn.Release()
	return promise
}

func rejectPromise(err error) any {
	execFn := js.FuncOf(func(this js.Value, promiseArgs []js.Value) any {
		reject := promiseArgs[1]
		reject.Invoke(js.ValueOf(map[string]any{"message": err.Error()}))
		return nil
	})
	promise := js.Global().Get("Promise").New(execFn)
	execFn.Release()
	return promise
}

func (w wasmInvoker) Invoke(ctx context.Context, call command.Call) (value.Value, error) {
	_ = ctx
	invokeFn := js.Global().Get("__starscriptInvokeCommand")
	if !invokeFn.Truthy() || invokeFn.Type() != js.TypeFunction {
		return value.Value{}, fmt.Errorf("host command bridge __starscriptInvokeCommand is not installed")
	}

	args := make([]any, 0, len(call.Args))
	for _, item := range call.Args {
		args = append(args, item.ToAny())
	}

	response := invokeFn.Invoke(w.engineID, call.Name, js.ValueOf(args))
	if response.Type() != js.TypeObject || !response.Get("then").Truthy() {
		converted, err := value.FromAny(jsToAny(response))
		if err != nil {
			return value.Value{}, fmt.Errorf("invalid sync command result: %w", err)
		}
		return converted, nil
	}

	type promiseOutcome struct {
		result value.Value
		err    error
	}

	resultCh := make(chan promiseOutcome, 1)

	thenFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		converted, err := value.FromAny(jsToAny(args[0]))
		if err != nil {
			resultCh <- promiseOutcome{err: fmt.Errorf("invalid async command result: %w", err)}
			return nil
		}
		resultCh <- promiseOutcome{result: converted}
		return nil
	})
	catchFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		msg := "command rejected"
		if len(args) > 0 {
			if m := args[0].Get("message"); m.Truthy() {
				msg = m.String()
			} else {
				msg = args[0].String()
			}
		}
		resultCh <- promiseOutcome{err: fmt.Errorf(msg)}
		return nil
	})

	response.Call("then", thenFn).Call("catch", catchFn)
	outcome := <-resultCh
	thenFn.Release()
	catchFn.Release()

	if outcome.err != nil {
		return value.Value{}, outcome.err
	}
	return outcome.result, nil
}

func jsToAny(v js.Value) any {
	switch v.Type() {
	case js.TypeNull, js.TypeUndefined:
		return nil
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeNumber:
		return v.Float()
	case js.TypeString:
		return v.String()
	default:
		// Arrays and objects are not supported - return nil
		return nil
	}
}
