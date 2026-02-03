package polling

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type StepMode int

const (
	SyncMode StepMode = iota
	AsyncMode
	CleanupMode // Uses Assert (non-fatal) - errors logged but don't stop other cleanup steps
)

type StepModeProvider interface {
	provider.StepCtx
	StepMode() StepMode
}

func GetStepMode(stepCtx provider.StepCtx) StepMode {
	if smp, ok := stepCtx.(StepModeProvider); ok {
		return smp.StepMode()
	}

	return SyncMode
}

func GetAssertionModeFromStepMode(stepMode StepMode) AssertionMode {
	if stepMode == AsyncMode || stepMode == CleanupMode {
		return AssertionAssert
	}
	return AssertionRequire
}
