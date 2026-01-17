package polling

import "github.com/ozontech/allure-go/pkg/framework/provider"

type AssertionMode int

const (
	AssertionRequire AssertionMode = iota
	AssertionAssert
)

func PickAsserter(stepCtx provider.StepCtx, mode AssertionMode) provider.Asserts {
	if mode == AssertionAssert {
		return stepCtx.Assert()
	}
	return stepCtx.Require()
}

func NoError(stepCtx provider.StepCtx, mode AssertionMode, err error, msgAndArgs ...any) {
	a := PickAsserter(stepCtx, mode)
	a.NoError(err, msgAndArgs...)
}

func True(stepCtx provider.StepCtx, mode AssertionMode, condition bool, msgAndArgs ...any) {
	a := PickAsserter(stepCtx, mode)
	a.True(condition, msgAndArgs...)
}

func Equal(stepCtx provider.StepCtx, mode AssertionMode, expected, actual any, msgAndArgs ...any) {
	a := PickAsserter(stepCtx, mode)
	a.Equal(expected, actual, msgAndArgs...)
}
