package dsl

import (
	"fmt"
	"net/http"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

func getJSONResult(raw []byte, path string) (gjson.Result, error) {
	if !gjson.ValidBytes(raw) {
		return gjson.Result{}, fmt.Errorf("invalid JSON")
	}
	return gjson.GetBytes(raw, path), nil
}

func preCheck(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
	if err != nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Request failed",
		}, false
	}
	if resp == nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Response is nil",
		}, false
	}
	if resp.NetworkError != "" {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Network error occurred",
		}, false
	}
	return polling.CheckResult{}, true
}

func preCheckWithBody(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
	if res, ok := preCheck(err, resp); !ok {
		return res, false
	}
	if len(resp.RawBody) == 0 {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Response body is empty",
		}, false
	}
	return polling.CheckResult{}, true
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

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldIsNull(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldIsNullExpectation(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldIsNotNull(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldIsNotNullExpectation(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldTrue(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldTrueExpectation(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectResponseBodyFieldFalse(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldFalseExpectation(path))
	return c
}

func makeResponseStatusExpectation(code int) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect response status %d %s", code, http.StatusText(code)),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if res, ok := preCheck(err, resp); !ok {
				return res
			}
			if resp.StatusCode != code {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected status %d %s, got %d %s", code, http.StatusText(code), resp.StatusCode, http.StatusText(resp.StatusCode)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
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
	name := "Expect response body not empty"
	return expect.New(
		name,
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Response[any]](name),
	)
}

func makeResponseBodyFieldNotEmptyExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect JSON field '%s' not empty", path)
	return expect.New(
		name,
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
			}
			jsonRes, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist yet", path),
				}
			}
			if jsonutil.IsEmpty(jsonRes) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' is empty", path),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Response[any]](name),
	)
}

func makeResponseBodyFieldValueExpectation(path string, expected any) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect JSON field '%s' == %v", path, expected),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
			}
			jsonRes, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Path '%s' does not exist in response yet", path),
				}
			}
			ok, msg := jsonutil.Compare(jsonRes, expected)
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
				a.True(false, "[Expect JSON field '%s' == %v] %s", path, expected, checkRes.Reason)
				return
			}

			if resp != nil && len(resp.RawBody) > 0 {
				res, parseErr := getJSONResult(resp.RawBody, path)
				if parseErr == nil && res.Exists() {
					actualValue := jsonutil.DebugValue(res)
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

func makeResponseBodyFieldIsNullExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect JSON field '%s' is null", path)
	return expect.New(
		name,
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
			}
			jsonRes, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist", path),
				}
			}
			if jsonRes.Type != gjson.Null {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected null, got %s: %s", jsonutil.TypeToString(jsonRes.Type), jsonutil.DebugValue(jsonRes)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Response[any]](name),
	)
}

func makeResponseBodyFieldIsNotNullExpectation(path string) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect JSON field '%s' is not null", path),
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
			}
			jsonRes, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist", path),
				}
			}
			if jsonRes.Type == gjson.Null {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Expected non-null value, got null",
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect JSON field '%s' is not null] %s", path, checkRes.Reason)
			} else {
				res, _ := getJSONResult(resp.RawBody, path)
				a.True(true, "[Expect JSON field '%s' is not null] actual: %s", path, jsonutil.DebugValue(res))
			}
		},
	)
}

func makeResponseBodyFieldTrueExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect JSON field '%s' is true", path)
	return expect.New(
		name,
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
			}
			jsonRes, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist", path),
				}
			}
			if jsonRes.Type != gjson.True {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected true, got %s: %s", jsonutil.TypeToString(jsonRes.Type), jsonutil.DebugValue(jsonRes)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Response[any]](name),
	)
}

func makeResponseBodyFieldFalseExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect JSON field '%s' is false", path)
	return expect.New(
		name,
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
			}
			jsonRes, parseErr := getJSONResult(resp.RawBody, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist", path),
				}
			}
			if jsonRes.Type != gjson.False {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected false, got %s: %s", jsonutil.TypeToString(jsonRes.Type), jsonutil.DebugValue(jsonRes)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[*client.Response[any]](name),
	)
}
