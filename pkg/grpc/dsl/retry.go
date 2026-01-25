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
		Ctx:          c.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  c.client.AsyncConfig,
		Expectations: expectations,

		Executor: func(ctx context.Context) (*client.Response[TResp], error) {
			return client.Invoke[TReq, TResp](ctx, c.client, c.fullMethod, c.body, c.metadata)
		},

		Convert: func(resp *client.Response[TResp]) *client.Response[any] {
			respAny := &client.Response[any]{
				Metadata: resp.Metadata,
				Duration: resp.Duration,
				Error:    resp.Error,
				RawBody:  resp.RawBody,
			}
			if resp.Body != nil {
				var bodyAny any = resp.Body
				respAny.Body = &bodyAny
			}
			return respAny
		},

		PostProcess: func(resp *client.Response[TResp], err error, summary *polling.PollingSummary) {
			retry.PostProcessSummary(resp, err, summary)
		},

		NilResultFactory: func(err error) *client.Response[TResp] {
			errMsg := errors.New(constants.ErrNilResponse)
			if err != nil {
				errMsg = err
			}
			return &client.Response[TResp]{Error: errMsg}
		},
	})
}
