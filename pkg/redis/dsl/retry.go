package dsl

import (
	"context"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

func (q *Query) executeSingle() (*client.Result, error, extension.PollingSummary) {
	executor := func(ctx context.Context) (*client.Result, error) {
		// Get value and TTL in one go
		result := q.client.Get(ctx, q.key)
		if result.Exists {
			// Also get TTL
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
) (*client.Result, error, extension.PollingSummary) {
	asyncCfg := q.client.AsyncConfig

	if !asyncCfg.Enabled {
		return q.executeSingle()
	}

	executor := func(ctx context.Context) (*client.Result, error) {
		// Get value and TTL in one go
		result := q.client.Get(ctx, q.key)
		if result.Exists {
			// Also get TTL
			ttlResult := q.client.TTL(ctx, q.key)
			result.TTL = ttlResult.TTL
		}

		if result == nil {
			return &client.Result{Key: q.key, Error: fmt.Errorf("nil result")}, fmt.Errorf("unexpected nil result")
		}

		return result, result.Error
	}

	checker := func(result *client.Result, err error) []retry.CheckResult {
		results := make([]retry.CheckResult, 0, len(expectations))
		for _, exp := range expectations {
			checkRes := exp.Check(err, result)
			results = append(results, retry.CheckResult{
				Ok:        checkRes.Ok,
				Retryable: checkRes.Retryable,
				Reason:    checkRes.Reason,
			})
		}

		return results
	}

	result, err, summary := retry.ExecuteWithRetry(q.ctx, stepCtx, asyncCfg, executor, checker)

	if err == nil && result != nil && result.Error != nil {
		if summary.Success {
			summary.Success = false
		}
		if summary.LastError == "" {
			summary.LastError = result.Error.Error()
		}
	}

	return result, err, summary
}
