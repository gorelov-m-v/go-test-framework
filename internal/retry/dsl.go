package retry

import (
	"context"
	"fmt"
	"reflect"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/constants"
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

// DSLConfig contains configuration for executing DSL operations with retry support.
// It abstracts the common retry logic used across all DSL packages (HTTP, gRPC, Kafka, DB, Redis).
//
// Type parameters:
//   - TResult: The type returned by the executor function
//   - TExpect: The type used for expectations (may differ from TResult)
type DSLConfig[TResult any, TExpect any] struct {
	Ctx         context.Context
	StepCtx     provider.StepCtx
	AsyncConfig config.AsyncConfig

	Executor     func(ctx context.Context) (TResult, error)
	Expectations []*expect.Expectation[TExpect]

	Convert          func(TResult) TExpect
	Checker          Checker[TResult]
	PostProcess      func(result TResult, err error, summary *polling.PollingSummary)
	NilResultFactory func(err error) TResult
}

// ExecuteDSL executes a DSL operation with optional retry support.
// In async mode with expectations, automatically retries with backoff until all expectations pass.
// Returns the result, any error, and a polling summary for reporting.
func ExecuteDSL[TResult any, TExpect any](cfg DSLConfig[TResult, TExpect]) (TResult, error, polling.PollingSummary) {
	mode := polling.GetStepMode(cfg.StepCtx)
	hasExpectations := len(cfg.Expectations) > 0 || cfg.Checker != nil
	useRetry := mode == polling.AsyncMode && hasExpectations && cfg.AsyncConfig.Enabled

	safeExecutor := wrapExecutor(cfg.Executor, cfg.NilResultFactory)

	var result TResult
	var err error
	var summary polling.PollingSummary

	if useRetry {
		checker := buildCheckerFromConfig(cfg)
		result, err, summary = ExecuteWithRetry(cfg.Ctx, cfg.StepCtx, cfg.AsyncConfig, safeExecutor, checker)
	} else {
		result, err, summary = ExecuteSingle(cfg.Ctx, safeExecutor)
	}

	if cfg.PostProcess != nil {
		cfg.PostProcess(result, err, &summary)
	}

	return result, err, summary
}

// ExecuteDSLSimple is a convenience wrapper for ExecuteDSL when TResult and TExpect are the same type.
func ExecuteDSLSimple[T any](cfg DSLConfig[T, T]) (T, error, polling.PollingSummary) {
	return ExecuteDSL(cfg)
}

func wrapExecutor[T any](executor func(context.Context) (T, error), nilFactory func(error) T) Executor[T] {
	return func(ctx context.Context) (T, error) {
		result, err := executor(ctx)

		if isNilResult(result) {
			if nilFactory != nil {
				result = nilFactory(err)
			}
			if err == nil {
				err = fmt.Errorf(constants.ErrUnexpectedNil)
			}
		}

		return result, err
	}
}

func isNilResult[T any](result T) bool {
	v := reflect.ValueOf(result)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan:
		return v.IsNil()
	}
	return false
}

func buildCheckerFromConfig[TResult any, TExpect any](cfg DSLConfig[TResult, TExpect]) Checker[TResult] {
	if cfg.Checker != nil {
		return cfg.Checker
	}

	if cfg.Convert != nil {
		return BuildExpectationsCheckerWithConvert(cfg.Expectations, cfg.Convert)
	}

	if len(cfg.Expectations) > 0 {
		var zeroResult TResult
		if _, ok := any(zeroResult).(TExpect); !ok {
			var zeroExpect TExpect
			cfg.StepCtx.Break(fmt.Sprintf(
				"DSL Configuration Error: TResult (%T) != TExpect (%T) but Convert function not provided. "+
					"This is a framework bug - please report it.",
				zeroResult, zeroExpect,
			))
			cfg.StepCtx.BrokenNow()
			return nil
		}
	}

	return func(result TResult, err error) []polling.CheckResult {
		var expectResult TExpect
		if v, ok := any(result).(TExpect); ok {
			expectResult = v
		}

		results := make([]polling.CheckResult, 0, len(cfg.Expectations))
		for _, exp := range cfg.Expectations {
			results = append(results, exp.Check(err, expectResult))
		}
		return results
	}
}
