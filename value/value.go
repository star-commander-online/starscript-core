package value

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Kind string

const (
	KindNull   Kind = "null"
	KindBool   Kind = "bool"
	KindNumber Kind = "number"
	KindString Kind = "string"
)

type Value struct {
	Kind   Kind    `json:"kind"`
	Bool   bool    `json:"bool,omitempty"`
	Number float64 `json:"number,omitempty"`
	String string  `json:"string,omitempty"`
}

func Null() Value { return Value{Kind: KindNull} }

func Bool(v bool) Value { return Value{Kind: KindBool, Bool: v} }

func Number(v float64) Value { return Value{Kind: KindNumber, Number: v} }

func String(v string) Value { return Value{Kind: KindString, String: v} }

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
