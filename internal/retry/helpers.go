package retry

import (
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

func BuildExpectationsChecker[T any](expectations []*expect.Expectation[T]) Checker[T] {
	return func(result T, err error) []polling.CheckResult {
		results := make([]polling.CheckResult, 0, len(expectations))
		for _, exp := range expectations {
			checkRes := exp.Check(err, result)
			results = append(results, checkRes)
		}
		return results
	}
}

func BuildExpectationsCheckerWithConvert[T any, E any](
	expectations []*expect.Expectation[E],
	convert func(T) E,
) Checker[T] {
	return func(result T, err error) []polling.CheckResult {
		converted := convert(result)
		results := make([]polling.CheckResult, 0, len(expectations))
		for _, exp := range expectations {
			checkRes := exp.Check(err, converted)
			results = append(results, checkRes)
		}
		return results
	}
}

type ErrorGetter interface {
	GetError() error
}

func PostProcessSummary[T ErrorGetter](result T, err error, summary *polling.PollingSummary) {
	if err != nil {
		return
	}
	if resultErr := result.GetError(); resultErr != nil {
		summary.Success = false
		if summary.LastError == "" {
			summary.LastError = resultErr.Error()
		}
	}
}

func PostProcessNetworkError[T interface{ GetNetworkError() string }](result T, err error, summary *polling.PollingSummary) {
	if err != nil {
		return
	}
	if networkErr := result.GetNetworkError(); networkErr != "" {
		summary.Success = false
		if summary.LastError == "" {
			summary.LastError = networkErr
		}
	}
}
