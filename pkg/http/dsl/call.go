package dsl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go-test-framework/pkg/http/client"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type AssertionMode int

const (
	AssertionsRequire AssertionMode = iota
	AssertionsAssert
)

type Call[TReq any, TResp any] struct {
	sCtx          provider.StepCtx
	client        *client.Client
	ctx           context.Context
	assertionMode AssertionMode

	stepName string

	req  *client.Request[TReq]
	resp *client.Response[TResp]

	sent         bool
	expectations []func(parent provider.StepCtx)
}

func NewCall[TReq any, TResp any](sCtx provider.StepCtx, httpClient *client.Client) *Call[TReq, TResp] {
	return &Call[TReq, TResp]{
		sCtx:          sCtx,
		client:        httpClient,
		ctx:           context.Background(),
		assertionMode: AssertionsRequire,
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

func (c *Call[TReq, TResp]) Assert() *Call[TReq, TResp] {
	c.assertionMode = AssertionsAssert
	return c
}

func (c *Call[TReq, TResp]) Require() *Call[TReq, TResp] {
	c.assertionMode = AssertionsRequire
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

func (c *Call[TReq, TResp]) addExpectation(fn func(parent provider.StepCtx)) {
	if c.sent {
		panic("httpdsl: expectations must be added before RequestSend()")
	}
	c.expectations = append(c.expectations, fn)
}

func (c *Call[TReq, TResp]) RequestSend() *Call[TReq, TResp] {
	c.validate()

	name := c.stepName
	if name == "" {
		name = fmt.Sprintf("%s %s", c.req.Method, c.req.Path)
	}

	c.sCtx.WithNewStep(name, func(stepCtx provider.StepCtx) {
		attachRequest(stepCtx, c.client, c.req)

		resp, err := client.DoTyped[TReq, TResp](c.ctx, c.client, c.req)
		if err != nil && resp == nil {
			resp = &client.Response[TResp]{NetworkError: err.Error()}
		}
		c.resp = resp

		attachResponse(stepCtx, c.client, c.resp)

		c.sent = true
		for _, fn := range c.expectations {
			fn(stepCtx)
		}
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

func (c *Call[TReq, TResp]) pickAsserter(ctx provider.StepCtx) provider.Asserts {
	if c.assertionMode == AssertionsAssert {
		return ctx.Assert()
	}
	return ctx.Require()
}
