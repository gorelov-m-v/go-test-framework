package typeconv

import (
	"database/sql"
	"reflect"
	"strings"
)

func ToBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case int:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int8:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int16:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int32:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int64:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint8:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint16:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint32:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint64:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case sql.NullBool:
		if !x.Valid {
			return false, false
		}
		return x.Bool, true
	case *sql.NullBool:
		if x == nil || !x.Valid {
			return false, false
		}
		return x.Bool, true
	case sql.NullInt64:
		if !x.Valid {
			return false, false
		}
		if x.Int64 == 0 {
			return false, true
		} else if x.Int64 == 1 {
			return true, true
		}
	case *sql.NullInt64:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Int64 == 0 {
			return false, true
		} else if x.Int64 == 1 {
			return true, true
		}
	case sql.NullInt32:
		if !x.Valid {
			return false, false
		}
		if x.Int32 == 0 {
			return false, true
		} else if x.Int32 == 1 {
			return true, true
		}
	case *sql.NullInt32:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Int32 == 0 {
			return false, true
		} else if x.Int32 == 1 {
			return true, true
		}
	case sql.NullInt16:
		if !x.Valid {
			return false, false
		}
		if x.Int16 == 0 {
			return false, true
		} else if x.Int16 == 1 {
			return true, true
		}
	case *sql.NullInt16:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Int16 == 0 {
			return false, true
		} else if x.Int16 == 1 {
			return true, true
		}
	case sql.NullByte:
		if !x.Valid {
			return false, false
		}
		if x.Byte == 0 {
			return false, true
		} else if x.Byte == 1 {
			return true, true
		}
	case *sql.NullByte:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Byte == 0 {
			return false, true
		} else if x.Byte == 1 {
			return true, true
		}
	}
	return false, false
}

func IsEmpty(v any) bool {
	if v == nil {
		return true
	}

	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) == ""
	case *string:
		return x == nil || strings.TrimSpace(*x) == ""

	case sql.NullString:
		return !x.Valid || strings.TrimSpace(x.String) == ""
	case *sql.NullString:
		return x == nil || !x.Valid || strings.TrimSpace(x.String) == ""

	case sql.NullInt64:
		return !x.Valid
	case *sql.NullInt64:
		return x == nil || !x.Valid

	case sql.NullInt32:
		return !x.Valid
	case *sql.NullInt32:
		return x == nil || !x.Valid

	case sql.NullInt16:
		return !x.Valid
	case *sql.NullInt16:
		return x == nil || !x.Valid

	case sql.NullByte:
		return !x.Valid
	case *sql.NullByte:
		return x == nil || !x.Valid

	case sql.NullFloat64:
		return !x.Valid
	case *sql.NullFloat64:
		return x == nil || !x.Valid

	case sql.NullBool:
		return !x.Valid
	case *sql.NullBool:
		return x == nil || !x.Valid

	case sql.NullTime:
		return !x.Valid
	case *sql.NullTime:
		return x == nil || !x.Valid

	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			return rv.IsNil()
		}
		switch rv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() == 0
		case reflect.String:
			return strings.TrimSpace(rv.String()) == ""
		default:
			return false
		}
	}
}

func IsNull(v any) bool {
	if v == nil {
		return true
	}

	switch x := v.(type) {
	case sql.NullString:
		return !x.Valid
	case *sql.NullString:
		return x == nil || !x.Valid

	case sql.NullInt64:
		return !x.Valid
	case *sql.NullInt64:
		return x == nil || !x.Valid

	case sql.NullInt32:
		return !x.Valid
	case *sql.NullInt32:
		return x == nil || !x.Valid

	case sql.NullInt16:
		return !x.Valid
	case *sql.NullInt16:
		return x == nil || !x.Valid

	case sql.NullByte:
		return !x.Valid
	case *sql.NullByte:
		return x == nil || !x.Valid

	case sql.NullFloat64:
		return !x.Valid
	case *sql.NullFloat64:
		return x == nil || !x.Valid

	case sql.NullBool:
		return !x.Valid
	case *sql.NullBool:
		return x == nil || !x.Valid

	case sql.NullTime:
		return !x.Valid
	case *sql.NullTime:
		return x == nil || !x.Valid

	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
			return rv.IsNil()
		default:
			return false
		}
	}
}

func ToNumber(v any) (float64, bool) {
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

func ToString(v any) (string, bool) {
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
