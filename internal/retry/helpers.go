package retry

import (
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

// BuildExpectationsChecker creates a Checker function from a list of expectations.
// The returned checker evaluates all expectations against the result.
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

// BuildExpectationsCheckerWithConvert creates a Checker that converts the result type before evaluation.
// Useful when the executor returns a different type than what expectations check against.
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

// ErrorGetter is implemented by response types that may contain an error field.
// Used by PostProcessSummary to detect errors stored in the result rather than returned.
type ErrorGetter interface {
	GetError() error
}

// NetworkErrorGetter is implemented by HTTP response types that may contain network errors.
// Used by PostProcessNetworkError to detect network failures stored in the result.
type NetworkErrorGetter interface {
	GetNetworkError() string
}

// PostProcessSummary updates the polling summary if the result contains an error.
// Call this after execution to ensure errors stored in the result are reflected in the summary.
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

// PostProcessNetworkError updates the polling summary if the result contains a network error.
// Call this after HTTP execution to ensure network errors are reflected in the summary.
func PostProcessNetworkError[T NetworkErrorGetter](result T, err error, summary *polling.PollingSummary) {
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
