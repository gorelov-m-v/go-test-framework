package dsl

import (
	"context"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"google.golang.org/grpc/metadata"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/validation"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

// Call represents a gRPC request builder with fluent interface.
// It supports unary RPC calls with typed request/response, metadata,
// expectations, and automatic retry in async mode.
//
// Type parameters:
//   - TReq: Request protobuf message type
//   - TResp: Response protobuf message type
//
// Example:
//
//	dsl.NewCall[pb.GetUserRequest, pb.GetUserResponse](sCtx, grpcClient).
//	    Method("/user.UserService/GetUser").
//	    RequestBody(pb.GetUserRequest{Id: "123"}).
//	    ExpectNoError().
//	    ExpectFieldEquals("name", "John").
//	    Send()
type Call[TReq any, TResp any] struct {
	sCtx   provider.StepCtx
	client *client.Client
	ctx    context.Context

	fullMethod string
	body       *TReq
	metadata   metadata.MD

	resp *client.Response[TResp]
	sent bool

	expectations []*expect.Expectation[*client.Response[any]]
}

// NewCall creates a new gRPC request builder.
//
// Parameters:
//   - sCtx: Allure step context for test reporting
//   - grpcClient: gRPC client configured with target address
//
// Returns a Call builder that can be configured with method, request, and expectations.
func NewCall[TReq any, TResp any](sCtx provider.StepCtx, grpcClient *client.Client) *Call[TReq, TResp] {
	return &Call[TReq, TResp]{
		sCtx:     sCtx,
		client:   grpcClient,
		ctx:      context.Background(),
		metadata: metadata.MD{},
	}
}

// Method sets the full gRPC method path in format "/package.Service/Method".
func (c *Call[TReq, TResp]) Method(fullMethod string) *Call[TReq, TResp] {
	c.fullMethod = fullMethod
	return c
}

// RequestBody sets the protobuf request message.
func (c *Call[TReq, TResp]) RequestBody(body TReq) *Call[TReq, TResp] {
	c.body = &body
	return c
}

// Metadata adds a metadata key-value pair to the gRPC call.
func (c *Call[TReq, TResp]) Metadata(key, value string) *Call[TReq, TResp] {
	c.metadata.Append(key, value)
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

// Send executes the gRPC call and validates all expectations.
// In async mode (AsyncStep), automatically retries with backoff until expectations pass.
// Returns the response containing the protobuf message, metadata, and any error.
func (c *Call[TReq, TResp]) Send() *client.Response[TResp] {
	c.validate()

	c.sCtx.WithNewStep(c.stepName(), func(stepCtx provider.StepCtx) {
		attachRequest(stepCtx, c)

		resp, err, summary := c.execute(stepCtx, c.expectations)
		c.resp = resp
		c.sent = true

		c.attachResults(stepCtx, summary)
		c.assertResults(stepCtx, err)
	})

	return c.resp
}

func (c *Call[TReq, TResp]) stepName() string {
	return fmt.Sprintf("gRPC %s", c.fullMethod)
}

func (c *Call[TReq, TResp]) attachResults(stepCtx provider.StepCtx, summary polling.PollingSummary) {
	polling.AttachIfAsync(stepCtx, summary)
	attachResponse(stepCtx, c.resp)
}

func (c *Call[TReq, TResp]) assertResults(stepCtx provider.StepCtx, err error) {
	expect.AssertExpectations(stepCtx, c.expectations, err, c.convertToAny(), c.assertNoExpectations)
}

func (c *Call[TReq, TResp]) assertNoExpectations(stepCtx provider.StepCtx, mode polling.AssertionMode, err error) {
	if err != nil {
		polling.NoError(stepCtx, mode, err, "gRPC call failed: %v", err)
	}
}

func (c *Call[TReq, TResp]) convertToAny() *client.Response[any] {
	respAny := &client.Response[any]{
		Metadata: c.resp.Metadata,
		Duration: c.resp.Duration,
		Error:    c.resp.Error,
		RawBody:  c.resp.RawBody,
	}
	if c.resp.Body != nil {
		var bodyAny any = c.resp.Body
		respAny.Body = &bodyAny
	}
	return respAny
}

func (c *Call[TReq, TResp]) validate() {
	v := validation.New(c.sCtx, "gRPC")
	v.RequireNotNil(c.client, "gRPC client")
	v.RequireNotEmptyWithHint(c.fullMethod, "gRPC method", "Use .Method(\"/package.Service/Method\").")
}
