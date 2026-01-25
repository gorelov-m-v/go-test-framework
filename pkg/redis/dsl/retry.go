package dsl

import (
	"context"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

func (q *Query) executeSingle() (*client.Result, error, polling.PollingSummary) {
	executor := func(ctx context.Context) (*client.Result, error) {
		result := q.client.Get(ctx, q.key)
		if result.Exists {
			ttlResult := q.client.TTL(ctx, q.key)
			result.TTL = ttlResult.TTL
		}

		if result == nil {
			return &client.Result{Key: q.key, Error: fmt.Errorf("nil result")}, fmt.Errorf("unexpected nil result")
		}

		return result, result.Error
	}

	result, err, summary := retry.ExecuteSingle(q.ctx, executor)

	if err == nil && result != nil && result.Error != nil {
		summary.Success = false
		summary.LastError = result.Error.Error()
	}

	return result, err, summary
}

func (q *Query) executeWithRetry(
	stepCtx provider.StepCtx,
	expectations []*expect.Expectation[*client.Result],
) (*client.Result, error, polling.PollingSummary) {
	asyncCfg := q.client.AsyncConfig

	if !asyncCfg.Enabled {
		return q.executeSingle()
	}

	executor := func(ctx context.Context) (*client.Result, error) {
		result := q.client.Get(ctx, q.key)
		if result.Exists {
			ttlResult := q.client.TTL(ctx, q.key)
			result.TTL = ttlResult.TTL
		}

		if result == nil {
			return &client.Result{Key: q.key, Error: fmt.Errorf("nil result")}, fmt.Errorf("unexpected nil result")
		}

		return result, result.Error
	}

	result, err, summary := retry.ExecuteWithRetry(
		q.ctx,
		stepCtx,
		asyncCfg,
		executor,
		retry.BuildExpectationsChecker(expectations),
	)

	retry.PostProcessSummary(result, err, &summary)

	return result, err, summary
}
