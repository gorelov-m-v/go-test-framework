package dsl

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/gorelov-m-v/go-test-framework/internal/typeconv"
)

func equalsLoose(expected, actual any) (bool, bool, string) {
	if expected == nil && actual == nil {
		return true, true, ""
	}
	if expected == nil || actual == nil {
		if typeconv.IsNull(actual) && expected == nil {
			return true, true, ""
		}
		return false, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	if typeconv.IsNull(actual) {
		return false, true, "column is NULL yet"
	}

	expBool, expIsBool := typeconv.ToBool(expected)
	actBool, actIsBool := typeconv.ToBool(actual)
	if expIsBool && actIsBool {
		if expBool == actBool {
			return true, true, ""
		}
		return false, true, fmt.Sprintf("expected %v, got %v", expBool, actBool)
	}

	expNum, expIsNum := toComparableNumber(expected)
	actNum, actIsNum := toComparableNumber(actual)
	if expIsNum && actIsNum {
		equal := expNum == actNum
		return equal, true, fmt.Sprintf("expected %v, got %v", expNum, actNum)
	}

	if expIsNum != actIsNum {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	expStr, expIsStr := toComparableString(expected)
	actStr, actIsStr := toComparableString(actual)
	if expIsStr && actIsStr {
		equal := expStr == actStr
		return equal, true, fmt.Sprintf("expected %v, got %v", expStr, actStr)
	}

	if expIsStr != actIsStr {
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
	case *string:
		if x != nil {
			return *x, true
		}
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
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr && !rv.IsNil() {
			elem := rv.Elem()
			if elem.Kind() == reflect.String {
				return elem.String(), true
			}
		}
	}
	return "", false
}
