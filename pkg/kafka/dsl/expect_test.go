package dsl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

func TestFieldEquals_Success(t *testing.T) {
	exp := bytesSource.FieldEquals("user.id", 123)
	jsonData := []byte(`{"user": {"id": 123, "name": "John"}}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
	assert.Empty(t, result.Reason)
}

func TestFieldEquals_Failure(t *testing.T) {
	exp := bytesSource.FieldEquals("user.id", 456)
	jsonData := []byte(`{"user": {"id": 123, "name": "John"}}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
	assert.NotEmpty(t, result.Reason)
}

func TestFieldEquals_PathNotExists(t *testing.T) {
	exp := bytesSource.FieldEquals("nonexistent.path", "value")
	jsonData := []byte(`{"user": {"id": 123}}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldEquals_StringValue(t *testing.T) {
	exp := bytesSource.FieldEquals("name", "John")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldEquals_BoolValue(t *testing.T) {
	exp := bytesSource.FieldEquals("active", true)
	jsonData := []byte(`{"active": true}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldEquals_FloatValue(t *testing.T) {
	exp := bytesSource.FieldEquals("price", 19.99)
	jsonData := []byte(`{"price": 19.99}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldEquals_NestedPath(t *testing.T) {
	exp := bytesSource.FieldEquals("data.items.0.name", "first")
	jsonData := []byte(`{"data": {"items": [{"name": "first"}, {"name": "second"}]}}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldNotEmpty_Success(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("name")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldNotEmpty_EmptyString(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("name")
	jsonData := []byte(`{"name": ""}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldNotEmpty_NullValue(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("name")
	jsonData := []byte(`{"name": null}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldNotEmpty_NonEmptyArray(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("items")
	jsonData := []byte(`{"items": [1, 2, 3]}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldNotEmpty_EmptyArray(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("items")
	jsonData := []byte(`{"items": []}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldNotEmpty_Zero(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("count")
	jsonData := []byte(`{"count": 0}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldNotEmpty_NonZeroNumber(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("count")
	jsonData := []byte(`{"count": 42}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldEmpty_EmptyString(t *testing.T) {
	exp := bytesSource.FieldEmpty("name")
	jsonData := []byte(`{"name": ""}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldEmpty_NonEmptyString(t *testing.T) {
	exp := bytesSource.FieldEmpty("name")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldEmpty_EmptyArray(t *testing.T) {
	exp := bytesSource.FieldEmpty("items")
	jsonData := []byte(`{"items": []}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldEmpty_FieldNotExists(t *testing.T) {
	exp := bytesSource.FieldEmpty("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldIsNull_Success(t *testing.T) {
	exp := bytesSource.FieldIsNull("value")
	jsonData := []byte(`{"value": null}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldIsNull_NotNull(t *testing.T) {
	exp := bytesSource.FieldIsNull("value")
	jsonData := []byte(`{"value": "something"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldIsNotNull_Success(t *testing.T) {
	exp := bytesSource.FieldIsNotNull("value")
	jsonData := []byte(`{"value": "something"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldIsNotNull_IsNull(t *testing.T) {
	exp := bytesSource.FieldIsNotNull("value")
	jsonData := []byte(`{"value": null}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldTrue_Success(t *testing.T) {
	exp := bytesSource.FieldTrue("active")
	jsonData := []byte(`{"active": true}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldTrue_False(t *testing.T) {
	exp := bytesSource.FieldTrue("active")
	jsonData := []byte(`{"active": false}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldTrue_NotBoolean(t *testing.T) {
	exp := bytesSource.FieldTrue("active")
	jsonData := []byte(`{"active": "yes"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldFalse_Success(t *testing.T) {
	exp := bytesSource.FieldFalse("deleted")
	jsonData := []byte(`{"deleted": false}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldFalse_True(t *testing.T) {
	exp := bytesSource.FieldFalse("deleted")
	jsonData := []byte(`{"deleted": true}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestBodyEquals_ExactMatch(t *testing.T) {
	expected := map[string]interface{}{
		"id":   float64(123),
		"name": "test",
	}
	exp := bytesSource.BodyEquals(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestBodyEquals_ExtraFieldAllowed(t *testing.T) {
	expected := map[string]interface{}{
		"id":   float64(123),
		"name": "test",
	}
	exp := bytesSource.BodyEquals(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": "field"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestBodyEquals_ExactMatchWithAllFields(t *testing.T) {
	expected := map[string]interface{}{
		"id":    float64(123),
		"name":  "test",
		"extra": "field",
	}
	exp := bytesSource.BodyEquals(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": "field"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestBodyEquals_MissingField(t *testing.T) {
	expected := map[string]interface{}{
		"id":    float64(123),
		"name":  "test",
		"extra": "required",
	}
	exp := bytesSource.BodyEquals(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestBodyPartial_Match(t *testing.T) {
	expected := map[string]interface{}{
		"id": float64(123),
	}
	exp := bytesSource.BodyPartial(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": "field"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestBodyPartial_Mismatch(t *testing.T) {
	expected := map[string]interface{}{
		"id": float64(456),
	}
	exp := bytesSource.BodyPartial(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestBodyPartial_NestedMatch(t *testing.T) {
	expected := map[string]interface{}{
		"user": map[string]interface{}{
			"id": float64(123),
		},
	}
	exp := bytesSource.BodyPartial(expected)
	jsonData := []byte(`{"user": {"id": 123, "name": "John"}, "timestamp": 12345}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestBodyEquals_WithStruct(t *testing.T) {
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	expected := TestStruct{ID: 123, Name: "test"}
	exp := bytesSource.BodyEquals(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestBodyPartial_WithStruct(t *testing.T) {
	type TestStruct struct {
		ID int `json:"id"`
	}
	expected := TestStruct{ID: 123}
	exp := bytesSource.BodyPartial(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": true}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestExpectation_CheckResultRetryable(t *testing.T) {
	exp := bytesSource.FieldEquals("id", 456)
	jsonData := []byte(`{"id": 123}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestExpectation_Name(t *testing.T) {
	exp := bytesSource.FieldEquals("user.id", 123)

	assert.Contains(t, exp.Name, "user.id")
	assert.Contains(t, exp.Name, "123")
}

func TestExpectation_CheckWithError(t *testing.T) {
	exp := bytesSource.FieldEquals("id", 123)
	jsonData := []byte(`{"id": 123}`)
	testErr := polling.CheckResult{Ok: false, Reason: "test error"}

	result := exp.Check(nil, jsonData)

	assert.NotEqual(t, testErr.Reason, result.Reason)
}

func TestFieldEquals_NumericTypeConversion(t *testing.T) {
	testCases := []struct {
		name     string
		expected interface{}
		json     string
		wantOk   bool
	}{
		{"int to json number", 123, `{"val": 123}`, true},
		{"int64 to json number", int64(123), `{"val": 123}`, true},
		{"float64 to json number", float64(123), `{"val": 123}`, true},
		{"float with decimal", 19.99, `{"val": 19.99}`, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exp := bytesSource.FieldEquals("val", tc.expected)
			result := exp.Check(nil, []byte(tc.json))
			assert.Equal(t, tc.wantOk, result.Ok, "Reason: %s", result.Reason)
		})
	}
}

func TestFieldEquals_ArrayAccess(t *testing.T) {
	exp := bytesSource.FieldEquals("items.1.name", "second")
	jsonData := []byte(`{"items": [{"name": "first"}, {"name": "second"}]}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldEquals_ArrayLength(t *testing.T) {
	exp := bytesSource.FieldEquals("items.#", float64(3))
	jsonData := []byte(`{"items": ["a", "b", "c"]}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestFieldIsNull_PathNotExists(t *testing.T) {
	exp := bytesSource.FieldIsNull("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestFieldIsNotNull_PathNotExists(t *testing.T) {
	exp := bytesSource.FieldIsNotNull("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestFieldNotEmpty_PathNotExists(t *testing.T) {
	exp := bytesSource.FieldNotEmpty("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestExpectation_InvalidJSON(t *testing.T) {
	exp := bytesSource.FieldEquals("id", 123)
	jsonData := []byte(`not valid json`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestExpectation_EmptyJSON(t *testing.T) {
	exp := bytesSource.FieldEquals("id", 123)
	jsonData := []byte(`{}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}
