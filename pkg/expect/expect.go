package expect

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/extension"
)

type CheckResult struct {
	Ok        bool
	Retryable bool
	Reason    string
}

type Expectation[T any] struct {
	Name   string
	Check  func(err error, value T) CheckResult
	Report func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, value T, res CheckResult)
}

func New[T any](
	name string,
	check func(err error, value T) CheckResult,
	report func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, value T, res CheckResult),
) *Expectation[T] {
	return &Expectation[T]{
		Name:   name,
		Check:  check,
		Report: report,
	}
}

func ReportAll[T any](
	stepCtx provider.StepCtx,
	mode extension.AssertionMode,
	exps []*Expectation[T],
	err error,
	value T,
) {
	for _, exp := range exps {
		checkRes := exp.Check(err, value)
		exp.Report(stepCtx, mode, err, value, checkRes)
	}
}
