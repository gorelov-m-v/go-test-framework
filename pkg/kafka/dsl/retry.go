package dsl

import (
	"context"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
)

func (e *Expectation) executeSingle() ([]byte, bool, polling.PollingSummary) {
	executor := func(ctx context.Context) ([]byte, error) {
		return e.doSearch()
	}

	result, err, summary := retry.ExecuteSingle(context.Background(), executor)

	if err != nil {
		summary.Success = false
		summary.LastError = err.Error()
		return nil, false, summary
	}

	summary.Success = true
	return result, true, summary
}

func (e *Expectation) executeWithRetry(stepCtx provider.StepCtx) ([]byte, bool, polling.PollingSummary) {
	asyncCfg := e.kafkaClient.GetAsyncConfig()

	executor := func(ctx context.Context) ([]byte, error) {
		return e.doSearch()
	}

	checker := func(result []byte, err error) []polling.CheckResult {
		if err != nil {
			return []polling.CheckResult{{
				Ok:        false,
				Retryable: true,
				Reason:    err.Error(),
			}}
		}

		return []polling.CheckResult{{
			Ok:        true,
			Retryable: false,
		}}
	}

	result, err, summary := retry.ExecuteWithRetry(
		context.Background(),
		stepCtx,
		asyncCfg,
		executor,
		checker,
	)

	if err != nil {
		return nil, false, summary
	}

	return result, true, summary
}
