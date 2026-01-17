package dsl

import (
	"context"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
)

func (q *Query[T]) executeWithRetry(stepCtx provider.StepCtx, expectations []*expect.Expectation[T]) (T, error, polling.PollingSummary) {
	cfg := q.client.AsyncConfig

	if !cfg.Enabled {
		return q.executeSingle()
	}

	executor := func(ctx context.Context) (T, error) {
		var result T
		err := q.client.DB.GetContext(ctx, &result, q.sql, q.args...)
		return result, err
	}

	checker := func(result T, err error) []polling.CheckResult {
		results := make([]polling.CheckResult, 0, len(expectations))
		for _, exp := range expectations {
			checkRes := exp.Check(err, result)
			results = append(results, polling.CheckResult{
				Ok:        checkRes.Ok,
				Retryable: checkRes.Retryable,
				Reason:    checkRes.Reason,
			})
		}
		return results
	}

	return retry.ExecuteWithRetry(q.ctx, stepCtx, cfg, executor, checker)
}

func (q *Query[T]) executeSingle() (T, error, polling.PollingSummary) {
	executor := func(ctx context.Context) (T, error) {
		var result T
		err := q.client.DB.GetContext(ctx, &result, q.sql, q.args...)
		return result, err
	}

	return retry.ExecuteSingle(q.ctx, executor)
}
