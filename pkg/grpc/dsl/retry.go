package dsl

import (
	"context"
	"errors"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/constants"
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

func (c *Call[TReq, TResp]) execute(
	stepCtx provider.StepCtx,
	expectations []*expect.Expectation[*client.Response[any]],
) (*client.Response[TResp], error, polling.PollingSummary) {
	return retry.ExecuteDSL(retry.DSLConfig[*client.Response[TResp], *client.Response[any]]{
		Ctx:              c.ctx,
		StepCtx:          stepCtx,
		AsyncConfig:      c.client.AsyncConfig,
		Expectations:     expectations,
		Executor:         c.doRequest,
		Convert:          func(resp *client.Response[TResp]) *client.Response[any] { return resp.ToAny() },
		PostProcess:      postProcessGRPC[TResp],
		NilResultFactory: newGRPCErrorResponse[TResp],
	})
}

func (c *Call[TReq, TResp]) doRequest(ctx context.Context) (*client.Response[TResp], error) {
	return client.Invoke[TReq, TResp](ctx, c.client, c.fullMethod, c.body, c.metadata)
}

func postProcessGRPC[TResp any](resp *client.Response[TResp], err error, summary *polling.PollingSummary) {
	retry.PostProcessSummary(resp, err, summary)
}

func newGRPCErrorResponse[TResp any](err error) *client.Response[TResp] {
	errMsg := errors.New(constants.ErrNilResponse)
	if err != nil {
		errMsg = err
	}
	return &client.Response[TResp]{Error: errMsg}
}
