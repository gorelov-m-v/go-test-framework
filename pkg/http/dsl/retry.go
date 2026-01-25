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
		Ctx:          c.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  c.client.AsyncConfig,
		Expectations: expectations,

		Executor: func(ctx context.Context) (*client.Response[TResp], error) {
			return client.DoTyped[TReq, TResp](ctx, c.client, c.req)
		},

		Convert: func(resp *client.Response[TResp]) *client.Response[any] {
			return &client.Response[any]{
				StatusCode:   resp.StatusCode,
				Headers:      resp.Headers,
				RawBody:      resp.RawBody,
				Error:        resp.Error,
				Duration:     resp.Duration,
				NetworkError: resp.NetworkError,
			}
		},

		PostProcess: func(resp *client.Response[TResp], err error, summary *polling.PollingSummary) {
			retry.PostProcessNetworkError(resp, err, summary)
		},

		NilResultFactory: func(err error) *client.Response[TResp] {
			msg := constants.ErrNilResponse
			if err != nil {
				msg = err.Error()
			}
			return &client.Response[TResp]{NetworkError: msg}
		},
	})
}
