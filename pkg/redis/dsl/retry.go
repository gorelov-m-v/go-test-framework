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
		Ctx:              q.ctx,
		StepCtx:          stepCtx,
		AsyncConfig:      q.client.AsyncConfig,
		Expectations:     expectations,
		Executor:         q.doQuery,
		PostProcess:      postProcessRedis,
		NilResultFactory: q.newRedisErrorResult,
	})
}

func (q *Query) doQuery(ctx context.Context) (*client.Result, error) {
	result := q.client.Get(ctx, q.key)
	if result.Exists {
		ttlResult := q.client.TTL(ctx, q.key)
		result.TTL = ttlResult.TTL
	}
	return result, result.Error
}

func postProcessRedis(result *client.Result, err error, summary *polling.PollingSummary) {
	retry.PostProcessSummary(result, err, summary)
}

func (q *Query) newRedisErrorResult(err error) *client.Result {
	return &client.Result{
		Key:   q.key,
		Error: fmt.Errorf("%s: %w", constants.ErrNilResult, err),
	}
}
