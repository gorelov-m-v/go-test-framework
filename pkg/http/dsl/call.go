package dsl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

// Call represents an HTTP request builder with fluent interface.
// It supports all HTTP methods, request/response body typing, headers,
// path/query parameters, expectations, and automatic retry in async mode.
//
// Type parameters:
//   - TReq: Request body type (use any for requests without body)
//   - TResp: Response body type for automatic JSON deserialization
//
// Example:
//
//	dsl.NewCall[CreateUserReq, CreateUserResp](sCtx, httpClient).
//	    POST("/api/users").
//	    RequestBody(CreateUserReq{Name: "John"}).
//	    ExpectResponseStatus(201).
//	    ExpectFieldNotEmpty("id").
//	    Send()
type Call[TReq any, TResp any] struct {
	sCtx   provider.StepCtx
	client *client.Client
	ctx    context.Context

	req  *client.Request[TReq]
	resp *client.Response[TResp]

	sent             bool
	expectations     []*expect.Expectation[*client.Response[any]]
	validateContract bool
	contractSchema   string
}

// NewCall creates a new HTTP request builder.
//
// Parameters:
//   - sCtx: Allure step context for test reporting
//   - httpClient: HTTP client configured with base URL and settings
//
// Returns a Call builder that can be configured with HTTP method, path, and expectations.
func NewCall[TReq any, TResp any](sCtx provider.StepCtx, httpClient *client.Client) *Call[TReq, TResp] {
	return &Call[TReq, TResp]{
		sCtx:   sCtx,
		client: httpClient,
		ctx:    context.Background(),
		req: &client.Request[TReq]{
			Headers:     make(map[string]string),
			PathParams:  make(map[string]string),
			QueryParams: make(map[string]string),
		},
	}
}

// GET sets the HTTP method to GET and specifies the request path.
func (c *Call[TReq, TResp]) GET(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodGet, path
	return c
}

// POST sets the HTTP method to POST and specifies the request path.
func (c *Call[TReq, TResp]) POST(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodPost, path
	return c
}

// PUT sets the HTTP method to PUT and specifies the request path.
func (c *Call[TReq, TResp]) PUT(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodPut, path
	return c
}

// PATCH sets the HTTP method to PATCH and specifies the request path.
func (c *Call[TReq, TResp]) PATCH(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodPatch, path
	return c
}

// DELETE sets the HTTP method to DELETE and specifies the request path.
func (c *Call[TReq, TResp]) DELETE(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodDelete, path
	return c
}

// Header adds an HTTP header to the request.
func (c *Call[TReq, TResp]) Header(key, value string) *Call[TReq, TResp] {
	c.req.Headers[key] = value
	return c
}

// PathParam adds a path parameter that will replace {key} or :key in the URL path.
func (c *Call[TReq, TResp]) PathParam(key, value string) *Call[TReq, TResp] {
	c.req.PathParams[key] = value
	return c
}

// QueryParam adds a query parameter to the request URL.
func (c *Call[TReq, TResp]) QueryParam(key, value string) *Call[TReq, TResp] {
	c.req.QueryParams[key] = value
	return c
}

// RequestBody sets the typed request body that will be serialized to JSON.
func (c *Call[TReq, TResp]) RequestBody(body TReq) *Call[TReq, TResp] {
	c.req.Body = &body
	return c
}

// RequestBodyMap sets the request body as a map for untyped requests.
// Useful for negative tests with missing fields, extra fields, or wrong types.
func (c *Call[TReq, TResp]) RequestBodyMap(body map[string]interface{}) *Call[TReq, TResp] {
	c.req.BodyMap = body
	return c
}

func (c *Call[TReq, TResp]) addExpectation(exp *expect.Expectation[*client.Response[any]]) {
	if c.sent {
		c.sCtx.Break("HTTP DSL Error: Expectations must be added before Send(). Call ExpectResponseStatus(), ExpectResponseBodyNotEmpty(), etc. before Send().")
		c.sCtx.BrokenNow()
		return
	}
	c.expectations = append(c.expectations, exp)
}

// Send executes the HTTP request and validates all expectations.
// In async mode (AsyncStep), automatically retries with backoff until expectations pass.
// Returns the response containing status code, headers, and deserialized body.
func (c *Call[TReq, TResp]) Send() *client.Response[TResp] {
	c.validate()
	c.validateContractConfig()

	c.sCtx.WithNewStep(c.stepName(), func(stepCtx provider.StepCtx) {
		attachRequest(stepCtx, c.client, c.req)

		resp, err, summary := c.execute(stepCtx, c.expectations)
		c.resp = resp
		c.sent = true

		c.attachResults(stepCtx, summary)
		c.assertResults(stepCtx, err)
		c.performContractValidation(stepCtx, c.resp)
	})

	return c.resp
}

func (c *Call[TReq, TResp]) stepName() string {
	return fmt.Sprintf("%s %s", c.req.Method, c.req.Path)
}

func (c *Call[TReq, TResp]) attachResults(stepCtx provider.StepCtx, summary polling.PollingSummary) {
	polling.AttachIfAsync(stepCtx, summary)
	attachResponse(stepCtx, c.client, c.resp)
}

func (c *Call[TReq, TResp]) assertResults(stepCtx provider.StepCtx, err error) {
	expect.AssertExpectations(stepCtx, c.expectations, err, c.convertToAny(), c.assertNoExpectations)
}

func (c *Call[TReq, TResp]) assertNoExpectations(stepCtx provider.StepCtx, mode polling.AssertionMode, err error) {
	if err != nil {
		polling.NoError(stepCtx, mode, err, "HTTP request failed: %v", err)
		return
	}
	if c.resp.NetworkError != "" {
		polling.Equal(stepCtx, mode, "", c.resp.NetworkError, "HTTP network error")
	}
}

func (c *Call[TReq, TResp]) convertToAny() *client.Response[any] {
	return &client.Response[any]{
		StatusCode:   c.resp.StatusCode,
		Headers:      c.resp.Headers,
		RawBody:      c.resp.RawBody,
		Error:        c.resp.Error,
		Duration:     c.resp.Duration,
		NetworkError: c.resp.NetworkError,
	}
}

func (c *Call[TReq, TResp]) validate() {
	if c.client == nil {
		c.sCtx.Break("HTTP DSL Error: HTTP client is nil. Check test configuration.")
		c.sCtx.BrokenNow()
		return
	}
	if c.req == nil {
		c.sCtx.Break("HTTP DSL Error: HTTP request is nil. This is an internal error.")
		c.sCtx.BrokenNow()
		return
	}
	if strings.TrimSpace(c.req.Method) == "" {
		c.sCtx.Break("HTTP DSL Error: HTTP method is not set. Use .GET(), .POST(), .PUT(), .PATCH(), or .DELETE().")
		c.sCtx.BrokenNow()
		return
	}
	if strings.TrimSpace(c.req.Path) == "" {
		c.sCtx.Break("HTTP DSL Error: HTTP path is not set. Provide path in method call like .GET(\"/api/users\").")
		c.sCtx.BrokenNow()
		return
	}
}

func (c *Call[TReq, TResp]) validateContractConfig() {
	if !c.validateContract && c.contractSchema == "" {
		return
	}

	if c.client.ContractValidator == nil {
		c.sCtx.Break("HTTP DSL Error: Contract validation requested but no contractSpec configured for this client. Add 'contractSpec' to your HTTP client config.")
		c.sCtx.BrokenNow()
		return
	}
}

func (c *Call[TReq, TResp]) performContractValidation(stepCtx provider.StepCtx, resp *client.Response[TResp]) {
	if !c.validateContract && c.contractSchema == "" {
		return
	}

	if c.client.ContractValidator == nil || resp == nil {
		return
	}

	if resp.NetworkError != "" {
		return
	}

	var validationErr error

	if c.contractSchema != "" {
		validationErr = c.client.ContractValidator.ValidateResponseBySchema(c.contractSchema, resp.RawBody)
	} else if c.validateContract {
		path := c.req.Path
		for key, value := range c.req.PathParams {
			path = strings.ReplaceAll(path, "{"+key+"}", value)
			path = strings.ReplaceAll(path, ":"+key, value)
		}
		if c.client.ContractBasePath != "" {
			path = c.client.ContractBasePath + path
		}
		validationErr = c.client.ContractValidator.ValidateResponse(c.req.Method, path, resp.StatusCode, resp.RawBody)
	}

	if validationErr != nil {
		stepCtx.Require().NoError(validationErr, "Contract validation failed")
	}
}
