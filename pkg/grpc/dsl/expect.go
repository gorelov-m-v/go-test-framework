package dsl

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

// ExpectNoError checks that the gRPC call succeeds without error.
func (c *Call[TReq, TResp]) ExpectNoError() *Call[TReq, TResp] {
	c.addExpectation(makeNoErrorExpectation())
	return c
}

// ExpectError checks that the gRPC call returns an error.
func (c *Call[TReq, TResp]) ExpectError() *Call[TReq, TResp] {
	c.addExpectation(makeErrorExpectation())
	return c
}

// ExpectStatusCode checks that the gRPC response status code equals the expected value.
func (c *Call[TReq, TResp]) ExpectStatusCode(code codes.Code) *Call[TReq, TResp] {
	c.addExpectation(makeStatusCodeExpectation(code))
	return c
}

// ExpectFieldEquals checks that a JSON field at the given GJSON path equals the expected value.
// Path uses GJSON syntax on the JSON-serialized protobuf response.
func (c *Call[TReq, TResp]) ExpectFieldEquals(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(makeFieldValueExpectation(path, expected))
	return c
}

// Deprecated: Use ExpectFieldEquals instead. Will be removed in v2.0.
func (c *Call[TReq, TResp]) ExpectFieldValue(path string, expected any) *Call[TReq, TResp] {
	return c.ExpectFieldEquals(path, expected)
}

// ExpectFieldNotEmpty checks that a JSON field at the given GJSON path is not empty.
func (c *Call[TReq, TResp]) ExpectFieldNotEmpty(path string) *Call[TReq, TResp] {
	c.addExpectation(makeFieldNotEmptyExpectation(path))
	return c
}

// ExpectFieldExists checks that a JSON field at the given GJSON path exists in the response.
func (c *Call[TReq, TResp]) ExpectFieldExists(path string) *Call[TReq, TResp] {
	c.addExpectation(makeFieldExistsExpectation(path))
	return c
}

// ExpectMetadata checks that the response metadata contains the expected key-value pair.
func (c *Call[TReq, TResp]) ExpectMetadata(key, value string) *Call[TReq, TResp] {
	c.addExpectation(makeMetadataExpectation(key, value))
	return c
}

func preCheck(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
	if err != nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Call failed with error",
		}, false
	}
	if resp == nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Response is nil",
		}, false
	}
	return polling.CheckResult{}, true
}

func preCheckWithBody(err error, resp *client.Response[any]) (polling.CheckResult, bool) {
	if err != nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Call failed with error",
		}, false
	}
	if resp == nil || resp.Body == nil {
		return polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Response body is nil",
		}, false
	}
	return polling.CheckResult{}, true
}

func makeNoErrorExpectation() *expect.Expectation[*client.Response[any]] {
	name := "Expect: No error"
	return expect.New(
		name,
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
		expect.StandardReport[*client.Response[any]](name),
	)
}

func makeErrorExpectation() *expect.Expectation[*client.Response[any]] {
	name := "Expect: Error"
	return expect.New(
		name,
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
		expect.StandardReport[*client.Response[any]](name),
	)
}

func makeStatusCodeExpectation(code codes.Code) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect: Status code %s (%d)", code.String(), code)
	return expect.New(
		name,
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
		expect.StandardReport[*client.Response[any]](name),
	)
}

func getResponseJSON(resp *client.Response[any]) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("response body is nil")
	}

	if len(resp.RawBody) > 0 {
		if gjson.ValidBytes(resp.RawBody) {
			return resp.RawBody, nil
		}
	}

	jsonBytes, err := json.Marshal(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return jsonBytes, nil
}

func makeFieldValueExpectation(path string, expected any) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect: Field '%s' = %v", path, expected)
	return expect.BuildJSONFieldExpectation(expect.JSONFieldExpectationConfig[*client.Response[any]]{
		Path:       path,
		ExpectName: name,
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return getResponseJSON(resp) },
		PreCheck:   preCheckWithBody,
		Check:      expect.JSONCheckEquals(expected),
	})
}

func makeFieldNotEmptyExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect: Field '%s' not empty", path)
	return expect.BuildJSONFieldExpectation(expect.JSONFieldExpectationConfig[*client.Response[any]]{
		Path:       path,
		ExpectName: name,
		GetJSON:    func(resp *client.Response[any]) ([]byte, error) { return getResponseJSON(resp) },
		PreCheck:   preCheckWithBody,
		Check:      expect.JSONCheckNotEmpty(),
	})
}

func makeFieldExistsExpectation(path string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect: Field '%s' exists", path)
	return expect.New(
		name,
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if res, ok := preCheckWithBody(err, resp); !ok {
				return res
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
		expect.StandardReport[*client.Response[any]](name),
	)
}

func makeMetadataExpectation(key, expectedValue string) *expect.Expectation[*client.Response[any]] {
	name := fmt.Sprintf("Expect: Metadata '%s' = '%s'", key, expectedValue)
	return expect.New(
		name,
		func(err error, resp *client.Response[any]) polling.CheckResult {
			if res, ok := preCheck(err, resp); !ok {
				return res
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
		expect.StandardReport[*client.Response[any]](name),
	)
}
