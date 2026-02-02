package dsl

import (
	"context"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/constants"
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
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
		PostProcess:      postProcessHTTP[TResp],
		NilResultFactory: newHTTPErrorResponse[TResp],
	})
}

func (c *Call[TReq, TResp]) doRequest(ctx context.Context) (*client.Response[TResp], error) {
	return client.DoTyped[TReq, TResp](ctx, c.client, c.req)
}

func postProcessHTTP[TResp any](resp *client.Response[TResp], err error, summary *polling.PollingSummary) {
	retry.PostProcessSummary(resp, err, summary)
}

func newHTTPErrorResponse[TResp any](err error) *client.Response[TResp] {
	msg := constants.ErrNilResponse
	if err != nil {
		msg = err.Error()
	}
	return &client.Response[TResp]{NetworkError: msg}
}
