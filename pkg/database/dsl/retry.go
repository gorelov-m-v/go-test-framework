package dsl

import (
	"context"
	"strconv"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/extension"
)

func (q *Query[T]) executeWithRetry(stepCtx provider.StepCtx, expectations []*expectation) (T, error, extension.PollingSummary) {
	cfg := q.asyncCfg

	if !cfg.Enabled {
		return q.executeSingle()
	}

	ctx := q.ctx
	deadline := time.Now().Add(cfg.Timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	ctxWithDeadline, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

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
		err := sqlscan.Get(ctxWithDeadline, q.client.DB, &result, q.sql, q.args...)
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

		if time.Now().After(deadline) || ctxWithDeadline.Err() != nil {
			elapsed := time.Since(startTime)
			summary := extension.PollingSummary{
				Attempts:      retryCtx.Attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.FailedReasons,
				TimeoutReason: "Timeout or context cancelled",
			}
			finalErr := err
			if ctxWithDeadline.Err() != nil {
				finalErr = ctxWithDeadline.Err()
				summary.LastError = finalErr.Error()
			} else if err != nil {
				summary.LastError = err.Error()
			}
			return result, finalErr, summary
		}

		delay := retryCtx.CalculateNextDelay()

		if len(retryCtx.FailedReasons) > 0 {
			reasonsPreview := sanitizeForLog(retryCtx.FailedReasons[0])
			if len(retryCtx.FailedReasons) > 1 {
				reasonsPreview += " (and " + strconv.Itoa(len(retryCtx.FailedReasons)-1) + " more)"
			}
			stepCtx.Logf("Retry attempt %d: %d failed check(s), delay %s. Reason: %s",
				retryCtx.Attempt, len(retryCtx.FailedReasons), delay, reasonsPreview)
		}

		select {
		case <-time.After(delay):
		case <-ctxWithDeadline.Done():
			elapsed := time.Since(startTime)
			summary := extension.PollingSummary{
				Attempts:      retryCtx.Attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.FailedReasons,
				TimeoutReason: "Context cancelled",
			}
			finalErr := ctxWithDeadline.Err()
			if finalErr != nil {
				summary.LastError = finalErr.Error()
			}
			return result, finalErr, summary
		}
	}
}

func (q *Query[T]) executeSingle() (T, error, extension.PollingSummary) {
	startTime := time.Now()

	var result T
	err := sqlscan.Get(q.ctx, q.client.DB, &result, q.sql, q.args...)

	elapsed := time.Since(startTime)
	summary := extension.PollingSummary{
		Attempts:    1,
		ElapsedTime: elapsed.String(),
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

func sanitizeForLog(reason string) string {
	const maxLength = 80
	const keepChars = 20

	if len(reason) <= maxLength {
		return reason
	}

	return reason[:keepChars] + "... [truncated for security]"
}
