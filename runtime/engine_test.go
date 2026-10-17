package runtime

import (
	"context"
	"errors"
	"testing"

	"github.com/star-commander-online/starscript-core/command"
	"github.com/star-commander-online/starscript-core/value"
)

type testInvoker struct {
	fn func(ctx context.Context, call command.Call) (value.Value, error)
}

func (i testInvoker) Invoke(ctx context.Context, call command.Call) (value.Value, error) {
	return i.fn(ctx, call)
}

func TestExecute_AssignsAndReturns(t *testing.T) {
	invoker := testInvoker{fn: func(ctx context.Context, call command.Call) (value.Value, error) {
		return value.Null(), errors.New("no command should be called in this test")
	}}

	// Test assignment and return statement
	engine := NewEngine(invoker, Options{MaxSteps: 20})
	result, err := engine.Execute(context.Background(), "target = \"sector-7\"\ndoubled = 1 + 1\nreturn $target + \":\" + $doubled")
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if got := result.Value.Stringify(); got != "sector-7:2" {
		t.Fatalf("unexpected return value: %s", got)
	}
}

func TestExecute_IfWhileMathAndInterpolation(t *testing.T) {
	engine := NewEngine(testInvoker{fn: func(ctx context.Context, call command.Call) (value.Value, error) {
		return value.Null(), nil
	}}, Options{MaxSteps: 200})

	// Test control flow with while and if blocks with return
	script := `
sum = 0
i = 0
while $i < 4 do
  sum = $sum + $i
  i = $i + 1
end
if $sum == 6 do
  return "sum=" + $sum
else do
  return "bad"
end
`

	result, err := engine.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("execute returned error: %v", err)
	}
	if got := result.Value.Stringify(); got != "sum=6" {
		t.Fatalf("unexpected return value: %s", got)
	}
}

func TestExecute_BudgetExceeded(t *testing.T) {
	engine := NewEngine(testInvoker{fn: func(ctx context.Context, call command.Call) (value.Value, error) {
		return value.Null(), nil
	}}, Options{MaxSteps: 1})

	_, err := engine.Execute(context.Background(), "a = 1\nb = 2")
	if err == nil {
		t.Fatalf("expected budget error")
	}

	var scriptErr *ScriptError
	if !AsScriptError(err, &scriptErr) {
		t.Fatalf("expected ScriptError, got %T", err)
	}
	if scriptErr.Code != ErrBudgetExceeded {
		t.Fatalf("expected %q, got %q", ErrBudgetExceeded, scriptErr.Code)
	}
}
