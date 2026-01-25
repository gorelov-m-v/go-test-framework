package dsl

import (
	"context"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/constants"
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

func (q *Query) execute(
	stepCtx provider.StepCtx,
	expectations []*expect.Expectation[*client.Result],
) (*client.Result, error, polling.PollingSummary) {
	return retry.ExecuteDSLSimple(retry.DSLConfig[*client.Result, *client.Result]{
		Ctx:          q.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  q.client.AsyncConfig,
		Expectations: expectations,

		Executor: func(ctx context.Context) (*client.Result, error) {
			result := q.client.Get(ctx, q.key)
			if result.Exists {
				ttlResult := q.client.TTL(ctx, q.key)
				result.TTL = ttlResult.TTL
			}
			return result, result.Error
		},

		PostProcess: func(result *client.Result, err error, summary *polling.PollingSummary) {
			retry.PostProcessSummary(result, err, summary)
		},

		NilResultFactory: func(err error) *client.Result {
			return &client.Result{
				Key:   q.key,
				Error: fmt.Errorf("%s: %w", constants.ErrNilResult, err),
			}
		},
	})
}
