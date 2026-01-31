package dsl

import (
	"context"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
)

func (q *Query[T]) execute(
	stepCtx provider.StepCtx,
	expectations []*expect.Expectation[T],
) (T, time.Duration, error, polling.PollingSummary) {
	var lastDuration time.Duration

	result, err, summary := retry.ExecuteDSLSimple(retry.DSLConfig[T, T]{
		Ctx:          q.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  q.client.AsyncConfig,
		Expectations: expectations,
		Executor:     q.timedQuery(&lastDuration),
	})

	return result, lastDuration, err, summary
}

func (q *Query[T]) executeAll(stepCtx provider.StepCtx) ([]T, time.Duration, error, polling.PollingSummary) {
	var lastDuration time.Duration

	results, err, summary := retry.ExecuteDSLSimple(retry.DSLConfig[[]T, []T]{
		Ctx:          q.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  q.client.AsyncConfig,
		Expectations: q.expectationsAll,
		Executor:     q.timedQueryAll(&lastDuration),
	})

	return results, lastDuration, err, summary
}

func (q *Query[T]) timedQuery(durationPtr *time.Duration) func(context.Context) (T, error) {
	return func(ctx context.Context) (T, error) {
		start := time.Now()
		var result T
		err := q.client.DB.GetContext(ctx, &result, q.sql, q.args...)
		*durationPtr = time.Since(start)
		return result, err
	}
}

func (q *Query[T]) timedQueryAll(durationPtr *time.Duration) func(context.Context) ([]T, error) {
	return func(ctx context.Context) ([]T, error) {
		start := time.Now()
		var results []T
		err := q.client.DB.SelectContext(ctx, &results, q.sql, q.args...)
		*durationPtr = time.Since(start)
		return results, err
	}
}
