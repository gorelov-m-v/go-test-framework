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

func (c *Call[TReq, TResp]) GET(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodGet, path
	return c
}

func (c *Call[TReq, TResp]) POST(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodPost, path
	return c
}

func (c *Call[TReq, TResp]) PUT(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodPut, path
	return c
}

func (c *Call[TReq, TResp]) PATCH(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodPatch, path
	return c
}

func (c *Call[TReq, TResp]) DELETE(path string) *Call[TReq, TResp] {
	c.req.Method, c.req.Path = http.MethodDelete, path
	return c
}

func (c *Call[TReq, TResp]) Header(key, value string) *Call[TReq, TResp] {
	c.req.Headers[key] = value
	return c
}

func (c *Call[TReq, TResp]) PathParam(key, value string) *Call[TReq, TResp] {
	c.req.PathParams[key] = value
	return c
}

func (c *Call[TReq, TResp]) QueryParam(key, value string) *Call[TReq, TResp] {
	c.req.QueryParams[key] = value
	return c
}

func (c *Call[TReq, TResp]) RequestBody(body TReq) *Call[TReq, TResp] {
	c.req.Body = &body
	return c
}

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

func (c *Call[TReq, TResp]) Send() *client.Response[TResp] {
	c.validate()
	c.validateContractConfig()

	stepName := fmt.Sprintf("%s %s", c.req.Method, c.req.Path)

	c.sCtx.WithNewStep(stepName, func(stepCtx provider.StepCtx) {
		attachRequest(stepCtx, c.client, c.req)

		mode := polling.GetStepMode(stepCtx)
		useRetry := mode == polling.AsyncMode && len(c.expectations) > 0

		var (
			resp    *client.Response[TResp]
			err     error
			summary polling.PollingSummary
		)

		if useRetry {
			resp, err, summary = c.executeWithRetry(stepCtx, c.expectations)
		} else {
			resp, err, summary = c.executeSingle()
		}

		if resp == nil {
			resp = &client.Response[TResp]{NetworkError: "nil response"}
			if err == nil {
				err = fmt.Errorf("unexpected nil response")
			}
		}

		c.resp = resp
		c.sent = true

		if mode == polling.AsyncMode {
			polling.AttachPollingSummary(stepCtx, summary)
		}

		attachResponse(stepCtx, c.client, c.resp)

		respAny := &client.Response[any]{
			StatusCode:   resp.StatusCode,
			Headers:      resp.Headers,
			RawBody:      resp.RawBody,
			Error:        resp.Error,
			Duration:     resp.Duration,
			NetworkError: resp.NetworkError,
		}

		assertionMode := polling.GetAssertionModeFromStepMode(mode)

		if len(c.expectations) == 0 {
			if err != nil {
				polling.NoError(stepCtx, assertionMode, err, "HTTP request failed: %v", err)
				return
			}
			if c.resp.NetworkError != "" {
				polling.Equal(stepCtx, assertionMode, "", c.resp.NetworkError, "HTTP network error")
				return
			}
		} else {
			expect.ReportAll(stepCtx, assertionMode, c.expectations, err, respAny)
		}

		c.performContractValidation(stepCtx, resp)
	})

	return c.resp
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
