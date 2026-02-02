package dsl

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

var preCheck = client.BuildPreCheck()
var preCheckWithBody = client.BuildPreCheckWithBody()

var jsonSource = &expect.JSONExpectationSource[*client.Response[any]]{
	GetJSON:          getResponseJSON,
	PreCheck:         preCheck,
	PreCheckWithBody: preCheckWithBody,
}

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

func (c *Call[TReq, TResp]) ExpectFieldEquals(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldEquals(path, expected))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldNotEmpty(path string) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldNotEmpty(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldExists(path string) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldExists(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectMetadata(key, value string) *Call[TReq, TResp] {
	c.addExpectation(makeMetadataExpectation(key, value))
	return c
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
		if err := jsonutil.ValidateBytes(resp.RawBody); err == nil {
			return resp.RawBody, nil
		}
	}

	jsonBytes, err := json.Marshal(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return jsonBytes, nil
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
