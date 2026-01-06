package dsl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/expect"
	"go-test-framework/pkg/extension"
	"go-test-framework/pkg/http/client"
)

type Call[TReq any, TResp any] struct {
	sCtx   provider.StepCtx
	client *client.Client
	ctx    context.Context

	stepName string

	req  *client.Request[TReq]
	resp *client.Response[TResp]

	sent         bool
	expectations []*expect.Expectation[*client.Response[any]]
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

func (c *Call[TReq, TResp]) StepName(name string) *Call[TReq, TResp] {
	c.stepName = strings.TrimSpace(name)
	return c
}

func (c *Call[TReq, TResp]) Context(ctx context.Context) *Call[TReq, TResp] {
	if ctx != nil {
		c.ctx = ctx
	}
	return c
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

func (c *Call[TReq, TResp]) addExpectation(exp *expect.Expectation[*client.Response[any]]) {
	if c.sent {
		panic("httpdsl: expectations must be added before RequestSend()")
	}
	c.expectations = append(c.expectations, exp)
}

func (c *Call[TReq, TResp]) RequestSend() *Call[TReq, TResp] {
	c.validate()

	name := c.stepName
	if name == "" {
		name = fmt.Sprintf("%s %s", c.req.Method, c.req.Path)
	}

	c.sCtx.WithNewStep(name, func(stepCtx provider.StepCtx) {
		attachRequest(stepCtx, c.client, c.req)

		mode := extension.GetStepMode(stepCtx)
		// IMPORTANT: Retries are only enabled in AsyncMode when expectations are present.
		// Without expectations, requests execute once even in AsyncMode (no automatic retry on network errors).
		// To enable retries: add at least one expectation (ExpectResponseStatus, ExpectResponseBodyNotEmpty, etc.)
		useRetry := mode == extension.AsyncMode && len(c.expectations) > 0

		var (
			resp    *client.Response[TResp]
			err     error
			summary extension.PollingSummary
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

		if mode == extension.AsyncMode {
			extension.AttachPollingSummary(stepCtx, summary)
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

		assertionMode := extension.GetAssertionModeFromStepMode(mode)

		if len(c.expectations) == 0 {
			if err != nil {
				extension.NoError(stepCtx, assertionMode, err, "HTTP request failed: %v", err)
				return
			}
			if c.resp.NetworkError != "" {
				extension.Equal(stepCtx, assertionMode, "", c.resp.NetworkError, "HTTP network error")
				return
			}
			return
		}

		expect.ReportAll(stepCtx, assertionMode, c.expectations, err, respAny)
	})

	return c
}

func (c *Call[TReq, TResp]) Response() *client.Response[TResp] {
	return c.resp
}

func (c *Call[TReq, TResp]) validate() {
	if c.client == nil {
		panic("httpdsl: client is nil")
	}
	if c.req == nil {
		panic("httpdsl: request is nil")
	}
	if strings.TrimSpace(c.req.Method) == "" {
		panic("httpdsl: HTTP method is not set")
	}
	if strings.TrimSpace(c.req.Path) == "" {
		panic("httpdsl: HTTP path is not set")
	}
}
