package dsl

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"go-test-framework/internal/expect"
	"go-test-framework/pkg/extension"
	"go-test-framework/pkg/http/client"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"
)

// ToDo: ExpectResponseToMatchJSONSchema("path/to/schema.json")
// ToDo: WithRetries(count: 3, delay: 2*time.Second)
// ToDo: MultipartBody(form *client.MultipartForm)
// ToDo: parseErrorResponse жестко зашита на определенную структуру, сделать гибкой

func getJSONResult(raw []byte, path string) (gjson.Result, error) {
	if !gjson.ValidBytes(raw) {
		return gjson.Result{}, fmt.Errorf("invalid JSON")
	}
	return gjson.GetBytes(raw, path), nil
}

func gjsonTypeToString(t gjson.Type) string {
	switch t {
	case gjson.Null:
		return "null"
	case gjson.False, gjson.True:
		return "boolean"
	case gjson.Number:
		return "number"
	case gjson.String:
		return "string"
	case gjson.JSON:
		return "object/array"
	default:
		return "unknown"
	}
}

func debugValue(res gjson.Result) string {
	if res.Raw != "" {
		return res.Raw
	}
	return fmt.Sprintf("%v", res.Value())
}

func toInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	default:
		return 0, false
	}
}

func toUint64(v any) (uint64, bool) {
	switch val := v.(type) {
	case uint:
		return uint64(val), true
	case uint8:
		return uint64(val), true
	case uint16:
		return uint64(val), true
	case uint32:
		return uint64(val), true
	case uint64:
		return val, true
	default:
		return 0, false
	}
}

func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

func isEmptyJSONResult(res gjson.Result) bool {
	if !res.Exists() {
		return true
	}
	if res.Type == gjson.Null {
		return true
	}

	switch res.Type {
	case gjson.String:
		return strings.TrimSpace(res.String()) == ""
	case gjson.JSON:
		if res.IsArray() {
			return len(res.Array()) == 0
		}
		if res.IsObject() {
			return len(res.Map()) == 0
		}
		return false
	default:
		return false
	}
}

func compareJSONResult(res gjson.Result, expected any) (bool, string) {
	if expected == nil {
		if !res.Exists() {
			return false, "field does not exist (expected null)"
		}
		if res.Type != gjson.Null {
			return false, fmt.Sprintf("expected null, got %s: %s", gjsonTypeToString(res.Type), debugValue(res))
		}
		return true, ""
	}

	switch exp := expected.(type) {
	case string:
		if res.Type != gjson.String {
			return false, fmt.Sprintf("expected string %q, got %s: %s", exp, gjsonTypeToString(res.Type), debugValue(res))
		}
		actual := res.String()
		if actual != exp {
			return false, fmt.Sprintf("expected %q, got %q", exp, actual)
		}
		return true, ""

	case bool:
		if res.Type != gjson.True && res.Type != gjson.False {
			return false, fmt.Sprintf("expected boolean %v, got %s: %s", exp, gjsonTypeToString(res.Type), debugValue(res))
		}
		actual := res.Bool()
		if actual != exp {
			return false, fmt.Sprintf("expected %v, got %v", exp, actual)
		}
		return true, ""

	default:
		if expectedInt, ok := toInt64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedInt, gjsonTypeToString(res.Type), debugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedInt, actualFloat)
			}
			actualInt := res.Int()
			if actualInt != expectedInt {
				return false, fmt.Sprintf("expected %d, got %d", expectedInt, actualInt)
			}
			return true, ""
		}

		if expectedUint, ok := toUint64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedUint, gjsonTypeToString(res.Type), debugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedUint, actualFloat)
			}
			actualUint := res.Uint()
			if actualUint != expectedUint {
				return false, fmt.Sprintf("expected %d, got %d", expectedUint, actualUint)
			}
			return true, ""
		}

		if expectedFloat, ok := toFloat64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %v, got %s: %s", expectedFloat, gjsonTypeToString(res.Type), debugValue(res))
			}
			actualFloat := res.Float()
			if actualFloat != expectedFloat {
				return false, fmt.Sprintf("expected %v, got %v", expectedFloat, actualFloat)
			}
			return true, ""
		}

		return false, fmt.Sprintf("unsupported expected type %T; supported: string/bool/int*/uint*/float*/nil", expected)
	}
}

func (c *Call[TReq, TResp]) ExpectResponseStatus(code int) *Call[TReq, TResp] {
	c.addExpectation(makeResponseStatusExpectation(code))
	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyNotEmpty() *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyNotEmptyExpectation())
	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldNotEmpty(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldNotEmptyExpectation(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldValue(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldValueExpectation(path, expected))
	return c
}

func makeResponseStatusExpectation(code int) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect response status %d %s", code, http.StatusText(code)),
		func(err error, resp *client.Response[any]) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Request failed",
				}
			}
			if resp == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response is nil",
				}
			}
			if resp.NetworkError != "" {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Network error occurred",
				}
			}
			if resp.StatusCode != code {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected status %d %s, got %d %s", code, http.StatusText(code), resp.StatusCode, http.StatusText(resp.StatusCode)),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, resp *client.Response[any], checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect response status %d %s] %s", code, http.StatusText(code), checkRes.Reason)
				return
			}
			if resp != nil {
				a.Equal(code, resp.StatusCode, "[Expect response status %d %s]", code, http.StatusText(code))
			}
		},
	)
}

func makeResponseBodyNotEmptyExpectation() *expect.Expectation[*client.Response[any]] {
	return expect.New(
		"Expect response body not empty",
		func(err error, resp *client.Response[any]) expect.CheckResult {
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Request failed",
				}
			}
			if resp == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response is nil",
				}
			}
			if resp.NetworkError != "" {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Network error occurred",
				}
			}
			if len(resp.RawBody) == 0 {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response body is empty",
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, resp *client.Response[any], checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect response body not empty] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect response body not empty]")
			}
		},
	)
}

func makeResponseBodyFieldNotEmptyExpectation(path string) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect JSON field not empty: %s", path),
		func(err error, resp *client.Response[any]) expect.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}

			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Request failed",
				}
			}
			if resp == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response is nil",
				}
			}
			if resp.NetworkError != "" {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Network error occurred",
				}
			}
			if len(resp.RawBody) == 0 {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response body is empty",
				}
			}

			res, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}

			if !res.Exists() {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist yet", path),
				}
			}

			if isEmptyJSONResult(res) {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' is empty", path),
				}
			}

			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, resp *client.Response[any], checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect JSON field not empty: %s] %s", path, checkRes.Reason)
			} else {
				a.True(true, "[Expect JSON field '%s' not empty]", path)
			}
		},
	)
}

func makeResponseBodyFieldValueExpectation(path string, expected any) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect JSON field '%s' == %v", path, expected),
		func(err error, resp *client.Response[any]) expect.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}

			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Request failed",
				}
			}
			if resp == nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response is nil",
				}
			}
			if resp.NetworkError != "" {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Network error occurred",
				}
			}
			if len(resp.RawBody) == 0 {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Response body is empty",
				}
			}

			res, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}

			if !res.Exists() {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Path '%s' does not exist in response yet", path),
				}
			}

			ok, msg := compareJSONResult(res, expected)
			if !ok {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    msg,
				}
			}

			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, resp *client.Response[any], checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect JSON field '%s' == %v] %s", path, expected, checkRes.Reason)
				return
			}

			if resp != nil && len(resp.RawBody) > 0 {
				res, parseErr := getJSONResult(resp.RawBody, path)
				if parseErr == nil && res.Exists() {
					actualValue := debugValue(res)
					a.True(true, "[Expect JSON field '%s' == %v] actual: %s", path, expected, actualValue)
				} else {
					a.True(true, "[Expect JSON field '%s' == %v]", path, expected)
				}
			} else {
				a.True(true, "[Expect JSON field '%s' == %v]", path, expected)
			}
		},
	)
}
