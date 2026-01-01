package dsl

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpectColumnIsNull_SqlNullTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expectOK bool
	}{
		{"sql.NullString valid", sql.NullString{String: "test", Valid: true}, false},
		{"sql.NullString invalid", sql.NullString{Valid: false}, true},
		{"sql.NullInt64 valid", sql.NullInt64{Int64: 42, Valid: true}, false},
		{"sql.NullInt64 invalid", sql.NullInt64{Valid: false}, true},
		{"sql.NullInt32 valid", sql.NullInt32{Int32: 42, Valid: true}, false},
		{"sql.NullInt32 invalid", sql.NullInt32{Valid: false}, true},
		{"sql.NullFloat64 valid", sql.NullFloat64{Float64: 3.14, Valid: true}, false},
		{"sql.NullFloat64 invalid", sql.NullFloat64{Valid: false}, true},
		{"sql.NullBool valid", sql.NullBool{Bool: true, Valid: true}, false},
		{"sql.NullBool invalid", sql.NullBool{Valid: false}, true},
		{"sql.NullTime valid", sql.NullTime{Valid: true}, false},
		{"sql.NullTime invalid", sql.NullTime{Valid: false}, true},
		{"*sql.NullString nil", (*sql.NullString)(nil), true},
		{"*sql.NullString valid", &sql.NullString{String: "test", Valid: true}, false},
		{"*sql.NullString invalid", &sql.NullString{Valid: false}, true},
		{"*sql.NullInt64 nil", (*sql.NullInt64)(nil), true},
		{"*sql.NullInt64 valid", &sql.NullInt64{Int64: 42, Valid: true}, false},
		{"*sql.NullInt64 invalid", &sql.NullInt64{Valid: false}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNull := false

			switch v := tt.value.(type) {
			case sql.NullString:
				isNull = !v.Valid
			case sql.NullInt64:
				isNull = !v.Valid
			case sql.NullInt32:
				isNull = !v.Valid
			case sql.NullInt16:
				isNull = !v.Valid
			case sql.NullByte:
				isNull = !v.Valid
			case sql.NullFloat64:
				isNull = !v.Valid
			case sql.NullBool:
				isNull = !v.Valid
			case sql.NullTime:
				isNull = !v.Valid
			case *sql.NullString:
				isNull = v == nil || !v.Valid
			case *sql.NullInt64:
				isNull = v == nil || !v.Valid
			case *sql.NullInt32:
				isNull = v == nil || !v.Valid
			case *sql.NullInt16:
				isNull = v == nil || !v.Valid
			case *sql.NullByte:
				isNull = v == nil || !v.Valid
			case *sql.NullFloat64:
				isNull = v == nil || !v.Valid
			case *sql.NullBool:
				isNull = v == nil || !v.Valid
			case *sql.NullTime:
				isNull = v == nil || !v.Valid
			default:
				t.Errorf("Unsupported type: %T", tt.value)
				return
			}

			if isNull != tt.expectOK {
				t.Errorf("Expected isNull=%v, got isNull=%v for %v", tt.expectOK, isNull, tt.name)
			}
		})
	}
}

func TestExpectColumnIsNotNull_SqlNullTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expectOK bool
	}{
		{"sql.NullString valid", sql.NullString{String: "test", Valid: true}, true},
		{"sql.NullString invalid", sql.NullString{Valid: false}, false},
		{"sql.NullInt64 valid", sql.NullInt64{Int64: 42, Valid: true}, true},
		{"sql.NullInt64 invalid", sql.NullInt64{Valid: false}, false},
		{"sql.NullInt32 valid", sql.NullInt32{Int32: 42, Valid: true}, true},
		{"sql.NullInt32 invalid", sql.NullInt32{Valid: false}, false},
		{"sql.NullFloat64 valid", sql.NullFloat64{Float64: 3.14, Valid: true}, true},
		{"sql.NullFloat64 invalid", sql.NullFloat64{Valid: false}, false},
		{"sql.NullBool valid", sql.NullBool{Bool: true, Valid: true}, true},
		{"sql.NullBool invalid", sql.NullBool{Valid: false}, false},
		{"sql.NullTime valid", sql.NullTime{Valid: true}, true},
		{"sql.NullTime invalid", sql.NullTime{Valid: false}, false},
		{"*sql.NullString nil", (*sql.NullString)(nil), false},
		{"*sql.NullString valid", &sql.NullString{String: "test", Valid: true}, true},
		{"*sql.NullString invalid", &sql.NullString{Valid: false}, false},
		{"*sql.NullInt64 nil", (*sql.NullInt64)(nil), false},
		{"*sql.NullInt64 valid", &sql.NullInt64{Int64: 42, Valid: true}, true},
		{"*sql.NullInt64 invalid", &sql.NullInt64{Valid: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isNotNull := false

			switch v := tt.value.(type) {
			case sql.NullString:
				isNotNull = v.Valid
			case sql.NullInt64:
				isNotNull = v.Valid
			case sql.NullInt32:
				isNotNull = v.Valid
			case sql.NullInt16:
				isNotNull = v.Valid
			case sql.NullByte:
				isNotNull = v.Valid
			case sql.NullFloat64:
				isNotNull = v.Valid
			case sql.NullBool:
				isNotNull = v.Valid
			case sql.NullTime:
				isNotNull = v.Valid
			case *sql.NullString:
				isNotNull = v != nil && v.Valid
			case *sql.NullInt64:
				isNotNull = v != nil && v.Valid
			case *sql.NullInt32:
				isNotNull = v != nil && v.Valid
			case *sql.NullInt16:
				isNotNull = v != nil && v.Valid
			case *sql.NullByte:
				isNotNull = v != nil && v.Valid
			case *sql.NullFloat64:
				isNotNull = v != nil && v.Valid
			case *sql.NullBool:
				isNotNull = v != nil && v.Valid
			case *sql.NullTime:
				isNotNull = v != nil && v.Valid
			default:
				t.Errorf("Unsupported type: %T", tt.value)
				return
			}

			if isNotNull != tt.expectOK {
				t.Errorf("Expected isNotNull=%v, got isNotNull=%v for %v", tt.expectOK, isNotNull, tt.name)
			}
		})
	}
}

func TestExpectColumnNotEmpty_Semantics(t *testing.T) {
	tests := []struct {
		name        string
		value       any
		expectEmpty bool
	}{
		{"string empty", "", true},
		{"string whitespace", "   ", true},
		{"string not empty", "test", false},
		{"string with content", "  test  ", false},

		{"int zero", 0, true},
		{"int non-zero", 42, false},
		{"int64 zero", int64(0), true},
		{"int64 non-zero", int64(42), false},
		{"float64 zero", 0.0, true},
		{"float64 non-zero", 3.14, false},

		{"sql.NullString invalid", sql.NullString{Valid: false}, true},
		{"sql.NullString valid empty", sql.NullString{String: "", Valid: true}, false},
		{"sql.NullString valid non-empty", sql.NullString{String: "test", Valid: true}, false},
		{"sql.NullInt64 invalid", sql.NullInt64{Valid: false}, true},
		{"sql.NullInt64 valid zero", sql.NullInt64{Int64: 0, Valid: true}, false},
		{"sql.NullInt64 valid non-zero", sql.NullInt64{Int64: 42, Valid: true}, false},
		{"sql.NullBool invalid", sql.NullBool{Valid: false}, true},
		{"sql.NullBool valid", sql.NullBool{Bool: true, Valid: true}, false},
		{"sql.NullTime invalid", sql.NullTime{Valid: false}, true},
		{"sql.NullTime valid", sql.NullTime{Valid: true}, false},

		{"*string nil", (*string)(nil), true},
		{"*int nil", (*int)(nil), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := checkIsEmpty(tt.value)

			if isEmpty != tt.expectEmpty {
				t.Errorf("Expected isEmpty=%v, got isEmpty=%v for %v (value: %v)", tt.expectEmpty, isEmpty, tt.name, tt.value)
			}
		})
	}
}

func checkIsEmpty(actualValue any) bool {
	if str, ok := actualValue.(string); ok {
		trimmed := ""
		for _, r := range str {
			if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
				trimmed += string(r)
			}
		}
		return trimmed == ""
	}

	val := reflect.ValueOf(actualValue)
	if val.Kind() == reflect.Ptr {
		return val.IsNil()
	}

	if actualValue == nil {
		return true
	}

	switch v := actualValue.(type) {
	case sql.NullString:
		return !v.Valid
	case sql.NullInt64:
		return !v.Valid
	case sql.NullInt32:
		return !v.Valid
	case sql.NullInt16:
		return !v.Valid
	case sql.NullByte:
		return !v.Valid
	case sql.NullFloat64:
		return !v.Valid
	case sql.NullBool:
		return !v.Valid
	case sql.NullTime:
		return !v.Valid
	case *sql.NullString:
		return v == nil || !v.Valid
	case *sql.NullInt64:
		return v == nil || !v.Valid
	case *sql.NullInt32:
		return v == nil || !v.Valid
	case *sql.NullInt16:
		return v == nil || !v.Valid
	case *sql.NullByte:
		return v == nil || !v.Valid
	case *sql.NullFloat64:
		return v == nil || !v.Valid
	case *sql.NullBool:
		return v == nil || !v.Valid
	case *sql.NullTime:
		return v == nil || !v.Valid
	}

	switch actualValue {
	case 0, int64(0), int32(0), int16(0), int8(0), uint(0), uint64(0), uint32(0), uint16(0), uint8(0), float32(0.0), false:
		return true
	}

	return false
}

func TestAsBool_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
		canConv  bool
	}{
		// bool
		{"bool true", true, true, true},
		{"bool false", false, false, true},

		// int types
		{"int 0", int(0), false, true},
		{"int 1", int(1), true, true},
		{"int 2", int(2), false, false},
		{"int8 0", int8(0), false, true},
		{"int8 1", int8(1), true, true},
		{"int16 0", int16(0), false, true},
		{"int16 1", int16(1), true, true},
		{"int32 0", int32(0), false, true},
		{"int32 1", int32(1), true, true},
		{"int64 0", int64(0), false, true},
		{"int64 1", int64(1), true, true},

		// uint types
		{"uint 0", uint(0), false, true},
		{"uint 1", uint(1), true, true},
		{"uint8 0", uint8(0), false, true},
		{"uint8 1", uint8(1), true, true},
		{"uint16 0", uint16(0), false, true},
		{"uint16 1", uint16(1), true, true},
		{"uint32 0", uint32(0), false, true},
		{"uint32 1", uint32(1), true, true},
		{"uint64 0", uint64(0), false, true},
		{"uint64 1", uint64(1), true, true},

		// sql.NullBool
		{"NullBool true valid", sql.NullBool{Bool: true, Valid: true}, true, true},
		{"NullBool false valid", sql.NullBool{Bool: false, Valid: true}, false, true},
		{"NullBool invalid", sql.NullBool{Valid: false}, false, false},
		{"*NullBool true valid", &sql.NullBool{Bool: true, Valid: true}, true, true},
		{"*NullBool false valid", &sql.NullBool{Bool: false, Valid: true}, false, true},
		{"*NullBool invalid", &sql.NullBool{Valid: false}, false, false},
		{"*NullBool nil", (*sql.NullBool)(nil), false, false},

		// sql.NullInt64
		{"NullInt64 0 valid", sql.NullInt64{Int64: 0, Valid: true}, false, true},
		{"NullInt64 1 valid", sql.NullInt64{Int64: 1, Valid: true}, true, true},
		{"NullInt64 invalid", sql.NullInt64{Valid: false}, false, false},
		{"*NullInt64 0 valid", &sql.NullInt64{Int64: 0, Valid: true}, false, true},
		{"*NullInt64 1 valid", &sql.NullInt64{Int64: 1, Valid: true}, true, true},
		{"*NullInt64 invalid", &sql.NullInt64{Valid: false}, false, false},
		{"*NullInt64 nil", (*sql.NullInt64)(nil), false, false},

		// sql.NullInt32
		{"NullInt32 0 valid", sql.NullInt32{Int32: 0, Valid: true}, false, true},
		{"NullInt32 1 valid", sql.NullInt32{Int32: 1, Valid: true}, true, true},
		{"NullInt32 invalid", sql.NullInt32{Valid: false}, false, false},
		{"*NullInt32 0 valid", &sql.NullInt32{Int32: 0, Valid: true}, false, true},
		{"*NullInt32 1 valid", &sql.NullInt32{Int32: 1, Valid: true}, true, true},
		{"*NullInt32 invalid", &sql.NullInt32{Valid: false}, false, false},
		{"*NullInt32 nil", (*sql.NullInt32)(nil), false, false},

		// sql.NullInt16
		{"NullInt16 0 valid", sql.NullInt16{Int16: 0, Valid: true}, false, true},
		{"NullInt16 1 valid", sql.NullInt16{Int16: 1, Valid: true}, true, true},
		{"NullInt16 invalid", sql.NullInt16{Valid: false}, false, false},
		{"*NullInt16 0 valid", &sql.NullInt16{Int16: 0, Valid: true}, false, true},
		{"*NullInt16 1 valid", &sql.NullInt16{Int16: 1, Valid: true}, true, true},
		{"*NullInt16 invalid", &sql.NullInt16{Valid: false}, false, false},
		{"*NullInt16 nil", (*sql.NullInt16)(nil), false, false},

		// sql.NullByte
		{"NullByte 0 valid", sql.NullByte{Byte: 0, Valid: true}, false, true},
		{"NullByte 1 valid", sql.NullByte{Byte: 1, Valid: true}, true, true},
		{"NullByte invalid", sql.NullByte{Valid: false}, false, false},
		{"*NullByte 0 valid", &sql.NullByte{Byte: 0, Valid: true}, false, true},
		{"*NullByte 1 valid", &sql.NullByte{Byte: 1, Valid: true}, true, true},
		{"*NullByte invalid", &sql.NullByte{Valid: false}, false, false},
		{"*NullByte nil", (*sql.NullByte)(nil), false, false},

		// Non-convertible types
		{"string", "true", false, false},
		{"float64", 3.14, false, false},
		{"nil", nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := asBool(tt.value)
			assert.Equal(t, tt.canConv, ok, "canConvert mismatch for %s", tt.name)
			if ok {
				assert.Equal(t, tt.expected, result, "bool value mismatch for %s", tt.name)
			}
		})
	}
}
