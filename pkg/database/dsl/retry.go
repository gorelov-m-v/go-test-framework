package dsl

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/config"
)

type pollingSummary struct {
	Attempts      int      `json:"attempts"`
	ElapsedTime   string   `json:"elapsed_time"`
	Success       bool     `json:"success"`
	LastError     string   `json:"last_error,omitempty"`
	FailedChecks  []string `json:"failed_checks,omitempty"`
	TimeoutReason string   `json:"timeout_reason,omitempty"`
}

type retryContext struct {
	attempt       int
	cfg           config.AsyncConfig
	deadline      time.Time
	currentDelay  time.Duration
	lastErr       error
	lastResult    any
	failedReasons []string
}

func (q *Query[T]) executeWithRetry(stepCtx provider.StepCtx, expectations []*expectation) (T, error, pollingSummary) {
	cfg := config.GetAsyncConfig()

	if !cfg.Enabled {
		// If async is disabled, fall back to single execution
		return q.executeSingle(stepCtx)
	}

	ctx := q.ctx
	deadline := time.Now().Add(cfg.Timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	retryCtx := &retryContext{
		attempt:      0,
		cfg:          cfg,
		deadline:     deadline,
		currentDelay: cfg.Interval,
	}

	startTime := time.Now()

	for {
		retryCtx.attempt++

		var result T
		err := sqlscan.Get(ctx, q.client.DB, &result, q.sql, q.args...)
		retryCtx.lastErr = err
		retryCtx.lastResult = result

		allOk := true
		hasRetryable := false
		retryCtx.failedReasons = []string{}

		for _, exp := range expectations {
			checkRes := exp.check(err, result)
			if !checkRes.ok {
				allOk = false
				retryCtx.failedReasons = append(retryCtx.failedReasons, checkRes.reason)
				if checkRes.retryable {
					hasRetryable = true
				}
			}
		}

		if allOk {
			elapsed := time.Since(startTime)
			summary := pollingSummary{
				Attempts:    retryCtx.attempt,
				ElapsedTime: elapsed.String(),
				Success:     true,
			}
			return result, nil, summary
		}

		if !hasRetryable {
			elapsed := time.Since(startTime)
			summary := pollingSummary{
				Attempts:      retryCtx.attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.failedReasons,
				TimeoutReason: "Non-retryable error encountered",
			}
			if err != nil {
				summary.LastError = err.Error()
			}
			return result, err, summary
		}

		if time.Now().After(deadline) || ctx.Err() != nil {
			elapsed := time.Since(startTime)
			summary := pollingSummary{
				Attempts:      retryCtx.attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.failedReasons,
				TimeoutReason: "Timeout or context cancelled",
			}
			if err != nil {
				summary.LastError = err.Error()
			}
			return result, err, summary
		}

		delay := retryCtx.calculateNextDelay()
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			elapsed := time.Since(startTime)
			summary := pollingSummary{
				Attempts:      retryCtx.attempt,
				ElapsedTime:   elapsed.String(),
				Success:       false,
				FailedChecks:  retryCtx.failedReasons,
				TimeoutReason: "Context cancelled",
			}
			return result, ctx.Err(), summary
		}
	}
}

func (q *Query[T]) executeSingle(stepCtx provider.StepCtx) (T, error, pollingSummary) {
	var result T
	err := sqlscan.Get(q.ctx, q.client.DB, &result, q.sql, q.args...)

	summary := pollingSummary{
		Attempts:    1,
		ElapsedTime: "0s",
		Success:     err == nil,
	}
	if err != nil {
		summary.LastError = err.Error()
	}

	return result, err, summary
}

func (rc *retryContext) calculateNextDelay() time.Duration {
	delay := rc.currentDelay

	if rc.cfg.Backoff.Enabled && rc.attempt > 1 {
		delay = time.Duration(float64(delay) * rc.cfg.Backoff.Factor)
		if delay > rc.cfg.Backoff.MaxInterval {
			delay = rc.cfg.Backoff.MaxInterval
		}
		rc.currentDelay = delay
	}

	if rc.cfg.Jitter > 0 {
		jitterAmount := float64(delay) * rc.cfg.Jitter
		jitterDelta := (rand.Float64()*2 - 1) * jitterAmount
		delay = time.Duration(float64(delay) + jitterDelta)
		if delay < 0 {
			delay = rc.cfg.Interval
		}
	}

	return delay
}

func attachPollingSummary(stepCtx provider.StepCtx, summary pollingSummary) {
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	stepCtx.WithNewAttachment("Polling Summary", allure.JSON, summaryJSON)
}

func reportExpectations(stepCtx provider.StepCtx, mode assertMode, expectations []*expectation, err error, result any) {
	for _, exp := range expectations {
		checkRes := exp.check(err, result)
		exp.report(stepCtx, mode, err, result, checkRes)
	}
}

func finalFailureMessage(summary pollingSummary) string {
	if len(summary.FailedChecks) == 0 {
		return "DB query expectations not met within timeout"
	}

	msg := fmt.Sprintf("DB query expectations not met after %d attempts (%s):\n", summary.Attempts, summary.ElapsedTime)
	for i, reason := range summary.FailedChecks {
		msg += fmt.Sprintf("  [%d] %s\n", i+1, reason)
	}
	if summary.LastError != "" {
		msg += fmt.Sprintf("Last error: %s", summary.LastError)
	}
	return msg
}
