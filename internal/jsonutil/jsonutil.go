package jsonutil

import (
	"fmt"
	"math"
	"strings"

	"github.com/tidwall/gjson"
)

func IsEmpty(res gjson.Result) bool {
	if !res.Exists() {
		return true
	}
	if res.Type == gjson.Null {
		return true
	}

	switch res.Type {
	case gjson.String:
		return strings.TrimSpace(res.String()) == ""
	case gjson.JSON:
		if res.IsArray() {
			return len(res.Array()) == 0
		}
		if res.IsObject() {
			return len(res.Map()) == 0
		}
		return false
	default:
		return false
	}
}

func TypeToString(t gjson.Type) string {
	switch t {
	case gjson.Null:
		return "null"
	case gjson.False, gjson.True:
		return "boolean"
	case gjson.Number:
		return "number"
	case gjson.String:
		return "string"
	case gjson.JSON:
		return "object/array"
	default:
		return "unknown"
	}
}

func DebugValue(res gjson.Result) string {
	if res.Raw != "" {
		return res.Raw
	}
	return fmt.Sprintf("%v", res.Value())
}

func ToInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	default:
		return 0, false
	}
}

func ToUint64(v any) (uint64, bool) {
	switch val := v.(type) {
	case uint:
		return uint64(val), true
	case uint8:
		return uint64(val), true
	case uint16:
		return uint64(val), true
	case uint32:
		return uint64(val), true
	case uint64:
		return val, true
	default:
		return 0, false
	}
}

func ToFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

func Compare(res gjson.Result, expected any) (bool, string) {
	if expected == nil {
		if !res.Exists() {
			return false, "field does not exist (expected null)"
		}
		if res.Type != gjson.Null {
			return false, fmt.Sprintf("expected null, got %s: %s", TypeToString(res.Type), DebugValue(res))
		}
		return true, ""
	}

	switch exp := expected.(type) {
	case string:
		if res.Type != gjson.String {
			return false, fmt.Sprintf("expected string %q, got %s: %s", exp, TypeToString(res.Type), DebugValue(res))
		}
		actual := res.String()
		if actual != exp {
			return false, fmt.Sprintf("expected %q, got %q", exp, actual)
		}
		return true, ""

	case bool:
		if res.Type != gjson.True && res.Type != gjson.False {
			return false, fmt.Sprintf("expected boolean %v, got %s: %s", exp, TypeToString(res.Type), DebugValue(res))
		}
		actual := res.Bool()
		if actual != exp {
			return false, fmt.Sprintf("expected %v, got %v", exp, actual)
		}
		return true, ""

	default:
		if expectedInt, ok := ToInt64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedInt, TypeToString(res.Type), DebugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedInt, actualFloat)
			}
			actualInt := res.Int()
			if actualInt != expectedInt {
				return false, fmt.Sprintf("expected %d, got %d", expectedInt, actualInt)
			}
			return true, ""
		}

		if expectedUint, ok := ToUint64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedUint, TypeToString(res.Type), DebugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedUint, actualFloat)
			}
			actualUint := res.Uint()
			if actualUint != expectedUint {
				return false, fmt.Sprintf("expected %d, got %d", expectedUint, actualUint)
			}
			return true, ""
		}

		if expectedFloat, ok := ToFloat64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %v, got %s: %s", expectedFloat, TypeToString(res.Type), DebugValue(res))
			}
			actualFloat := res.Float()
			if actualFloat != expectedFloat {
				return false, fmt.Sprintf("expected %v, got %v", expectedFloat, actualFloat)
			}
			return true, ""
		}

		return false, fmt.Sprintf("unsupported expected type %T; supported: string/bool/int*/uint*/float*/nil", expected)
	}
}
