package dsl

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
)

func equalsLoose(expected, actual any) (bool, bool, string) {
	if expected == nil && actual == nil {
		return true, true, ""
	}
	if expected == nil || actual == nil {
		if isValueNull(actual) && expected == nil {
			return true, true, ""
		}
		return false, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	if isValueNull(actual) {
		return false, true, "column is NULL yet"
	}

	expNum, expIsNum := toComparableNumber(expected)
	actNum, actIsNum := toComparableNumber(actual)
	if expIsNum && actIsNum {
		equal := expNum == actNum
		return equal, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	if expIsNum != actIsNum {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	expStr, expIsStr := toComparableString(expected)
	actStr, actIsStr := toComparableString(actual)
	if expIsStr && actIsStr {
		equal := expStr == actStr
		return equal, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	if expIsStr != actIsStr {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	expBool, expIsBool := asBool(expected)
	actBool, actIsBool := asBool(actual)
	if expIsBool && actIsBool {
		equal := expBool == actBool
		return equal, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	if expIsBool != actIsBool {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	if expVal.Type() != actVal.Type() {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	if !expVal.Type().Comparable() {
		equal := reflect.DeepEqual(expected, actual)
		return equal, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	equal := expected == actual
	return equal, true, fmt.Sprintf("expected %v, got %v", expected, actual)
}

func toComparableNumber(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint8:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	case float32:
		return float64(x), true
	case float64:
		return x, true
	case sql.NullInt64:
		if x.Valid {
			return float64(x.Int64), true
		}
	case *sql.NullInt64:
		if x != nil && x.Valid {
			return float64(x.Int64), true
		}
	case sql.NullInt32:
		if x.Valid {
			return float64(x.Int32), true
		}
	case *sql.NullInt32:
		if x != nil && x.Valid {
			return float64(x.Int32), true
		}
	case sql.NullInt16:
		if x.Valid {
			return float64(x.Int16), true
		}
	case *sql.NullInt16:
		if x != nil && x.Valid {
			return float64(x.Int16), true
		}
	case sql.NullByte:
		if x.Valid {
			return float64(x.Byte), true
		}
	case *sql.NullByte:
		if x != nil && x.Valid {
			return float64(x.Byte), true
		}
	case sql.NullFloat64:
		if x.Valid {
			return x.Float64, true
		}
	case *sql.NullFloat64:
		if x != nil && x.Valid {
			return x.Float64, true
		}
	}
	return 0, false
}

func toComparableString(v any) (string, bool) {
	switch x := v.(type) {
	case string:
		return x, true
	case []byte:
		return string(x), true
	case sql.NullString:
		if x.Valid {
			return x.String, true
		}
	case *sql.NullString:
		if x != nil && x.Valid {
			return x.String, true
		}
	}
	return "", false
}

func equalsBytesOrString(a, b any) bool {
	aBytes, aOk := a.([]byte)
	bBytes, bOk := b.([]byte)

	if aOk && bOk {
		return bytes.Equal(aBytes, bBytes)
	}

	if aOk {
		if bStr, ok := b.(string); ok {
			return string(aBytes) == bStr
		}
	}

	if bOk {
		if aStr, ok := a.(string); ok {
			return aStr == string(bBytes)
		}
	}

	return false
}
