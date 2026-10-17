package conformance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/star-commander-online/starscript-core/command"
	"github.com/star-commander-online/starscript-core/runtime"
	"github.com/star-commander-online/starscript-core/value"
)

// TestCase represents one conformance test case
type TestCase struct {
	Name     string                 `json:"name"`
	Source   string                 `json:"source"`
	Commands map[string]CommandSpec `json:"commands"`
	Options  Options                `json:"options"`
	Expected Expected               `json:"expected"`
}

// CommandSpec defines a mocked command behavior
type CommandSpec struct {
	Result *ValueSpec `json:"result"`
	Error  *string    `json:"error"`
}

// ValueSpec represents an expected value
type ValueSpec struct {
	Kind  string      `json:"kind"`
	Value interface{} `json:"value"`
}

// Options for execution
type Options struct {
	MaxSteps       int `json:"maxSteps"`
	MaxSourceBytes int `json:"maxSourceBytes"`
}

// Expected result
type Expected struct {
	ResultValue interface{}           `json:"resultValue"`
	Variables   map[string]*ValueSpec `json:"variables"`
	Error       *ErrorSpec            `json:"error"`
}

// ErrorSpec represents an expected error
type ErrorSpec struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Line    int    `json:"line"`
}

// MockInvoker implements command.Invoker with test command specs
type MockInvoker struct {
	specs map[string]CommandSpec
}

func (m *MockInvoker) Invoke(ctx context.Context, call command.Call) (value.Value, error) {
	// Extract command name from rawLine (first token)
	tokens := strings.Fields(call.RawLine)
	if len(tokens) == 0 {
		return value.Value{}, fmt.Errorf("empty command line")
	}

	cmdName := tokens[0]
	spec, ok := m.specs[cmdName]
	if !ok {
		return value.Value{}, fmt.Errorf("unknown command: %s", cmdName)
	}

	if spec.Error != nil {
		return value.Value{}, fmt.Errorf(*spec.Error)
	}

	if spec.Result == nil {
		return value.Null(), nil
	}

	return specToValue(spec.Result), nil
}

// specToValue converts a ValueSpec to a value.Value
func specToValue(spec *ValueSpec) value.Value {
	switch spec.Kind {
	case "null":
		return value.Null()
	case "bool":
		if b, ok := spec.Value.(bool); ok {
			return value.Bool(b)
		}
		return value.Bool(false)
	case "number":
		if n, ok := spec.Value.(float64); ok {
			return value.Number(n)
		}
		return value.Number(0)
	case "string":
		if s, ok := spec.Value.(string); ok {
			return value.String(s)
		}
		return value.String("")
	default:
		return value.Null()
	}
}

// TestConformanceCorpus runs all conformance test cases against the Go interpreter
func TestConformanceCorpus(t *testing.T) {
	// Load all test cases from conformance/cases/*.json
	casesDir := "cases"
	files, err := filepath.Glob(filepath.Join(casesDir, "*.json"))
	if err != nil {
		t.Fatalf("failed to find conformance cases: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("no conformance case files found in %s", casesDir)
	}

	totalCases := 0
	passedCases := 0
	failedCases := 0

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read %s: %v", file, err)
		}

		var cases []TestCase
		if err := json.Unmarshal(data, &cases); err != nil {
			t.Fatalf("failed to unmarshal %s: %v", file, err)
		}

		t.Logf("Running %d cases from %s", len(cases), filepath.Base(file))

		for _, tc := range cases {
			totalCases++

			if runTestCase(t, tc) {
				passedCases++
			} else {
				failedCases++
				t.Logf("FAILED: %s", tc.Name)
			}
		}
	}

	t.Logf("\n=== Conformance Summary ===")
	t.Logf("Total:  %d", totalCases)
	t.Logf("Passed: %d", passedCases)
	t.Logf("Failed: %d", failedCases)

	if failedCases > 0 {
		t.Fatalf("conformance test failed: %d/%d cases failed", failedCases, totalCases)
	}
}

// runTestCase executes a single test case and returns true if it passes
func runTestCase(t *testing.T, tc TestCase) bool {
	invoker := &MockInvoker{specs: tc.Commands}
	opts := runtime.Options{
		MaxSteps:       tc.Options.MaxSteps,
		MaxSourceBytes: tc.Options.MaxSourceBytes,
	}

	// Set defaults if not specified
	if opts.MaxSteps == 0 {
		opts.MaxSteps = 10_000
	}
	if opts.MaxSourceBytes == 0 {
		opts.MaxSourceBytes = 256 * 1024
	}

	engine := runtime.NewEngine(invoker, opts)
	t.Logf("  Running: %s\n    Source: %q", tc.Name, tc.Source)
	result, err := engine.Execute(context.Background(), tc.Source)

	// Check if error was expected
	if tc.Expected.Error != nil {
		if err == nil {
			t.Logf("  FAIL %s: expected error %s, got success", tc.Name, tc.Expected.Error.Code)
			return false
		}

		var scriptErr *runtime.ScriptError
		if !runtime.AsScriptError(err, &scriptErr) {
			t.Logf("  FAIL %s: expected ScriptError, got %T: %v", tc.Name, err, err)
			return false
		}

		if string(scriptErr.Code) != tc.Expected.Error.Code {
			t.Logf("  FAIL %s: expected error code %s, got %s", tc.Name, tc.Expected.Error.Code, scriptErr.Code)
			return false
		}

		if tc.Expected.Error.Line != 0 && scriptErr.Line != tc.Expected.Error.Line {
			t.Logf("  FAIL %s: expected line %d, got %d", tc.Name, tc.Expected.Error.Line, scriptErr.Line)
			return false
		}

		t.Logf("  PASS %s (error: %s)", tc.Name, scriptErr.Code)
		return true
	}

	// Check if error was unexpected
	if err != nil {
		t.Logf("  FAIL %s: unexpected error: %v", tc.Name, err)
		return false
	}

	// Check value
	if tc.Expected.ResultValue != nil {
		var expectedVal value.Value
		switch v := tc.Expected.ResultValue.(type) {
		case map[string]interface{}:
			// Handle ValueSpec format
			if kindStr, ok := v["kind"].(string); ok {
				spec := &ValueSpec{
					Kind:  kindStr,
					Value: v["value"],
				}
				expectedVal = specToValue(spec)
			}
		default:
			expectedVal = valueFromSpec(tc.Expected.ResultValue)
		}

		if !valuesEqual(result.Value, expectedVal) {
			t.Logf("  FAIL %s: expected value %v, got %v", tc.Name, expectedVal, result.Value)
			return false
		}
	}

	// Check variables (if specified)
	if tc.Expected.Variables != nil {
		for name, spec := range tc.Expected.Variables {
			actual, ok := result.Variables[name]
			if !ok {
				t.Logf("  FAIL %s: expected variable %s, not found", tc.Name, name)
				return false
			}
			expected := specToValue(spec)
			if !valuesEqual(actual, expected) {
				t.Logf("  FAIL %s: variable %s: expected %v, got %v", tc.Name, name, expected, actual)
				return false
			}
		}
	}

	t.Logf("  PASS %s", tc.Name)
	return true
}

// valueFromSpec converts a JSON value to a value.Value
func valueFromSpec(spec interface{}) value.Value {
	if spec == nil {
		return value.Null()
	}

	switch v := spec.(type) {
	case bool:
		return value.Bool(v)
	case float64:
		return value.Number(v)
	case string:
		return value.String(v)
	default:
		return value.Null()
	}
}

// valuesEqual checks if two values are equal
func valuesEqual(a, b value.Value) bool {
	if a.Kind != b.Kind {
		return false
	}

	switch a.Kind {
	case value.KindNull:
		return true
	case value.KindBool:
		return a.Bool == b.Bool
	case value.KindNumber:
		// Handle NaN specially (NaN != NaN in IEEE 754)
		if a.Number != a.Number && b.Number != b.Number {
			return true // both are NaN
		}
		return a.Number == b.Number
	case value.KindString:
		return a.String == b.String
	default:
		return false
	}
}
