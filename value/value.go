package value

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Kind string

const (
	KindNull   Kind = "null"
	KindBool   Kind = "bool"
	KindNumber Kind = "number"
	KindString Kind = "string"
	KindArray  Kind = "array"  // Not implemented yet
	KindObject Kind = "object" // Not implemented yet
)

type Value struct {
	Kind   Kind             `json:"kind"`
	Bool   bool             `json:"bool,omitempty"`
	Number float64          `json:"number,omitempty"`
	String string           `json:"string,omitempty"`
	Array  []Value          `json:"array,omitempty"`  // Not implemented yet
	Object map[string]Value `json:"object,omitempty"` // Not implemented yet
}

func Null() Value { return Value{Kind: KindNull} }

func Bool(v bool) Value { return Value{Kind: KindBool, Bool: v} }

func Number(v float64) Value { return Value{Kind: KindNumber, Number: v} }

func String(v string) Value { return Value{Kind: KindString, String: v} }

func Array(v []Value) Value { return Value{Kind: KindArray, Array: v} }

func Object(v map[string]Value) Value { return Value{Kind: KindObject, Object: v} }

// TODO CHECK
// ToAny converts a Value to Go's native any interface for JSON marshaling.
// This is used internally by JSON encoding and command results.
func (v Value) ToAny() any {
	switch v.Kind {
	case KindNull:
		return nil
	case KindBool:
		return v.Bool
	case KindNumber:
		return v.Number
	case KindString:
		return v.String
	case KindArray:
		// Recursively convert array elements
		out := make([]any, 0, len(v.Array))
		for _, item := range v.Array {
			out = append(out, item.ToAny())
		}
		return out
	case KindObject:
		// Recursively convert object values
		out := make(map[string]any, len(v.Object))
		for k, item := range v.Object {
			out[k] = item.ToAny()
		}
		return out
	default:
		return nil
	}
}

func (v Value) Stringify() string {
	switch v.Kind {
	case KindNull:
		return "null"
	case KindBool:
		// Boolean to text
		if v.Bool {
			return "true"
		}
		return "false"
	case KindNumber:
		// Format number without unnecessary decimals
		return strconv.FormatFloat(v.Number, 'f', -1, 64)
	case KindString:
		// String as-is
		return v.String
	default: // TODO CHECK
		// Fallback: JSON encode complex types
		b, err := json.Marshal(v.ToAny())
		if err != nil {
			return "<invalid>"
		}
		return string(b)
	}
}

// TODO CHECK
// FromAny converts Go any interface to a Value.
// Used for consuming JSON and host command results.
func FromAny(v any) (Value, error) {
	switch cast := v.(type) {
	case nil:
		return Null(), nil
	case bool:
		return Bool(cast), nil
	case int:
		return Number(float64(cast)), nil
	case int32:
		return Number(float64(cast)), nil
	case int64:
		return Number(float64(cast)), nil
	case float32:
		return Number(float64(cast)), nil
	case float64:
		return Number(cast), nil
	case string:
		return String(cast), nil
	case []any: // TODO CHECK
		// Recursively convert array elements
		arr := make([]Value, 0, len(cast))
		for _, item := range cast {
			converted, err := FromAny(item)
			if err != nil {
				return Value{}, err
			}
			arr = append(arr, converted)
		}
		return Array(arr), nil
	case map[string]any: // TODO CHECK
		// Recursively convert object values
		obj := make(map[string]Value, len(cast))
		for key, item := range cast {
			converted, err := FromAny(item)
			if err != nil {
				return Value{}, err
			}
			obj[key] = converted
		}
		return Object(obj), nil
	default:
		return Value{}, fmt.Errorf("unsupported value type: %T", v)
	}
}

// TODO CHECK
// MarshalJSON encodes a Value as JSON by converting to native Go types.
func (v Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ToAny())
}

// TODO CHECK
// UnmarshalJSON decodes JSON into a Value.
func (v *Value) UnmarshalJSON(data []byte) error {
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	converted, err := FromAny(decoded)
	if err != nil {
		return err
	}
	*v = converted
	return nil
}

// TODO CHECK
func EnsureObject(v Value) (map[string]Value, error) {
	if v.Kind != KindObject {
		return nil, errors.New("value is not an object")
	}
	return v.Object, nil
}
