package dsl

import (
	"context"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
)

func (q *Query[T]) execute(stepCtx provider.StepCtx) ([]byte, bool, error, polling.PollingSummary) {
	result, err, summary := retry.ExecuteDSL(retry.DSLConfig[[]byte, []byte]{
		Ctx:          q.ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  q.client.GetAsyncConfig(),
		Expectations: q.expectations,

		Executor: func(ctx context.Context) ([]byte, error) {
			return q.doSearch()
		},

		Checker: q.buildChecker(),
	})

	if err != nil {
		return nil, false, err, summary
	}

	return result, true, nil, summary
}

func (q *Query[T]) buildChecker() retry.Checker[[]byte] {
	return func(result []byte, err error) []polling.CheckResult {
		if err != nil {
			return []polling.CheckResult{{
				Ok:        false,
				Retryable: true,
				Reason:    err.Error(),
			}}
		}

		if len(q.expectations) == 0 {
			return []polling.CheckResult{{Ok: true}}
		}

		results := make([]polling.CheckResult, 0, len(q.expectations))
		for _, exp := range q.expectations {
			checkRes := exp.Check(err, result)
			results = append(results, checkRes)
		}
		return results
	}
}
