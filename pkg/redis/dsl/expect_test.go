package dsl

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

func TestPreCheck_Success(t *testing.T) {
	result := &client.Result{Key: "test", Exists: true}
	checkResult, ok := preCheck(nil, result)

	assert.True(t, ok)
	assert.True(t, checkResult.Ok || checkResult.Reason == "")
}

func TestPreCheck_Error(t *testing.T) {
	checkResult, ok := preCheck(errors.New("test error"), nil)

	assert.False(t, ok)
	assert.False(t, checkResult.Ok)
	assert.True(t, checkResult.Retryable)
}

func TestPreCheck_NilResult(t *testing.T) {
	checkResult, ok := preCheck(nil, nil)

	assert.False(t, ok)
	assert.Contains(t, checkResult.Reason, "nil")
}

func TestPreCheck_ResultError(t *testing.T) {
	result := &client.Result{Error: errors.New("redis error")}
	checkResult, ok := preCheck(nil, result)

	assert.False(t, ok)
	assert.Contains(t, checkResult.Reason, "Redis error")
}

func TestPreCheckKeyExists_Success(t *testing.T) {
	result := &client.Result{Key: "test", Exists: true}
	checkResult, ok := preCheckKeyExists(nil, result)

	assert.True(t, ok)
	assert.True(t, checkResult.Ok || checkResult.Reason == "")
}

func TestPreCheckKeyExists_NotExists(t *testing.T) {
	result := &client.Result{Key: "test", Exists: false}
	checkResult, ok := preCheckKeyExists(nil, result)

	assert.False(t, ok)
	assert.Contains(t, checkResult.Reason, "does not exist")
}

func TestMakeExistsExpectation_Success(t *testing.T) {
	exp := makeExistsExpectation()
	result := &client.Result{Key: "test", Exists: true}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeExistsExpectation_NotExists(t *testing.T) {
	exp := makeExistsExpectation()
	result := &client.Result{Key: "test", Exists: false}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
	assert.True(t, checkResult.Retryable)
}

func TestMakeExistsExpectation_Error(t *testing.T) {
	exp := makeExistsExpectation()

	checkResult := exp.Check(errors.New("test error"), nil)

	assert.False(t, checkResult.Ok)
}

func TestMakeNotExistsExpectation_Success(t *testing.T) {
	exp := makeNotExistsExpectation()
	result := &client.Result{Key: "test", Exists: false}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeNotExistsExpectation_Exists(t *testing.T) {
	exp := makeNotExistsExpectation()
	result := &client.Result{Key: "test", Exists: true}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
	assert.True(t, checkResult.Retryable)
	assert.Contains(t, checkResult.Reason, "exists but expected not to")
}

func TestMakeValueExpectation_Match(t *testing.T) {
	exp := makeValueExpectation("expected_value")
	result := &client.Result{Key: "test", Exists: true, Value: "expected_value"}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeValueExpectation_Mismatch(t *testing.T) {
	exp := makeValueExpectation("expected_value")
	result := &client.Result{Key: "test", Exists: true, Value: "actual_value"}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
	assert.Contains(t, checkResult.Reason, "expected_value")
	assert.Contains(t, checkResult.Reason, "actual_value")
}

func TestMakeValueExpectation_NotExists(t *testing.T) {
	exp := makeValueExpectation("value")
	result := &client.Result{Key: "test", Exists: false}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeValueExpectation_EmptyValue(t *testing.T) {
	exp := makeValueExpectation("")
	result := &client.Result{Key: "test", Exists: true, Value: ""}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeValueNotEmptyExpectation_Success(t *testing.T) {
	exp := makeValueNotEmptyExpectation()
	result := &client.Result{Key: "test", Exists: true, Value: "some value"}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeValueNotEmptyExpectation_Empty(t *testing.T) {
	exp := makeValueNotEmptyExpectation()
	result := &client.Result{Key: "test", Exists: true, Value: ""}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
	assert.Contains(t, checkResult.Reason, "empty")
}

func TestMakeValueNotEmptyExpectation_Whitespace(t *testing.T) {
	exp := makeValueNotEmptyExpectation()
	result := &client.Result{Key: "test", Exists: true, Value: "   "}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeValueNotEmptyExpectation_NotExists(t *testing.T) {
	exp := makeValueNotEmptyExpectation()
	result := &client.Result{Key: "test", Exists: false}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeJSONFieldExpectation_Success(t *testing.T) {
	exp := makeJSONFieldExpectation("name", "John")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"name": "John", "age": 30}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok, "Reason: %s", checkResult.Reason)
}

func TestMakeJSONFieldExpectation_Mismatch(t *testing.T) {
	exp := makeJSONFieldExpectation("name", "John")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"name": "Jane"}`,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeJSONFieldExpectation_PathNotExists(t *testing.T) {
	exp := makeJSONFieldExpectation("nonexistent", "value")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"name": "John"}`,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeJSONFieldExpectation_InvalidJSON(t *testing.T) {
	exp := makeJSONFieldExpectation("name", "John")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `not valid json`,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeJSONFieldExpectation_NestedPath(t *testing.T) {
	exp := makeJSONFieldExpectation("user.name", "John")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"user": {"name": "John", "age": 30}}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok, "Reason: %s", checkResult.Reason)
}

func TestMakeJSONFieldExpectation_IntValue(t *testing.T) {
	exp := makeJSONFieldExpectation("age", 30)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"name": "John", "age": 30}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok, "Reason: %s", checkResult.Reason)
}

func TestMakeJSONFieldExpectation_BoolValue(t *testing.T) {
	exp := makeJSONFieldExpectation("active", true)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"active": true}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok, "Reason: %s", checkResult.Reason)
}

func TestMakeJSONFieldExpectation_ArrayAccess(t *testing.T) {
	exp := makeJSONFieldExpectation("items.0", "first")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"items": ["first", "second"]}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok, "Reason: %s", checkResult.Reason)
}

func TestMakeJSONFieldNotEmptyExpectation_Success(t *testing.T) {
	exp := makeJSONFieldNotEmptyExpectation("name")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"name": "John"}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeJSONFieldNotEmptyExpectation_Empty(t *testing.T) {
	exp := makeJSONFieldNotEmptyExpectation("name")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"name": ""}`,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeJSONFieldNotEmptyExpectation_Null(t *testing.T) {
	exp := makeJSONFieldNotEmptyExpectation("name")
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"name": null}`,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeTTLExpectation_InRange(t *testing.T) {
	exp := makeTTLExpectation(1*time.Minute, 10*time.Minute)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    5 * time.Minute,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeTTLExpectation_BelowMin(t *testing.T) {
	exp := makeTTLExpectation(5*time.Minute, 10*time.Minute)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    1 * time.Minute,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
	assert.Contains(t, checkResult.Reason, "TTL")
}

func TestMakeTTLExpectation_AboveMax(t *testing.T) {
	exp := makeTTLExpectation(1*time.Minute, 5*time.Minute)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    10 * time.Minute,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeTTLExpectation_ExactMin(t *testing.T) {
	exp := makeTTLExpectation(5*time.Minute, 10*time.Minute)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    5 * time.Minute,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeTTLExpectation_ExactMax(t *testing.T) {
	exp := makeTTLExpectation(5*time.Minute, 10*time.Minute)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    10 * time.Minute,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeTTLExpectation_NotExists(t *testing.T) {
	exp := makeTTLExpectation(1*time.Minute, 10*time.Minute)
	result := &client.Result{Key: "test", Exists: false}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestMakeNoTTLExpectation_Success(t *testing.T) {
	exp := makeNoTTLExpectation()
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    -1,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeNoTTLExpectation_HasTTL(t *testing.T) {
	exp := makeNoTTLExpectation()
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    5 * time.Minute,
	}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
	assert.Contains(t, checkResult.Reason, "TTL")
}

func TestMakeNoTTLExpectation_NotExists(t *testing.T) {
	exp := makeNoTTLExpectation()
	result := &client.Result{Key: "test", Exists: false}

	checkResult := exp.Check(nil, result)

	assert.False(t, checkResult.Ok)
}

func TestExpectation_Names(t *testing.T) {
	tests := []struct {
		name     string
		exp      *expect.Expectation[*client.Result]
		contains string
	}{
		{"Exists", makeExistsExpectation(), "exists"},
		{"NotExists", makeNotExistsExpectation(), "not exists"},
		{"Value", makeValueExpectation("test"), "test"},
		{"ValueNotEmpty", makeValueNotEmptyExpectation(), "not empty"},
		{"JSONField", makeJSONFieldExpectation("path", "val"), "path"},
		{"JSONFieldNotEmpty", makeJSONFieldNotEmptyExpectation("field"), "field"},
		{"TTL", makeTTLExpectation(1*time.Second, 10*time.Second), "TTL"},
		{"NoTTL", makeNoTTLExpectation(), "No TTL"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Contains(t, tc.exp.Name, tc.contains)
		})
	}
}

func TestMakeJSONFieldExpectation_FloatValue(t *testing.T) {
	exp := makeJSONFieldExpectation("price", 19.99)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"price": 19.99}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok, "Reason: %s", checkResult.Reason)
}

func TestMakeValueExpectation_JSONString(t *testing.T) {
	exp := makeValueExpectation(`{"key":"value"}`)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		Value:  `{"key":"value"}`,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}

func TestMakeTTLExpectation_ZeroTTL(t *testing.T) {
	exp := makeTTLExpectation(0, 10*time.Minute)
	result := &client.Result{
		Key:    "test",
		Exists: true,
		TTL:    0,
	}

	checkResult := exp.Check(nil, result)

	assert.True(t, checkResult.Ok)
}
