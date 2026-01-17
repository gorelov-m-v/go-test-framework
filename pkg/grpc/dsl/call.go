package dsl

import (
	"context"
	"fmt"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"google.golang.org/grpc/metadata"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

type Call[TReq any, TResp any] struct {
	sCtx   provider.StepCtx
	client *client.Client
	ctx    context.Context

	stepName string

	service    string
	method     string
	fullMethod string
	body       *TReq
	metadata   metadata.MD

	resp *client.Response[TResp]
	sent bool

	expectations []*expect.Expectation[*client.Response[any]]
}

func NewCall[TReq any, TResp any](sCtx provider.StepCtx, grpcClient *client.Client) *Call[TReq, TResp] {
	return &Call[TReq, TResp]{
		sCtx:     sCtx,
		client:   grpcClient,
		ctx:      context.Background(),
		metadata: metadata.MD{},
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

func (c *Call[TReq, TResp]) Method(fullMethod string) *Call[TReq, TResp] {
	c.fullMethod = fullMethod
	parts := strings.Split(fullMethod, "/")
	if len(parts) >= 3 {
		c.service = parts[1]
		c.method = parts[2]
	}
	return c
}

func (c *Call[TReq, TResp]) Service(service string) *Call[TReq, TResp] {
	c.service = service
	c.updateFullMethod()
	return c
}

func (c *Call[TReq, TResp]) MethodName(method string) *Call[TReq, TResp] {
	c.method = method
	c.updateFullMethod()
	return c
}

func (c *Call[TReq, TResp]) updateFullMethod() {
	if c.service != "" && c.method != "" {
		c.fullMethod = fmt.Sprintf("/%s/%s", c.service, c.method)
	}
}

func (c *Call[TReq, TResp]) RequestBody(body TReq) *Call[TReq, TResp] {
	c.body = &body
	return c
}

func (c *Call[TReq, TResp]) Metadata(key, value string) *Call[TReq, TResp] {
	c.metadata.Append(key, value)
	return c
}

func (c *Call[TReq, TResp]) MetadataMap(md map[string]string) *Call[TReq, TResp] {
	for k, v := range md {
		c.metadata.Append(k, v)
	}
	return c
}

func (c *Call[TReq, TResp]) addExpectation(exp *expect.Expectation[*client.Response[any]]) {
	if c.sent {
		c.sCtx.Break("gRPC DSL Error: Expectations must be added before Send(). Call ExpectNoError(), ExpectFieldValue(), etc. before Send().")
		c.sCtx.BrokenNow()
		return
	}
	c.expectations = append(c.expectations, exp)
}

func (c *Call[TReq, TResp]) Send() *client.Response[TResp] {
	c.validate()

	name := c.stepName
	if name == "" {
		name = fmt.Sprintf("gRPC %s", c.fullMethod)
	}

	c.sCtx.WithNewStep(name, func(stepCtx provider.StepCtx) {
		attachRequest(stepCtx, c)

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
			resp = &client.Response[TResp]{Error: fmt.Errorf("nil response")}
			if err == nil {
				err = fmt.Errorf("unexpected nil response")
			}
		}

		c.resp = resp
		c.sent = true

		if mode == polling.AsyncMode {
			polling.AttachPollingSummary(stepCtx, summary)
		}

		attachResponse(stepCtx, resp)

		var bodyAny any
		if resp.Body != nil {
			bodyAny = resp.Body
		}
		respAny := &client.Response[any]{
			Body:     &bodyAny,
			Metadata: resp.Metadata,
			Duration: resp.Duration,
			Error:    resp.Error,
			RawBody:  resp.RawBody,
		}

		assertionMode := polling.GetAssertionModeFromStepMode(mode)

		if len(c.expectations) == 0 {
			if err != nil {
				polling.NoError(stepCtx, assertionMode, err, "gRPC call failed: %v", err)
				return
			}
			return
		}

		expect.ReportAll(stepCtx, assertionMode, c.expectations, err, respAny)
	})

	return c.resp
}

func (c *Call[TReq, TResp]) validate() {
	if c.client == nil {
		c.sCtx.Break("gRPC DSL Error: gRPC client is nil. Check test configuration.")
		c.sCtx.BrokenNow()
		return
	}
	if strings.TrimSpace(c.fullMethod) == "" {
		c.sCtx.Break("gRPC DSL Error: gRPC method is not set. Use .Method(\"/package.Service/Method\") or .Service().MethodName().")
		c.sCtx.BrokenNow()
		return
	}
}
