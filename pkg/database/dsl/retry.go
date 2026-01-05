package dsl

import (
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/extension"
)

func (q *Query[T]) executeWithRetry(stepCtx provider.StepCtx, expectations []*expectation) (T, error, extension.PollingSummary) {
	cfg := q.asyncCfg

	if !cfg.Enabled {
		return q.executeSingle(stepCtx)
	}

	ctx := q.ctx
	deadline := time.Now().Add(cfg.Timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	retryCtx := &extension.RetryContext{
		Attempt:      0,
		Cfg:          cfg,
		Deadline:     deadline,
		CurrentDelay: cfg.Interval,
	}

	startTime := time.Now()

	for {
		retryCtx.Attempt++

		var result T
		err := sqlscan.Get(ctx, q.client.DB, &result, q.sql, q.args...)
		retryCtx.LastErr = err
		retryCtx.LastResult = result

		allOk := true
		hasRetryable := false
		retryCtx.FailedReasons = []string{}

		for _, exp := range expectations {
			checkRes := exp.check(err, result)
			if !checkRes.ok {
				allOk = false
				retryCtx.FailedReasons = append(retryCtx.FailedReasons, checkRes.reason)
				if checkRes.retryable {
					hasRetryable = true
				}
			}
		}

		if allOk {
			elapsed := time.Since(startTime)
			summary := extension.PollingSummary{
				Attempts:    retryCtx.Attempt,
				ElapsedTime: elapsed.String(),
				Success:     true,
			}
			return result, nil, summary
		}

		if !hasRetryable {
			elapsed := time.Since(startTime)
			summary := extension.PollingSummary{
				Attempts:      retryCtx.Attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.FailedReasons,
				TimeoutReason: "Non-retryable error encountered",
			}
			if err != nil {
				summary.LastError = err.Error()
			}
			return result, err, summary
		}

		if time.Now().After(deadline) || ctx.Err() != nil {
			elapsed := time.Since(startTime)
			summary := extension.PollingSummary{
				Attempts:      retryCtx.Attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.FailedReasons,
				TimeoutReason: "Timeout or context cancelled",
			}
			if err != nil {
				summary.LastError = err.Error()
			}
			return result, err, summary
		}

		delay := retryCtx.CalculateNextDelay()
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			elapsed := time.Since(startTime)
			summary := extension.PollingSummary{
				Attempts:      retryCtx.Attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.FailedReasons,
				TimeoutReason: "Context cancelled",
			}
			return result, ctx.Err(), summary
		}
	}
}

func (q *Query[T]) executeSingle(stepCtx provider.StepCtx) (T, error, extension.PollingSummary) {
	var result T
	err := sqlscan.Get(q.ctx, q.client.DB, &result, q.sql, q.args...)

	summary := extension.PollingSummary{
		Attempts:    1,
		ElapsedTime: "0s",
		Success:     err == nil,
	}
	if err != nil {
		summary.LastError = err.Error()
	}

	return result, err, summary
}

func reportExpectations(stepCtx provider.StepCtx, mode assertMode, expectations []*expectation, err error, result any) {
	for _, exp := range expectations {
		checkRes := exp.check(err, result)
		exp.report(stepCtx, mode, err, result, checkRes)
	}
}
