package dsl

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

func TestMakeFieldValueExpectation_Success(t *testing.T) {
	exp := makeFieldValueExpectation("user.id", 123)
	jsonData := []byte(`{"user": {"id": 123, "name": "John"}}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
	assert.Empty(t, result.Reason)
}

func TestMakeFieldValueExpectation_Failure(t *testing.T) {
	exp := makeFieldValueExpectation("user.id", 456)
	jsonData := []byte(`{"user": {"id": 123, "name": "John"}}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
	assert.NotEmpty(t, result.Reason)
}

func TestMakeFieldValueExpectation_PathNotExists(t *testing.T) {
	exp := makeFieldValueExpectation("nonexistent.path", "value")
	jsonData := []byte(`{"user": {"id": 123}}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldValueExpectation_StringValue(t *testing.T) {
	exp := makeFieldValueExpectation("name", "John")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldValueExpectation_BoolValue(t *testing.T) {
	exp := makeFieldValueExpectation("active", true)
	jsonData := []byte(`{"active": true}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldValueExpectation_FloatValue(t *testing.T) {
	exp := makeFieldValueExpectation("price", 19.99)
	jsonData := []byte(`{"price": 19.99}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldValueExpectation_NestedPath(t *testing.T) {
	exp := makeFieldValueExpectation("data.items.0.name", "first")
	jsonData := []byte(`{"data": {"items": [{"name": "first"}, {"name": "second"}]}}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_Success(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("name")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_EmptyString(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("name")
	jsonData := []byte(`{"name": ""}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_NullValue(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("name")
	jsonData := []byte(`{"name": null}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_NonEmptyArray(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("items")
	jsonData := []byte(`{"items": [1, 2, 3]}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_EmptyArray(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("items")
	jsonData := []byte(`{"items": []}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_Zero(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("count")
	jsonData := []byte(`{"count": 0}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_NonZeroNumber(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("count")
	jsonData := []byte(`{"count": 42}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldEmptyExpectation_EmptyString(t *testing.T) {
	exp := makeFieldEmptyExpectation("name")
	jsonData := []byte(`{"name": ""}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldEmptyExpectation_NonEmptyString(t *testing.T) {
	exp := makeFieldEmptyExpectation("name")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldEmptyExpectation_EmptyArray(t *testing.T) {
	exp := makeFieldEmptyExpectation("items")
	jsonData := []byte(`{"items": []}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldEmptyExpectation_FieldNotExists(t *testing.T) {
	exp := makeFieldEmptyExpectation("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldIsNullExpectation_Success(t *testing.T) {
	exp := makeFieldIsNullExpectation("value")
	jsonData := []byte(`{"value": null}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldIsNullExpectation_NotNull(t *testing.T) {
	exp := makeFieldIsNullExpectation("value")
	jsonData := []byte(`{"value": "something"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldIsNotNullExpectation_Success(t *testing.T) {
	exp := makeFieldIsNotNullExpectation("value")
	jsonData := []byte(`{"value": "something"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldIsNotNullExpectation_IsNull(t *testing.T) {
	exp := makeFieldIsNotNullExpectation("value")
	jsonData := []byte(`{"value": null}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldTrueExpectation_Success(t *testing.T) {
	exp := makeFieldTrueExpectation("active")
	jsonData := []byte(`{"active": true}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldTrueExpectation_False(t *testing.T) {
	exp := makeFieldTrueExpectation("active")
	jsonData := []byte(`{"active": false}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldTrueExpectation_NotBoolean(t *testing.T) {
	exp := makeFieldTrueExpectation("active")
	jsonData := []byte(`{"active": "yes"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldFalseExpectation_Success(t *testing.T) {
	exp := makeFieldFalseExpectation("deleted")
	jsonData := []byte(`{"deleted": false}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldFalseExpectation_True(t *testing.T) {
	exp := makeFieldFalseExpectation("deleted")
	jsonData := []byte(`{"deleted": true}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeMessageExpectation_ExactMatch(t *testing.T) {
	expected := map[string]interface{}{
		"id":   float64(123),
		"name": "test",
	}
	exp := makeMessageExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestMakeMessageExpectation_ExtraFieldAllowed(t *testing.T) {
	expected := map[string]interface{}{
		"id":   float64(123),
		"name": "test",
	}
	exp := makeMessageExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": "field"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestMakeMessageExpectation_ExactMatchWithAllFields(t *testing.T) {
	expected := map[string]interface{}{
		"id":    float64(123),
		"name":  "test",
		"extra": "field",
	}
	exp := makeMessageExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": "field"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestMakeMessageExpectation_MissingField(t *testing.T) {
	expected := map[string]interface{}{
		"id":    float64(123),
		"name":  "test",
		"extra": "required",
	}
	exp := makeMessageExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeMessagePartialExpectation_Match(t *testing.T) {
	expected := map[string]interface{}{
		"id": float64(123),
	}
	exp := makeMessagePartialExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": "field"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestMakeMessagePartialExpectation_Mismatch(t *testing.T) {
	expected := map[string]interface{}{
		"id": float64(456),
	}
	exp := makeMessagePartialExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeMessagePartialExpectation_NestedMatch(t *testing.T) {
	expected := map[string]interface{}{
		"user": map[string]interface{}{
			"id": float64(123),
		},
	}
	exp := makeMessagePartialExpectation(expected)
	jsonData := []byte(`{"user": {"id": 123, "name": "John"}, "timestamp": 12345}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestMakeMessageExpectation_WithStruct(t *testing.T) {
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	expected := TestStruct{ID: 123, Name: "test"}
	exp := makeMessageExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestMakeMessagePartialExpectation_WithStruct(t *testing.T) {
	type TestStruct struct {
		ID int `json:"id"`
	}
	expected := TestStruct{ID: 123}
	exp := makeMessagePartialExpectation(expected)
	jsonData := []byte(`{"id": 123, "name": "test", "extra": true}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestExpectation_CheckResultRetryable(t *testing.T) {
	exp := makeFieldValueExpectation("id", 456)
	jsonData := []byte(`{"id": 123}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestExpectation_Name(t *testing.T) {
	exp := makeFieldValueExpectation("user.id", 123)

	assert.Contains(t, exp.Name, "user.id")
	assert.Contains(t, exp.Name, "123")
}

func TestExpectation_CheckWithError(t *testing.T) {
	exp := makeFieldValueExpectation("id", 123)
	jsonData := []byte(`{"id": 123}`)
	testErr := polling.CheckResult{Ok: false, Reason: "test error"}

	result := exp.Check(nil, jsonData)

	assert.NotEqual(t, testErr.Reason, result.Reason)
}

func TestMakeFieldValueExpectation_NumericTypeConversion(t *testing.T) {
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
			exp := makeFieldValueExpectation("val", tc.expected)
			result := exp.Check(nil, []byte(tc.json))
			assert.Equal(t, tc.wantOk, result.Ok, "Reason: %s", result.Reason)
		})
	}
}

func TestMakeFieldValueExpectation_ArrayAccess(t *testing.T) {
	exp := makeFieldValueExpectation("items.1.name", "second")
	jsonData := []byte(`{"items": [{"name": "first"}, {"name": "second"}]}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldValueExpectation_ArrayLength(t *testing.T) {
	exp := makeFieldValueExpectation("items.#", float64(3))
	jsonData := []byte(`{"items": ["a", "b", "c"]}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok, "Result reason: %s", result.Reason)
}

func TestMakeFieldIsNullExpectation_PathNotExists(t *testing.T) {
	exp := makeFieldIsNullExpectation("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.True(t, result.Ok)
}

func TestMakeFieldIsNotNullExpectation_PathNotExists(t *testing.T) {
	exp := makeFieldIsNotNullExpectation("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestMakeFieldNotEmptyExpectation_PathNotExists(t *testing.T) {
	exp := makeFieldNotEmptyExpectation("nonexistent")
	jsonData := []byte(`{"name": "John"}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestExpectation_InvalidJSON(t *testing.T) {
	exp := makeFieldValueExpectation("id", 123)
	jsonData := []byte(`not valid json`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}

func TestExpectation_EmptyJSON(t *testing.T) {
	exp := makeFieldValueExpectation("id", 123)
	jsonData := []byte(`{}`)

	result := exp.Check(nil, jsonData)

	assert.False(t, result.Ok)
}
