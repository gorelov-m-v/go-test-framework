package typeconv

import (
	"database/sql"
	"testing"
)

func TestToBool(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    bool
		wantOK  bool
	}{
		{"bool true", true, true, true},
		{"bool false", false, false, true},

		{"int 0", int(0), false, true},
		{"int 1", int(1), true, true},
		{"int 2", int(2), false, false},
		{"int -1", int(-1), false, false},

		{"int8 0", int8(0), false, true},
		{"int8 1", int8(1), true, true},
		{"int8 2", int8(2), false, false},

		{"int16 0", int16(0), false, true},
		{"int16 1", int16(1), true, true},
		{"int16 2", int16(2), false, false},

		{"int32 0", int32(0), false, true},
		{"int32 1", int32(1), true, true},
		{"int32 2", int32(2), false, false},

		{"int64 0", int64(0), false, true},
		{"int64 1", int64(1), true, true},
		{"int64 2", int64(2), false, false},

		{"uint 0", uint(0), false, true},
		{"uint 1", uint(1), true, true},
		{"uint 2", uint(2), false, false},

		{"uint8 0", uint8(0), false, true},
		{"uint8 1", uint8(1), true, true},
		{"uint8 2", uint8(2), false, false},

		{"uint16 0", uint16(0), false, true},
		{"uint16 1", uint16(1), true, true},
		{"uint16 2", uint16(2), false, false},

		{"uint32 0", uint32(0), false, true},
		{"uint32 1", uint32(1), true, true},
		{"uint32 2", uint32(2), false, false},

		{"uint64 0", uint64(0), false, true},
		{"uint64 1", uint64(1), true, true},
		{"uint64 2", uint64(2), false, false},

		{"sql.NullBool valid true", sql.NullBool{Valid: true, Bool: true}, true, true},
		{"sql.NullBool valid false", sql.NullBool{Valid: true, Bool: false}, false, true},
		{"sql.NullBool invalid", sql.NullBool{Valid: false, Bool: true}, false, false},

		{"*sql.NullBool nil", (*sql.NullBool)(nil), false, false},
		{"*sql.NullBool valid true", &sql.NullBool{Valid: true, Bool: true}, true, true},
		{"*sql.NullBool valid false", &sql.NullBool{Valid: true, Bool: false}, false, true},
		{"*sql.NullBool invalid", &sql.NullBool{Valid: false}, false, false},

		{"sql.NullInt64 valid 0", sql.NullInt64{Valid: true, Int64: 0}, false, true},
		{"sql.NullInt64 valid 1", sql.NullInt64{Valid: true, Int64: 1}, true, true},
		{"sql.NullInt64 valid 2", sql.NullInt64{Valid: true, Int64: 2}, false, false},
		{"sql.NullInt64 invalid", sql.NullInt64{Valid: false}, false, false},

		{"*sql.NullInt64 nil", (*sql.NullInt64)(nil), false, false},
		{"*sql.NullInt64 valid 0", &sql.NullInt64{Valid: true, Int64: 0}, false, true},
		{"*sql.NullInt64 valid 1", &sql.NullInt64{Valid: true, Int64: 1}, true, true},
		{"*sql.NullInt64 valid 2", &sql.NullInt64{Valid: true, Int64: 2}, false, false},
		{"*sql.NullInt64 invalid", &sql.NullInt64{Valid: false}, false, false},

		{"sql.NullInt32 valid 0", sql.NullInt32{Valid: true, Int32: 0}, false, true},
		{"sql.NullInt32 valid 1", sql.NullInt32{Valid: true, Int32: 1}, true, true},
		{"sql.NullInt32 valid 2", sql.NullInt32{Valid: true, Int32: 2}, false, false},
		{"sql.NullInt32 invalid", sql.NullInt32{Valid: false}, false, false},

		{"*sql.NullInt32 nil", (*sql.NullInt32)(nil), false, false},
		{"*sql.NullInt32 valid 0", &sql.NullInt32{Valid: true, Int32: 0}, false, true},
		{"*sql.NullInt32 valid 1", &sql.NullInt32{Valid: true, Int32: 1}, true, true},
		{"*sql.NullInt32 invalid", &sql.NullInt32{Valid: false}, false, false},

		{"sql.NullInt16 valid 0", sql.NullInt16{Valid: true, Int16: 0}, false, true},
		{"sql.NullInt16 valid 1", sql.NullInt16{Valid: true, Int16: 1}, true, true},
		{"sql.NullInt16 valid 2", sql.NullInt16{Valid: true, Int16: 2}, false, false},
		{"sql.NullInt16 invalid", sql.NullInt16{Valid: false}, false, false},

		{"*sql.NullInt16 nil", (*sql.NullInt16)(nil), false, false},
		{"*sql.NullInt16 valid 0", &sql.NullInt16{Valid: true, Int16: 0}, false, true},
		{"*sql.NullInt16 valid 1", &sql.NullInt16{Valid: true, Int16: 1}, true, true},
		{"*sql.NullInt16 invalid", &sql.NullInt16{Valid: false}, false, false},

		{"sql.NullByte valid 0", sql.NullByte{Valid: true, Byte: 0}, false, true},
		{"sql.NullByte valid 1", sql.NullByte{Valid: true, Byte: 1}, true, true},
		{"sql.NullByte valid 2", sql.NullByte{Valid: true, Byte: 2}, false, false},
		{"sql.NullByte invalid", sql.NullByte{Valid: false}, false, false},

		{"*sql.NullByte nil", (*sql.NullByte)(nil), false, false},
		{"*sql.NullByte valid 0", &sql.NullByte{Valid: true, Byte: 0}, false, true},
		{"*sql.NullByte valid 1", &sql.NullByte{Valid: true, Byte: 1}, true, true},
		{"*sql.NullByte invalid", &sql.NullByte{Valid: false}, false, false},

		{"unsupported string", "true", false, false},
		{"unsupported float64", float64(1), false, false},
		{"unsupported nil", nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOK := ToBool(tt.input)
			if got != tt.want || gotOK != tt.wantOK {
				t.Errorf("ToBool(%v) = (%v, %v), want (%v, %v)", tt.input, got, gotOK, tt.want, tt.wantOK)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{"nil", nil, true},

		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"non-empty string", "hello", false},

		{"*string nil", (*string)(nil), true},
		{"*string empty", ptrString(""), true},
		{"*string whitespace", ptrString("  "), true},
		{"*string non-empty", ptrString("hello"), false},

		{"sql.NullString invalid", sql.NullString{Valid: false}, true},
		{"sql.NullString valid empty", sql.NullString{Valid: true, String: ""}, true},
		{"sql.NullString valid whitespace", sql.NullString{Valid: true, String: "  "}, true},
		{"sql.NullString valid non-empty", sql.NullString{Valid: true, String: "hello"}, false},

		{"*sql.NullString nil", (*sql.NullString)(nil), true},
		{"*sql.NullString invalid", &sql.NullString{Valid: false}, true},
		{"*sql.NullString valid empty", &sql.NullString{Valid: true, String: ""}, true},
		{"*sql.NullString valid non-empty", &sql.NullString{Valid: true, String: "hello"}, false},

		{"sql.NullInt64 invalid", sql.NullInt64{Valid: false}, true},
		{"sql.NullInt64 valid", sql.NullInt64{Valid: true, Int64: 0}, false},
		{"*sql.NullInt64 nil", (*sql.NullInt64)(nil), true},
		{"*sql.NullInt64 invalid", &sql.NullInt64{Valid: false}, true},
		{"*sql.NullInt64 valid", &sql.NullInt64{Valid: true, Int64: 42}, false},

		{"sql.NullInt32 invalid", sql.NullInt32{Valid: false}, true},
		{"sql.NullInt32 valid", sql.NullInt32{Valid: true, Int32: 0}, false},
		{"*sql.NullInt32 nil", (*sql.NullInt32)(nil), true},
		{"*sql.NullInt32 invalid", &sql.NullInt32{Valid: false}, true},
		{"*sql.NullInt32 valid", &sql.NullInt32{Valid: true, Int32: 42}, false},

		{"sql.NullInt16 invalid", sql.NullInt16{Valid: false}, true},
		{"sql.NullInt16 valid", sql.NullInt16{Valid: true, Int16: 0}, false},
		{"*sql.NullInt16 nil", (*sql.NullInt16)(nil), true},
		{"*sql.NullInt16 invalid", &sql.NullInt16{Valid: false}, true},
		{"*sql.NullInt16 valid", &sql.NullInt16{Valid: true, Int16: 42}, false},

		{"sql.NullByte invalid", sql.NullByte{Valid: false}, true},
		{"sql.NullByte valid", sql.NullByte{Valid: true, Byte: 0}, false},
		{"*sql.NullByte nil", (*sql.NullByte)(nil), true},
		{"*sql.NullByte invalid", &sql.NullByte{Valid: false}, true},
		{"*sql.NullByte valid", &sql.NullByte{Valid: true, Byte: 42}, false},

		{"sql.NullFloat64 invalid", sql.NullFloat64{Valid: false}, true},
		{"sql.NullFloat64 valid", sql.NullFloat64{Valid: true, Float64: 0}, false},
		{"*sql.NullFloat64 nil", (*sql.NullFloat64)(nil), true},
		{"*sql.NullFloat64 invalid", &sql.NullFloat64{Valid: false}, true},
		{"*sql.NullFloat64 valid", &sql.NullFloat64{Valid: true, Float64: 3.14}, false},

		{"sql.NullBool invalid", sql.NullBool{Valid: false}, true},
		{"sql.NullBool valid", sql.NullBool{Valid: true, Bool: false}, false},
		{"*sql.NullBool nil", (*sql.NullBool)(nil), true},
		{"*sql.NullBool invalid", &sql.NullBool{Valid: false}, true},
		{"*sql.NullBool valid", &sql.NullBool{Valid: true, Bool: true}, false},

		{"sql.NullTime invalid", sql.NullTime{Valid: false}, true},
		{"*sql.NullTime nil", (*sql.NullTime)(nil), true},
		{"*sql.NullTime invalid", &sql.NullTime{Valid: false}, true},

		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1, 2, 3}, false},
		{"nil slice", ([]int)(nil), true},

		{"empty map", map[string]int{}, true},
		{"non-empty map", map[string]int{"a": 1}, false},
		{"nil map", (map[string]int)(nil), true},

		{"empty array", [0]int{}, true},
		{"non-empty array", [3]int{1, 2, 3}, false},

		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", ptrInt(42), false},

		{"int zero", 0, false},
		{"int non-zero", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmpty(tt.input)
			if got != tt.want {
				t.Errorf("IsEmpty(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsNull(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  bool
	}{
		{"nil", nil, true},

		{"sql.NullString invalid", sql.NullString{Valid: false}, true},
		{"sql.NullString valid", sql.NullString{Valid: true, String: ""}, false},
		{"*sql.NullString nil", (*sql.NullString)(nil), true},
		{"*sql.NullString invalid", &sql.NullString{Valid: false}, true},
		{"*sql.NullString valid", &sql.NullString{Valid: true, String: "hello"}, false},

		{"sql.NullInt64 invalid", sql.NullInt64{Valid: false}, true},
		{"sql.NullInt64 valid", sql.NullInt64{Valid: true, Int64: 0}, false},
		{"*sql.NullInt64 nil", (*sql.NullInt64)(nil), true},
		{"*sql.NullInt64 invalid", &sql.NullInt64{Valid: false}, true},
		{"*sql.NullInt64 valid", &sql.NullInt64{Valid: true, Int64: 42}, false},

		{"sql.NullInt32 invalid", sql.NullInt32{Valid: false}, true},
		{"sql.NullInt32 valid", sql.NullInt32{Valid: true, Int32: 0}, false},
		{"*sql.NullInt32 nil", (*sql.NullInt32)(nil), true},
		{"*sql.NullInt32 invalid", &sql.NullInt32{Valid: false}, true},
		{"*sql.NullInt32 valid", &sql.NullInt32{Valid: true, Int32: 42}, false},

		{"sql.NullInt16 invalid", sql.NullInt16{Valid: false}, true},
		{"sql.NullInt16 valid", sql.NullInt16{Valid: true, Int16: 0}, false},
		{"*sql.NullInt16 nil", (*sql.NullInt16)(nil), true},
		{"*sql.NullInt16 invalid", &sql.NullInt16{Valid: false}, true},
		{"*sql.NullInt16 valid", &sql.NullInt16{Valid: true, Int16: 42}, false},

		{"sql.NullByte invalid", sql.NullByte{Valid: false}, true},
		{"sql.NullByte valid", sql.NullByte{Valid: true, Byte: 0}, false},
		{"*sql.NullByte nil", (*sql.NullByte)(nil), true},
		{"*sql.NullByte invalid", &sql.NullByte{Valid: false}, true},
		{"*sql.NullByte valid", &sql.NullByte{Valid: true, Byte: 42}, false},

		{"sql.NullFloat64 invalid", sql.NullFloat64{Valid: false}, true},
		{"sql.NullFloat64 valid", sql.NullFloat64{Valid: true, Float64: 0}, false},
		{"*sql.NullFloat64 nil", (*sql.NullFloat64)(nil), true},
		{"*sql.NullFloat64 invalid", &sql.NullFloat64{Valid: false}, true},
		{"*sql.NullFloat64 valid", &sql.NullFloat64{Valid: true, Float64: 3.14}, false},

		{"sql.NullBool invalid", sql.NullBool{Valid: false}, true},
		{"sql.NullBool valid", sql.NullBool{Valid: true, Bool: false}, false},
		{"*sql.NullBool nil", (*sql.NullBool)(nil), true},
		{"*sql.NullBool invalid", &sql.NullBool{Valid: false}, true},
		{"*sql.NullBool valid", &sql.NullBool{Valid: true, Bool: true}, false},

		{"sql.NullTime invalid", sql.NullTime{Valid: false}, true},
		{"*sql.NullTime nil", (*sql.NullTime)(nil), true},
		{"*sql.NullTime invalid", &sql.NullTime{Valid: false}, true},

		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", ptrInt(42), false},

		{"nil slice", ([]int)(nil), true},
		{"empty slice", []int{}, false},
		{"non-empty slice", []int{1, 2, 3}, false},

		{"nil map", (map[string]int)(nil), true},
		{"empty map", map[string]int{}, false},
		{"non-empty map", map[string]int{"a": 1}, false},

		{"nil func", (func())(nil), true},

		{"nil chan", (chan int)(nil), true},
		{"non-nil chan", make(chan int), false},

		{"int", 42, false},
		{"string", "hello", false},
		{"bool", true, false},
		{"struct", struct{ X int }{X: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNull(tt.input)
			if got != tt.want {
				t.Errorf("IsNull(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}
