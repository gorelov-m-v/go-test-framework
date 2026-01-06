package dsl

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/extension"
	"go-test-framework/pkg/retry"
)

func (q *Query[T]) executeWithRetry(stepCtx provider.StepCtx, expectations []*expectation) (T, error, extension.PollingSummary) {
	cfg := q.asyncCfg

	if !cfg.Enabled {
		return q.executeSingle()
	}

	executor := func(ctx context.Context) (T, error) {
		var result T
		err := sqlscan.Get(ctx, q.client.DB, &result, q.sql, q.args...)
		return result, err
	}

	checker := func(result T, err error) []retry.CheckResult {
		results := make([]retry.CheckResult, 0, len(expectations))
		for _, exp := range expectations {
			checkRes := exp.check(err, result)
			results = append(results, retry.CheckResult{
				Ok:        checkRes.ok,
				Retryable: checkRes.retryable,
				Reason:    checkRes.reason,
			})
		}
		return results
	}

	return retry.ExecuteWithRetry(q.ctx, stepCtx, cfg, executor, checker)
}

func (q *Query[T]) executeSingle() (T, error, extension.PollingSummary) {
	executor := func(ctx context.Context) (T, error) {
		var result T
		err := sqlscan.Get(ctx, q.client.DB, &result, q.sql, q.args...)
		return result, err
	}

	return retry.ExecuteSingle(q.ctx, executor)
}

func reportExpectations(stepCtx provider.StepCtx, mode extension.AssertionMode, expectations []*expectation, err error, result any) {
	for _, exp := range expectations {
		checkRes := exp.check(err, result)
		exp.report(stepCtx, mode, err, result, checkRes)
	}
}
