package dsl

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

func (c *Call[TReq, TResp]) ExpectNoError() *Call[TReq, TResp] {
	c.addExpectation(makeNoErrorExpectation())
	return c
}

func (c *Call[TReq, TResp]) ExpectError() *Call[TReq, TResp] {
	c.addExpectation(makeErrorExpectation())
	return c
}

func (c *Call[TReq, TResp]) ExpectStatusCode(code codes.Code) *Call[TReq, TResp] {
	c.addExpectation(makeStatusCodeExpectation(code))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldValue(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(makeFieldValueExpectation(path, expected))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldNotEmpty(path string) *Call[TReq, TResp] {
	c.addExpectation(makeFieldNotEmptyExpectation(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldExists(path string) *Call[TReq, TResp] {
	c.addExpectation(makeFieldExistsExpectation(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectMetadata(key, value string) *Call[TReq, TResp] {
	c.addExpectation(makeMetadataExpectation(key, value))
	return c
}

func makeNoErrorExpectation() *expect.Expectation[*client.Response[any]] {
	return expect.New(
		"Expect: No error",
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected no error, got: %v", err),
				}
			}
			if resp != nil && resp.Error != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected no error, got: %v", resp.Error),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: No error] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: No error]")
			}
		},
	)
}

func makeErrorExpectation() *expect.Expectation[*client.Response[any]] {
	return expect.New(
		"Expect: Error",
		func(err error, resp *client.Response[any]) polling.CheckResult {
			hasError := err != nil || (resp != nil && resp.Error != nil)
			if !hasError {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Expected an error, but call succeeded",
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Error] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: Error]")
			}
		},
	)
}

func makeStatusCodeExpectation(code codes.Code) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect: Status code %s (%d)", code.String(), code),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			var actualCode codes.Code
			if err != nil {
				st, ok := status.FromError(err)
				if ok {
					actualCode = st.Code()
				} else {
					actualCode = codes.Unknown
				}
			} else if resp != nil && resp.Error != nil {
				st, ok := status.FromError(resp.Error)
				if ok {
					actualCode = st.Code()
				} else {
					actualCode = codes.Unknown
				}
			} else {
				actualCode = codes.OK
			}

			if actualCode != code {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected status %s (%d), got %s (%d)", code.String(), code, actualCode.String(), actualCode),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Status code %s (%d)] %s", code.String(), code, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Status code %s (%d)]", code.String(), code)
			}
		},
	)
}

func getResponseJSON(resp *client.Response[any]) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	// Try RawBody first
	if len(resp.RawBody) > 0 {
		// Check if it's valid JSON
		if gjson.ValidBytes(resp.RawBody) {
			return resp.RawBody, nil
		}
	}

	// Marshal the body to JSON
	jsonBytes, err := json.Marshal(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return jsonBytes, nil
}

func makeFieldValueExpectation(path string, expected any) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect: Field '%s' = %v", path, expected),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Call failed with error",
				}
			}
			if resp == nil || resp.Body == nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response body is nil",
				}
			}

			jsonBytes, jsonErr := getResponseJSON(resp)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot parse response: %v", jsonErr),
				}
			}

			result := gjson.GetBytes(jsonBytes, path)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Field '%s' does not exist", path),
				}
			}

			ok, msg := compareValues(result, expected)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    msg,
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Field '%s' = %v] %s", path, expected, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Field '%s' = %v]", path, expected)
			}
		},
	)
}

func makeFieldNotEmptyExpectation(path string) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect: Field '%s' not empty", path),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Call failed with error",
				}
			}
			if resp == nil || resp.Body == nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response body is nil",
				}
			}

			jsonBytes, jsonErr := getResponseJSON(resp)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot parse response: %v", jsonErr),
				}
			}

			result := gjson.GetBytes(jsonBytes, path)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Field '%s' does not exist", path),
				}
			}

			if isEmptyValue(result) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Field '%s' is empty", path),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Field '%s' not empty] %s", path, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Field '%s' not empty]", path)
			}
		},
	)
}

func makeFieldExistsExpectation(path string) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect: Field '%s' exists", path),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Call failed with error",
				}
			}
			if resp == nil || resp.Body == nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response body is nil",
				}
			}

			jsonBytes, jsonErr := getResponseJSON(resp)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot parse response: %v", jsonErr),
				}
			}

			result := gjson.GetBytes(jsonBytes, path)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Field '%s' does not exist", path),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Field '%s' exists] %s", path, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Field '%s' exists]", path)
			}
		},
	)
}

func makeMetadataExpectation(key, expectedValue string) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect: Metadata '%s' = '%s'", key, expectedValue),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Call failed with error",
				}
			}
			if resp == nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response is nil",
				}
			}

			values := resp.Metadata.Get(key)
			if len(values) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Metadata key '%s' not found", key),
				}
			}

			found := false
			for _, v := range values {
				if v == expectedValue {
					found = true
					break
				}
			}

			if !found {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Metadata '%s' = %v, expected '%s'", key, values, expectedValue),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Metadata '%s' = '%s'] %s", key, expectedValue, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Metadata '%s' = '%s']", key, expectedValue)
			}
		},
	)
}

func isEmptyValue(result gjson.Result) bool {
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

func compareValues(result gjson.Result, expected any) (bool, string) {
	if expected == nil {
		if result.Type != gjson.Null {
			return false, fmt.Sprintf("expected null, got %v", result.Value())
		}
		return true, ""
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
	case int, int8, int16, int32, int64:
		expInt := reflect.ValueOf(expected).Int()
		if result.Type != gjson.Number {
			return false, fmt.Sprintf("expected number %d, got %v", expInt, result.Value())
		}
		if result.Int() != expInt {
			return false, fmt.Sprintf("expected %d, got %d", expInt, result.Int())
		}
	case uint, uint8, uint16, uint32, uint64:
		expUint := reflect.ValueOf(expected).Uint()
		if result.Type != gjson.Number {
			return false, fmt.Sprintf("expected number %d, got %v", expUint, result.Value())
		}
		if result.Uint() != expUint {
			return false, fmt.Sprintf("expected %d, got %d", expUint, result.Uint())
		}
	case float32, float64:
		expFloat := reflect.ValueOf(expected).Float()
		if result.Type != gjson.Number {
			return false, fmt.Sprintf("expected number %v, got %v", expFloat, result.Value())
		}
		if result.Float() != expFloat {
			return false, fmt.Sprintf("expected %v, got %v", expFloat, result.Float())
		}
	default:
		return false, fmt.Sprintf("unsupported expected type %T", expected)
	}

	return true, ""
}
