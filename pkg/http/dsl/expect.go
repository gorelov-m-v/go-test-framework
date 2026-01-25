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

func validateJSONPath(path string) error {
	if path == "" {
		return fmt.Errorf("JSON path cannot be empty")
	}
	return nil
}

// ExpectResponseStatus checks that the HTTP response status code equals the expected value.
func (c *Call[TReq, TResp]) ExpectResponseStatus(code int) *Call[TReq, TResp] {
	c.addExpectation(makeResponseStatusExpectation(code))
	return c
}

// ExpectBodyNotEmpty checks that the response body is not empty.
func (c *Call[TReq, TResp]) ExpectBodyNotEmpty() *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyNotEmptyExpectation())
	return c
}

// Deprecated: Use ExpectBodyNotEmpty instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyNotEmpty() *Call[TReq, TResp] {
	return c.ExpectBodyNotEmpty()
}

// ExpectFieldNotEmpty checks that a JSON field at the given GJSON path is not empty.
// Path uses GJSON syntax: "user.id", "items.0.name", "data.#".
func (c *Call[TReq, TResp]) ExpectFieldNotEmpty(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldNotEmptyExpectation(path))
	return c
}

// Deprecated: Use ExpectFieldNotEmpty instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyFieldNotEmpty(path string) *Call[TReq, TResp] {
	return c.ExpectFieldNotEmpty(path)
}

// ExpectFieldEquals checks that a JSON field at the given GJSON path equals the expected value.
// Supports numeric type coercion (int, int64, float64 are compared by value).
// Path uses GJSON syntax: "user.id", "items.0.name", "data.#".
func (c *Call[TReq, TResp]) ExpectFieldEquals(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldValueExpectation(path, expected))
	return c
}

// Deprecated: Use ExpectFieldEquals instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyFieldValue(path string, expected any) *Call[TReq, TResp] {
	return c.ExpectFieldEquals(path, expected)
}

// ExpectFieldIsNull checks that a JSON field at the given GJSON path is null.
func (c *Call[TReq, TResp]) ExpectFieldIsNull(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldIsNullExpectation(path))
	return c
}

// Deprecated: Use ExpectFieldIsNull instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyFieldIsNull(path string) *Call[TReq, TResp] {
	return c.ExpectFieldIsNull(path)
}

// ExpectFieldIsNotNull checks that a JSON field at the given GJSON path is not null.
func (c *Call[TReq, TResp]) ExpectFieldIsNotNull(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldIsNotNullExpectation(path))
	return c
}

// Deprecated: Use ExpectFieldIsNotNull instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyFieldIsNotNull(path string) *Call[TReq, TResp] {
	return c.ExpectFieldIsNotNull(path)
}

// ExpectFieldTrue checks that a JSON boolean field at the given GJSON path is true.
func (c *Call[TReq, TResp]) ExpectFieldTrue(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldTrueExpectation(path))
	return c
}

// Deprecated: Use ExpectFieldTrue instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyFieldTrue(path string) *Call[TReq, TResp] {
	return c.ExpectFieldTrue(path)
}

// ExpectFieldFalse checks that a JSON boolean field at the given GJSON path is false.
func (c *Call[TReq, TResp]) ExpectFieldFalse(path string) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyFieldFalseExpectation(path))
	return c
}

// Deprecated: Use ExpectFieldFalse instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyFieldFalse(path string) *Call[TReq, TResp] {
	return c.ExpectFieldFalse(path)
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
	return expect.BuildJSONFieldExpectation(expect.JSONFieldExpectationConfig[*client.Response[any]]{
		Path:       path,
		ExpectName: name,
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return resp.RawBody, nil },
		PreCheck: func(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}, false
			}
			return preCheckWithBody(err, resp)
		},
		Check: expect.JSONCheckNotEmpty(),
	})
}

func makeResponseBodyFieldValueExpectation(path string, expected any) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect JSON field '%s' == %v", path, expected)
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
	return expect.BuildJSONFieldNullExpectation(expect.JSONFieldNullExpectationConfig[*client.Response[any]]{
		Path:       path,
		ExpectName: name,
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return resp.RawBody, nil },
		PreCheck: func(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}, false
			}
			return preCheckWithBody(err, resp)
		},
		ExpectedNull: true,
	})
}

func makeResponseBodyFieldIsNotNullExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect JSON field '%s' is not null", path)
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
			if jsonRes.Type == gjson.Null {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
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
	return expect.BuildJSONFieldExpectation(expect.JSONFieldExpectationConfig[*client.Response[any]]{
		Path:       path,
		ExpectName: name,
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return resp.RawBody, nil },
		PreCheck: func(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}, false
			}
			return preCheckWithBody(err, resp)
		},
		Check: expect.JSONCheckTrue(),
	})
}

func makeResponseBodyFieldFalseExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect JSON field '%s' is false", path)
	return expect.BuildJSONFieldExpectation(expect.JSONFieldExpectationConfig[*client.Response[any]]{
		Path:       path,
		ExpectName: name,
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return resp.RawBody, nil },
		PreCheck: func(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
			if pathErr := validateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}, false
			}
			return preCheckWithBody(err, resp)
		},
		Check: expect.JSONCheckFalse(),
	})
}

// ExpectMatchesContract validates the response against the OpenAPI spec.
// Auto-detects the operation by HTTP method and path to find the expected response schema.
func (c *Call[TReq, TResp]) ExpectMatchesContract() *Call[TReq, TResp] {
	if c.sent {
		c.sCtx.Break("HTTP DSL Error: ExpectMatchesContract must be called before Send()")
		c.sCtx.BrokenNow()
		return c
	}
	c.validateContract = true
	return c
}

// ExpectMatchesSchema validates the response against a specific schema from the OpenAPI spec.
func (c *Call[TReq, TResp]) ExpectMatchesSchema(schemaName string) *Call[TReq, TResp] {
	if c.sent {
		c.sCtx.Break("HTTP DSL Error: ExpectMatchesSchema must be called before Send()")
		c.sCtx.BrokenNow()
		return c
	}
	c.contractSchema = schemaName
	return c
}

// ExpectArrayContains checks that an array at the given GJSON path contains an object matching expected (partial match).
func (c *Call[TReq, TResp]) ExpectArrayContains(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(makeArrayContainsExpectation(path, expected))
	return c
}

func makeArrayContainsExpectation(path string, expected any) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect array '%s' contains matching object", path),
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
					Reason:    fmt.Sprintf("Path '%s' does not exist in response", path),
				}
			}
			if !jsonRes.IsArray() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected array at '%s', got %s", path, jsonutil.TypeToString(jsonRes.Type)),
				}
			}

			idx, _ := jsonutil.FindInArray(jsonRes, expected)
			if idx == -1 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("No matching object found in array '%s'", path),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect array '%s' contains matching object] %s", path, checkRes.Reason)
				return
			}

			if resp != nil && len(resp.RawBody) > 0 {
				jsonRes, parseErr := getJSONResult(resp.RawBody, path)
				if parseErr == nil && jsonRes.Exists() {
					idx, matchedItem := jsonutil.FindInArray(jsonRes, expected)
					if idx >= 0 {
						a.True(true, "[Expect array '%s' contains matching object] Found at index %d: %s", path, idx, jsonutil.DebugValue(matchedItem))
						return
					}
				}
			}
			a.True(true, "[Expect array '%s' contains matching object]", path)
		},
	)
}

// ExpectArrayContainsExact checks that an array at the given GJSON path contains an object matching expected (exact match).
func (c *Call[TReq, TResp]) ExpectArrayContainsExact(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(makeArrayContainsExactExpectation(path, expected))
	return c
}

func makeArrayContainsExactExpectation(path string, expected any) *expect.Expectation[*client.Response[any]] {
	return expect.New(
		fmt.Sprintf("Expect array '%s' contains exact matching object", path),
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
					Reason:    fmt.Sprintf("Path '%s' does not exist in response", path),
				}
			}
			if !jsonRes.IsArray() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected array at '%s', got %s", path, jsonutil.TypeToString(jsonRes.Type)),
				}
			}

			idx, _ := jsonutil.FindInArrayExact(jsonRes, expected)
			if idx == -1 {
				partialIdx, partialItem := jsonutil.FindInArray(jsonRes, expected)
				if partialIdx >= 0 {
					_, diff := jsonutil.CompareObjectExact(partialItem, expected)
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Found similar object at index %d but exact match failed: %s", partialIdx, diff),
					}
				}
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("No matching object found in array '%s'", path),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, resp *client.Response[any], checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect array '%s' contains exact matching object] %s", path, checkRes.Reason)
				return
			}

			if resp != nil && len(resp.RawBody) > 0 {
				jsonRes, parseErr := getJSONResult(resp.RawBody, path)
				if parseErr == nil && jsonRes.Exists() {
					idx, matchedItem := jsonutil.FindInArrayExact(jsonRes, expected)
					if idx >= 0 {
						a.True(true, "[Expect array '%s' contains exact matching object] Found at index %d: %s", path, idx, jsonutil.DebugValue(matchedItem))
						return
					}
				}
			}
			a.True(true, "[Expect array '%s' contains exact matching object]", path)
		},
	)
}

// ExpectBodyEquals checks that the response body exactly matches the expected struct or map (all fields must match).
func (c *Call[TReq, TResp]) ExpectBodyEquals(expected any) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyExpectation(expected))
	return c
}

// Deprecated: Use ExpectBodyEquals instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBody(expected any) *Call[TReq, TResp] {
	return c.ExpectBodyEquals(expected)
}

// ExpectBodyPartial checks that the response body contains fields from the expected struct or map (non-zero fields only).
func (c *Call[TReq, TResp]) ExpectBodyPartial(expected any) *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyPartialExpectation(expected))
	return c
}

// Deprecated: Use ExpectBodyPartial instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectResponseBodyPartial(expected any) *Call[TReq, TResp] {
	return c.ExpectBodyPartial(expected)
}

func makeResponseBodyExpectation(expected any) *expect.Expectation[*client.Response[any]] {
	return expect.BuildFullObjectExpectation(expect.FullObjectExpectationConfig[*client.Response[any]]{
		ExpectName: "Expect response body matches (exact)",
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return resp.RawBody, nil },
		PreCheck:   preCheckWithBody,
		Expected:   expected,
		Compare:    jsonutil.CompareObjectExact,
		Retryable:  true,
	})
}

func makeResponseBodyPartialExpectation(expected any) *expect.Expectation[*client.Response[any]] {
	return expect.BuildFullObjectExpectation(expect.FullObjectExpectationConfig[*client.Response[any]]{
		ExpectName: "Expect response body matches (partial)",
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return resp.RawBody, nil },
		PreCheck:   preCheckWithBody,
		Expected:   expected,
		Compare:    jsonutil.CompareObjectPartial,
		Retryable:  true,
	})
}
