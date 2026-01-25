package dsl

import (
	"context"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
)

func (q *Query[T]) execute(
	stepCtx provider.StepCtx,
	expectations []*expect.Expectation[T],
) (T, error, polling.PollingSummary) {
	return retry.ExecuteDSLSimple(retry.DSLConfig[T, T]{
		Ctx:          q.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  q.client.AsyncConfig,
		Expectations: expectations,

		Executor: func(ctx context.Context) (T, error) {
			var result T
			err := q.client.DB.GetContext(ctx, &result, q.sql, q.args...)
			return result, err
		},
	})
}

func (q *Query[T]) executeAll(stepCtx provider.StepCtx) ([]T, error, polling.PollingSummary) {
	return retry.ExecuteDSLSimple(retry.DSLConfig[[]T, []T]{
		Ctx:          q.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  q.client.AsyncConfig,
		Expectations: q.expectationsAll,

		Executor: func(ctx context.Context) ([]T, error) {
			var results []T
			err := q.client.DB.SelectContext(ctx, &results, q.sql, q.args...)
			return results, err
		},
	})
}
