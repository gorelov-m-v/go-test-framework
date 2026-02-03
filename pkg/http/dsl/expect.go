package dsl

import (
	"fmt"
	"net/http"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/errors"
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

var preCheck = client.BuildPreCheck()
var preCheckWithBody = client.BuildPreCheckWithBody()

var jsonSource = &expect.JSONExpectationSource[*client.Response[any]]{
	GetJSON:          func(r *client.Response[any]) ([]byte, error) { return r.RawBody, nil },
	PreCheck:         preCheck,
	PreCheckWithBody: preCheckWithBody,
}

func (c *Call[TReq, TResp]) ExpectResponseStatus(code int) *Call[TReq, TResp] {
	c.addExpectation(makeResponseStatusExpectation(code))
	return c
}

func (c *Call[TReq, TResp]) ExpectBodyNotEmpty() *Call[TReq, TResp] {
	c.addExpectation(makeResponseBodyNotEmptyExpectation())
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldNotEmpty(path string) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldNotEmpty(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldEquals(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldEquals(path, expected))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldIsNull(path string) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldIsNull(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldIsNotNull(path string) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldIsNotNull(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldTrue(path string) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldTrue(path))
	return c
}

func (c *Call[TReq, TResp]) ExpectFieldFalse(path string) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.FieldFalse(path))
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

func (c *Call[TReq, TResp]) ExpectMatchesContract() *Call[TReq, TResp] {
	if c.sent {
		c.stepCtx.Break(errors.MethodAfterSend("HTTP", "ExpectMatchesContract"))
		c.stepCtx.BrokenNow()
		return c
	}
	c.validateContract = true
	return c
}

func (c *Call[TReq, TResp]) ExpectMatchesSchema(schemaName string) *Call[TReq, TResp] {
	if c.sent {
		c.stepCtx.Break(errors.MethodAfterSend("HTTP", "ExpectMatchesSchema"))
		c.stepCtx.BrokenNow()
		return c
	}
	c.contractSchema = schemaName
	return c
}

func (c *Call[TReq, TResp]) ExpectArrayContains(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.ArrayContains(path, expected))
	return c
}

func (c *Call[TReq, TResp]) ExpectArrayContainsExact(path string, expected any) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.ArrayContainsExact(path, expected))
	return c
}

func (c *Call[TReq, TResp]) ExpectBodyEquals(expected any) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.BodyEquals(expected))
	return c
}

func (c *Call[TReq, TResp]) ExpectBodyPartial(expected any) *Call[TReq, TResp] {
	c.addExpectation(jsonSource.BodyPartial(expected))
	return c
}
