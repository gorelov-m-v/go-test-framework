package dsl

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

// ExpectExists expects the key to exist in Redis
func (q *Query) ExpectExists() *Query {
	q.addExpectation(makeExistsExpectation())
	return q
}

// ExpectNotExists expects the key to not exist in Redis
func (q *Query) ExpectNotExists() *Query {
	q.addExpectation(makeNotExistsExpectation())
	return q
}

// ExpectValue expects the key's value to equal the expected value
func (q *Query) ExpectValue(expected string) *Query {
	q.addExpectation(makeValueExpectation(expected))
	return q
}

// ExpectValueNotEmpty expects the key's value to be non-empty
func (q *Query) ExpectValueNotEmpty() *Query {
	q.addExpectation(makeValueNotEmptyExpectation())
	return q
}

// ExpectJSONField expects a JSON field in the value to have a specific value
// Uses GJSON path syntax for nested fields
func (q *Query) ExpectJSONField(path string, expected any) *Query {
	q.addExpectation(makeJSONFieldExpectation(path, expected))
	return q
}

// ExpectJSONFieldNotEmpty expects a JSON field in the value to be non-empty
func (q *Query) ExpectJSONFieldNotEmpty(path string) *Query {
	q.addExpectation(makeJSONFieldNotEmptyExpectation(path))
	return q
}

// ExpectTTL expects the key to have a specific TTL range
func (q *Query) ExpectTTL(minTTL, maxTTL time.Duration) *Query {
	q.addExpectation(makeTTLExpectation(minTTL, maxTTL))
	return q
}

// ExpectNoTTL expects the key to have no TTL (persistent)
func (q *Query) ExpectNoTTL() *Query {
	q.addExpectation(makeNoTTLExpectation())
	return q
}

func makeExistsExpectation() *expect.Expectation[*client.Result] {
	return expect.New(
		"Expect: Key exists",
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if !result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Key exists] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: Key exists]")
			}
		},
	)
}

func makeNotExistsExpectation() *expect.Expectation[*client.Result] {
	return expect.New(
		"Expect: Key not exists",
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' exists but expected not to", result.Key),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Key not exists] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: Key not exists]")
			}
		},
	)
}

func makeValueExpectation(expected string) *expect.Expectation[*client.Result] {
	return expect.New(
		fmt.Sprintf("Expect: Value = %q", expected),
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if !result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
				}
			}
			if result.Value != expected {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected value %q, got %q", expected, result.Value),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Value = %q] %s", expected, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Value = %q]", expected)
			}
		},
	)
}

func makeValueNotEmptyExpectation() *expect.Expectation[*client.Result] {
	return expect.New(
		"Expect: Value not empty",
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if !result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
				}
			}
			if strings.TrimSpace(result.Value) == "" {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Value is empty",
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Value not empty] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: Value not empty]")
			}
		},
	)
}

func makeJSONFieldExpectation(path string, expected any) *expect.Expectation[*client.Result] {
	return expect.New(
		fmt.Sprintf("Expect: JSON field '%s' = %v", path, expected),
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if !result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
				}
			}

			if !gjson.Valid(result.Value) {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Value is not valid JSON",
				}
			}

			gjResult := gjson.Get(result.Value, path)
			if !gjResult.Exists() {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist", path),
				}
			}

			ok, msg := compareJSONValue(gjResult, expected)
			if !ok {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    msg,
				}
			}

			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: JSON field '%s' = %v] %s", path, expected, checkRes.Reason)
			} else {
				a.True(true, "[Expect: JSON field '%s' = %v]", path, expected)
			}
		},
	)
}

func makeJSONFieldNotEmptyExpectation(path string) *expect.Expectation[*client.Result] {
	return expect.New(
		fmt.Sprintf("Expect: JSON field '%s' not empty", path),
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if !result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
				}
			}

			if !gjson.Valid(result.Value) {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Value is not valid JSON",
				}
			}

			gjResult := gjson.Get(result.Value, path)
			if !gjResult.Exists() {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist", path),
				}
			}

			if isEmptyJSONValue(gjResult) {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' is empty", path),
				}
			}

			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: JSON field '%s' not empty] %s", path, checkRes.Reason)
			} else {
				a.True(true, "[Expect: JSON field '%s' not empty]", path)
			}
		},
	)
}

func makeTTLExpectation(minTTL, maxTTL time.Duration) *expect.Expectation[*client.Result] {
	return expect.New(
		fmt.Sprintf("Expect: TTL between %v and %v", minTTL, maxTTL),
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if !result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
				}
			}
			if result.TTL < minTTL || result.TTL > maxTTL {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("TTL is %v, expected between %v and %v", result.TTL, minTTL, maxTTL),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: TTL between %v and %v] %s", minTTL, maxTTL, checkRes.Reason)
			} else {
				a.True(true, "[Expect: TTL between %v and %v]", minTTL, maxTTL)
			}
		},
	)
}

func makeNoTTLExpectation() *expect.Expectation[*client.Result] {
	return expect.New(
		"Expect: No TTL (persistent)",
		func(err error, result *client.Result) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			if result == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Result is nil",
				}
			}
			if result.Error != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Redis error: %v", result.Error),
				}
			}
			if !result.Exists {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key '%s' does not exist", result.Key),
				}
			}
			// TTL of -1 means no expiration
			if result.TTL != -1 {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Key has TTL of %v, expected no TTL", result.TTL),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result *client.Result, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: No TTL (persistent)] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: No TTL (persistent)]")
			}
		},
	)
}

func isEmptyJSONValue(result gjson.Result) bool {
	if !result.Exists() {
		return true
	}

	switch result.Type {
	case gjson.Null:
		return true
	case gjson.String:
		return strings.TrimSpace(result.String()) == ""
	case gjson.JSON:
		if result.IsArray() {
			return len(result.Array()) == 0
		}
		if result.IsObject() {
			return len(result.Map()) == 0
		}
	}

	return false
}

func compareJSONValue(result gjson.Result, expected any) (bool, string) {
	if expected == nil {
		if result.Type != gjson.Null {
			return false, fmt.Sprintf("expected null, got %v", result.Value())
		}
		return true, ""
	}

	// Marshal expected to JSON and compare
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return false, fmt.Sprintf("failed to marshal expected value: %v", err)
	}

	actualJSON := result.Raw
	if actualJSON == "" {
		actualJSON = fmt.Sprintf("%v", result.Value())
	}

	switch exp := expected.(type) {
	case string:
		if result.Type != gjson.String {
			return false, fmt.Sprintf("expected string %q, got %v", exp, result.Value())
		}
		if result.String() != exp {
			return false, fmt.Sprintf("expected %q, got %q", exp, result.String())
		}
	case bool:
		if result.Type != gjson.True && result.Type != gjson.False {
			return false, fmt.Sprintf("expected bool %v, got %v", exp, result.Value())
		}
		if result.Bool() != exp {
			return false, fmt.Sprintf("expected %v, got %v", exp, result.Bool())
		}
	case float64:
		if result.Type != gjson.Number {
			return false, fmt.Sprintf("expected number %v, got %v", exp, result.Value())
		}
		if result.Float() != exp {
			return false, fmt.Sprintf("expected %v, got %v", exp, result.Float())
		}
	case int:
		if result.Type != gjson.Number {
			return false, fmt.Sprintf("expected number %d, got %v", exp, result.Value())
		}
		if result.Int() != int64(exp) {
			return false, fmt.Sprintf("expected %d, got %d", exp, result.Int())
		}
	case int64:
		if result.Type != gjson.Number {
			return false, fmt.Sprintf("expected number %d, got %v", exp, result.Value())
		}
		if result.Int() != exp {
			return false, fmt.Sprintf("expected %d, got %d", exp, result.Int())
		}
	default:
		// For complex types, compare JSON
		if string(expectedJSON) != actualJSON {
			return false, fmt.Sprintf("expected %s, got %s", string(expectedJSON), actualJSON)
		}
	}

	return true, ""
}
