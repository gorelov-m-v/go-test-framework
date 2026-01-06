package extension

import "github.com/ozontech/allure-go/pkg/framework/provider"

// AssertionMode defines the assertion behavior (require stops test, assert continues).
type AssertionMode int

const (
	AssertionRequire AssertionMode = iota
	AssertionAssert
)

// GetAssertionModeFromStepMode converts StepMode to AssertionMode.
// AsyncMode → AssertionAssert, SyncMode → AssertionRequire
func GetAssertionModeFromStepMode(stepMode StepMode) AssertionMode {
	if stepMode == AsyncMode {
		return AssertionAssert
	}
	return AssertionRequire
}

// PickAsserter returns the appropriate asserter based on AssertionMode.
func PickAsserter(stepCtx provider.StepCtx, mode AssertionMode) provider.Asserts {
	if mode == AssertionAssert {
		return stepCtx.Assert()
	}
	return stepCtx.Require()
}

// NoError checks that error is nil using the specified assertion mode.
func NoError(stepCtx provider.StepCtx, mode AssertionMode, err error, msgAndArgs ...any) {
	a := PickAsserter(stepCtx, mode)
	a.NoError(err, msgAndArgs...)
}

// True checks that condition is true using the specified assertion mode.
func True(stepCtx provider.StepCtx, mode AssertionMode, condition bool, msgAndArgs ...any) {
	a := PickAsserter(stepCtx, mode)
	a.True(condition, msgAndArgs...)
}

// Equal checks that expected equals actual using the specified assertion mode.
func Equal(stepCtx provider.StepCtx, mode AssertionMode, expected, actual any, msgAndArgs ...any) {
	a := PickAsserter(stepCtx, mode)
	a.Equal(expected, actual, msgAndArgs...)
}
