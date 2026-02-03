package dsl

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorelov-m-v/go-test-framework/internal/typeconv"
)

func TestToNumber_Int(t *testing.T) {
	result, ok := typeconv.ToNumber(42)
	assert.True(t, ok)
	assert.Equal(t, float64(42), result)
}

func TestToNumber_Int8(t *testing.T) {
	result, ok := typeconv.ToNumber(int8(8))
	assert.True(t, ok)
	assert.Equal(t, float64(8), result)
}

func TestToNumber_Int16(t *testing.T) {
	result, ok := typeconv.ToNumber(int16(16))
	assert.True(t, ok)
	assert.Equal(t, float64(16), result)
}

func TestToNumber_Int32(t *testing.T) {
	result, ok := typeconv.ToNumber(int32(32))
	assert.True(t, ok)
	assert.Equal(t, float64(32), result)
}

func TestToNumber_Int64(t *testing.T) {
	result, ok := typeconv.ToNumber(int64(64))
	assert.True(t, ok)
	assert.Equal(t, float64(64), result)
}

func TestToNumber_Uint(t *testing.T) {
	result, ok := typeconv.ToNumber(uint(100))
	assert.True(t, ok)
	assert.Equal(t, float64(100), result)
}

func TestToNumber_Uint8(t *testing.T) {
	result, ok := typeconv.ToNumber(uint8(8))
	assert.True(t, ok)
	assert.Equal(t, float64(8), result)
}

func TestToNumber_Uint16(t *testing.T) {
	result, ok := typeconv.ToNumber(uint16(16))
	assert.True(t, ok)
	assert.Equal(t, float64(16), result)
}

func TestToNumber_Uint32(t *testing.T) {
	result, ok := typeconv.ToNumber(uint32(32))
	assert.True(t, ok)
	assert.Equal(t, float64(32), result)
}

func TestToNumber_Uint64(t *testing.T) {
	result, ok := typeconv.ToNumber(uint64(64))
	assert.True(t, ok)
	assert.Equal(t, float64(64), result)
}

func TestToNumber_Float32(t *testing.T) {
	result, ok := typeconv.ToNumber(float32(3.14))
	assert.True(t, ok)
	assert.InDelta(t, float64(3.14), result, 0.001)
}

func TestToNumber_Float64(t *testing.T) {
	result, ok := typeconv.ToNumber(float64(3.14159))
	assert.True(t, ok)
	assert.Equal(t, float64(3.14159), result)
}

func TestToNumber_SqlNullInt64_Valid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullInt64{Int64: 123, Valid: true})
	assert.True(t, ok)
	assert.Equal(t, float64(123), result)
}

func TestToNumber_SqlNullInt64_Invalid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullInt64{Int64: 123, Valid: false})
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullInt64_Pointer_Valid(t *testing.T) {
	val := &sql.NullInt64{Int64: 456, Valid: true}
	result, ok := typeconv.ToNumber(val)
	assert.True(t, ok)
	assert.Equal(t, float64(456), result)
}

func TestToNumber_SqlNullInt64_Pointer_Nil(t *testing.T) {
	var val *sql.NullInt64
	result, ok := typeconv.ToNumber(val)
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullInt32_Valid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullInt32{Int32: 32, Valid: true})
	assert.True(t, ok)
	assert.Equal(t, float64(32), result)
}

func TestToNumber_SqlNullInt16_Valid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullInt16{Int16: 16, Valid: true})
	assert.True(t, ok)
	assert.Equal(t, float64(16), result)
}

func TestToNumber_SqlNullFloat64_Valid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullFloat64{Float64: 99.99, Valid: true})
	assert.True(t, ok)
	assert.Equal(t, float64(99.99), result)
}

func TestToNumber_SqlNullFloat64_Invalid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullFloat64{Float64: 99.99, Valid: false})
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullByte_Valid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullByte{Byte: 255, Valid: true})
	assert.True(t, ok)
	assert.Equal(t, float64(255), result)
}

func TestToNumber_String(t *testing.T) {
	result, ok := typeconv.ToNumber("not a number")
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_Nil(t *testing.T) {
	result, ok := typeconv.ToNumber(nil)
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToString_String(t *testing.T) {
	result, ok := typeconv.ToString("hello")
	assert.True(t, ok)
	assert.Equal(t, "hello", result)
}

func TestToString_StringPointer(t *testing.T) {
	s := "world"
	result, ok := typeconv.ToString(&s)
	assert.True(t, ok)
	assert.Equal(t, "world", result)
}

func TestToString_StringPointer_Nil(t *testing.T) {
	var s *string
	result, ok := typeconv.ToString(s)
	assert.False(t, ok)
	assert.Equal(t, "", result)
}

func TestToString_ByteSlice(t *testing.T) {
	result, ok := typeconv.ToString([]byte("bytes"))
	assert.True(t, ok)
	assert.Equal(t, "bytes", result)
}

func TestToString_SqlNullString_Valid(t *testing.T) {
	result, ok := typeconv.ToString(sql.NullString{String: "valid", Valid: true})
	assert.True(t, ok)
	assert.Equal(t, "valid", result)
}

func TestToString_SqlNullString_Invalid(t *testing.T) {
	result, ok := typeconv.ToString(sql.NullString{String: "invalid", Valid: false})
	assert.False(t, ok)
	assert.Equal(t, "", result)
}

func TestToString_SqlNullString_Pointer_Valid(t *testing.T) {
	val := &sql.NullString{String: "pointer", Valid: true}
	result, ok := typeconv.ToString(val)
	assert.True(t, ok)
	assert.Equal(t, "pointer", result)
}

func TestToString_SqlNullString_Pointer_Nil(t *testing.T) {
	var val *sql.NullString
	result, ok := typeconv.ToString(val)
	assert.False(t, ok)
	assert.Equal(t, "", result)
}

func TestToString_Int(t *testing.T) {
	result, ok := typeconv.ToString(123)
	assert.False(t, ok)
	assert.Equal(t, "", result)
}

func TestEqualsLoose_BothNil(t *testing.T) {
	equal, _, _ := equalsLoose(nil, nil)
	assert.True(t, equal)
}

func TestEqualsLoose_ExpectedNil_ActualNotNil(t *testing.T) {
	equal, _, _ := equalsLoose(nil, "value")
	assert.False(t, equal)
}

func TestEqualsLoose_ExpectedNotNil_ActualNil(t *testing.T) {
	equal, _, _ := equalsLoose("value", nil)
	assert.False(t, equal)
}

func TestEqualsLoose_IntEqual(t *testing.T) {
	equal, _, _ := equalsLoose(42, 42)
	assert.True(t, equal)
}

func TestEqualsLoose_IntNotEqual(t *testing.T) {
	equal, _, _ := equalsLoose(42, 43)
	assert.False(t, equal)
}

func TestEqualsLoose_IntVsInt64(t *testing.T) {
	equal, _, _ := equalsLoose(42, int64(42))
	assert.True(t, equal)
}

func TestEqualsLoose_IntVsInt16(t *testing.T) {
	equal, _, _ := equalsLoose(42, int16(42))
	assert.True(t, equal)
}

func TestEqualsLoose_IntVsFloat64(t *testing.T) {
	equal, _, _ := equalsLoose(42, float64(42))
	assert.True(t, equal)
}

func TestEqualsLoose_Float64Equal(t *testing.T) {
	equal, _, _ := equalsLoose(3.14, 3.14)
	assert.True(t, equal)
}

func TestEqualsLoose_Float64NotEqual(t *testing.T) {
	equal, _, _ := equalsLoose(3.14, 3.15)
	assert.False(t, equal)
}

func TestEqualsLoose_StringEqual(t *testing.T) {
	equal, _, _ := equalsLoose("hello", "hello")
	assert.True(t, equal)
}

func TestEqualsLoose_StringNotEqual(t *testing.T) {
	equal, _, _ := equalsLoose("hello", "world")
	assert.False(t, equal)
}

func TestEqualsLoose_StringVsBytes(t *testing.T) {
	equal, _, _ := equalsLoose("test", []byte("test"))
	assert.True(t, equal)
}

func TestEqualsLoose_BoolEqual(t *testing.T) {
	equal, _, _ := equalsLoose(true, true)
	assert.True(t, equal)
}

func TestEqualsLoose_BoolNotEqual(t *testing.T) {
	equal, _, _ := equalsLoose(true, false)
	assert.False(t, equal)
}

func TestEqualsLoose_IntVsSqlNullInt64(t *testing.T) {
	equal, _, _ := equalsLoose(123, sql.NullInt64{Int64: 123, Valid: true})
	assert.True(t, equal)
}

func TestEqualsLoose_StringVsSqlNullString(t *testing.T) {
	equal, _, _ := equalsLoose("test", sql.NullString{String: "test", Valid: true})
	assert.True(t, equal)
}

func TestEqualsLoose_TypeMismatch_IntVsString(t *testing.T) {
	equal, retryable, _ := equalsLoose(123, "123")
	assert.False(t, equal)
	assert.False(t, retryable)
}

func TestEqualsLoose_SqlNullInt64_Nil(t *testing.T) {
	equal, _, _ := equalsLoose(nil, sql.NullInt64{Valid: false})
	assert.True(t, equal)
}

func TestEqualsLoose_SqlNullString_Nil(t *testing.T) {
	equal, _, _ := equalsLoose(nil, sql.NullString{Valid: false})
	assert.True(t, equal)
}

func TestEqualsLoose_SliceEqual(t *testing.T) {
	expected := []int{1, 2, 3}
	actual := []int{1, 2, 3}
	equal, _, _ := equalsLoose(expected, actual)
	assert.True(t, equal)
}

func TestEqualsLoose_SliceNotEqual(t *testing.T) {
	expected := []int{1, 2, 3}
	actual := []int{1, 2, 4}
	equal, _, _ := equalsLoose(expected, actual)
	assert.False(t, equal)
}

func TestEqualsLoose_MapEqual(t *testing.T) {
	expected := map[string]int{"a": 1}
	actual := map[string]int{"a": 1}
	equal, _, _ := equalsLoose(expected, actual)
	assert.True(t, equal)
}

func TestEqualsLoose_NegativeNumbers(t *testing.T) {
	equal, _, _ := equalsLoose(-42, int64(-42))
	assert.True(t, equal)
}

func TestEqualsLoose_Zero(t *testing.T) {
	equal, _, _ := equalsLoose(0, int64(0))
	assert.True(t, equal)
}

func TestEqualsLoose_LargeNumbers(t *testing.T) {
	equal, _, _ := equalsLoose(int64(9223372036854775807), int64(9223372036854775807))
	assert.True(t, equal)
}

func TestEqualsLoose_UintVsInt(t *testing.T) {
	equal, _, _ := equalsLoose(uint(42), int(42))
	assert.True(t, equal)
}

func TestEqualsLoose_EmptyString(t *testing.T) {
	equal, _, _ := equalsLoose("", "")
	assert.True(t, equal)
}

func TestEqualsLoose_EmptyVsNonEmpty(t *testing.T) {
	equal, _, _ := equalsLoose("", "non-empty")
	assert.False(t, equal)
}

func TestToNumber_SqlNullInt32_Pointer_Valid(t *testing.T) {
	val := &sql.NullInt32{Int32: 32, Valid: true}
	result, ok := typeconv.ToNumber(val)
	assert.True(t, ok)
	assert.Equal(t, float64(32), result)
}

func TestToNumber_SqlNullInt32_Pointer_Nil(t *testing.T) {
	var val *sql.NullInt32
	result, ok := typeconv.ToNumber(val)
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullInt32_Invalid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullInt32{Int32: 32, Valid: false})
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullInt16_Pointer_Valid(t *testing.T) {
	val := &sql.NullInt16{Int16: 16, Valid: true}
	result, ok := typeconv.ToNumber(val)
	assert.True(t, ok)
	assert.Equal(t, float64(16), result)
}

func TestToNumber_SqlNullInt16_Pointer_Nil(t *testing.T) {
	var val *sql.NullInt16
	result, ok := typeconv.ToNumber(val)
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullInt16_Invalid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullInt16{Int16: 16, Valid: false})
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullByte_Pointer_Valid(t *testing.T) {
	val := &sql.NullByte{Byte: 255, Valid: true}
	result, ok := typeconv.ToNumber(val)
	assert.True(t, ok)
	assert.Equal(t, float64(255), result)
}

func TestToNumber_SqlNullByte_Pointer_Nil(t *testing.T) {
	var val *sql.NullByte
	result, ok := typeconv.ToNumber(val)
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullByte_Invalid(t *testing.T) {
	result, ok := typeconv.ToNumber(sql.NullByte{Byte: 255, Valid: false})
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToNumber_SqlNullFloat64_Pointer_Valid(t *testing.T) {
	val := &sql.NullFloat64{Float64: 3.14, Valid: true}
	result, ok := typeconv.ToNumber(val)
	assert.True(t, ok)
	assert.Equal(t, float64(3.14), result)
}

func TestToNumber_SqlNullFloat64_Pointer_Nil(t *testing.T) {
	var val *sql.NullFloat64
	result, ok := typeconv.ToNumber(val)
	assert.False(t, ok)
	assert.Equal(t, float64(0), result)
}

func TestToString_ReflectPointerToString(t *testing.T) {
	type StringAlias string
	s := StringAlias("alias")
	result, ok := typeconv.ToString(&s)
	assert.True(t, ok)
	assert.Equal(t, "alias", result)
}

func TestToString_SqlNullString_Pointer_Invalid(t *testing.T) {
	val := &sql.NullString{String: "invalid", Valid: false}
	result, ok := typeconv.ToString(val)
	assert.False(t, ok)
	assert.Equal(t, "", result)
}

func TestEqualsLoose_ActualIsNullType(t *testing.T) {
	equal, _, msg := equalsLoose("expected", sql.NullString{Valid: false})
	assert.False(t, equal)
	assert.Contains(t, msg, "NULL")
}

func TestEqualsLoose_BothBoolOneFalse(t *testing.T) {
	equal, retryable, _ := equalsLoose(true, false)
	assert.False(t, equal)
	assert.True(t, retryable)
}

func TestEqualsLoose_StringTypeMismatchWithNumber(t *testing.T) {
	equal, retryable, _ := equalsLoose("string", 123)
	assert.False(t, equal)
	assert.False(t, retryable)
}

func TestEqualsLoose_ComparableTypes(t *testing.T) {
	type MyInt int
	equal, _, _ := equalsLoose(MyInt(42), MyInt(42))
	assert.True(t, equal)
}

func TestEqualsLoose_ComparableTypesNotEqual(t *testing.T) {
	type MyInt int
	equal, _, _ := equalsLoose(MyInt(42), MyInt(43))
	assert.False(t, equal)
}
