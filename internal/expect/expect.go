package expect

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type Expectation[T any] struct {
	Name   string
	Check  func(err error, value T) polling.CheckResult
	Report func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, value T, res polling.CheckResult)
}

func New[T any](
	name string,
	check func(err error, value T) polling.CheckResult,
	report func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, value T, res polling.CheckResult),
) *Expectation[T] {
	return &Expectation[T]{
		Name:   name,
		Check:  check,
		Report: report,
	}
}

func ReportAll[T any](
	stepCtx provider.StepCtx,
	mode polling.AssertionMode,
	exps []*Expectation[T],
	err error,
	value T,
) {
	for _, exp := range exps {
		checkRes := exp.Check(err, value)
		exp.Report(stepCtx, mode, err, value, checkRes)
	}
}

// AddExpectation appends an expectation to the slice, validating that Send() hasn't been called yet.
// Returns true if added successfully, false if already sent (and breaks the test).
func AddExpectation[T any](
	sCtx provider.StepCtx,
	sent bool,
	expectations *[]*Expectation[T],
	exp *Expectation[T],
	dslName string,
) bool {
	if sent {
		sCtx.Break(dslName + " DSL Error: Expectations must be added before Send().")
		sCtx.BrokenNow()
		return false
	}
	*expectations = append(*expectations, exp)
	return true
}
