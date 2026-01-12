package retry

import (
	"context"
	"strconv"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
)

type CheckResult struct {
	Ok        bool
	Retryable bool
	Reason    string
}

type Executor[T any] func(ctx context.Context) (T, error)

type Checker[T any] func(result T, err error) []CheckResult

func ExecuteWithRetry[T any](
	ctx context.Context,
	stepCtx provider.StepCtx,
	cfg config.AsyncConfig,
	executor Executor[T],
	checker Checker[T],
) (T, error, extension.PollingSummary) {
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

		result, err := executor(ctxWithDeadline)
		retryCtx.LastErr = err

		checkResults := checker(result, err)

		allOk := true
		hasRetryable := false
		retryCtx.FailedReasons = retryCtx.FailedReasons[:0]

		for _, checkRes := range checkResults {
			if !checkRes.Ok {
				allOk = false
				retryCtx.FailedReasons = append(retryCtx.FailedReasons, checkRes.Reason)
				if checkRes.Retryable {
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
			reasonsPreview := SanitizeForLog(retryCtx.FailedReasons[0])
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

func ExecuteSingle[T any](
	ctx context.Context,
	executor Executor[T],
) (T, error, extension.PollingSummary) {
	startTime := time.Now()

	result, err := executor(ctx)

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

func SanitizeForLog(reason string) string {
	const maxLength = 80
	const keepChars = 20

	runes := []rune(reason)
	if len(runes) <= maxLength {
		return reason
	}

	return string(runes[:keepChars]) + "... [truncated for security]"
}
