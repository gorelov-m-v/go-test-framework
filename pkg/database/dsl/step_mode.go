package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type StepMode int

const (
	SyncMode StepMode = iota
	AsyncMode
)

type StepModeProvider interface {
	provider.StepCtx
	StepMode() StepMode
}

func getStepMode(sCtx provider.StepCtx) StepMode {
	if smp, ok := sCtx.(StepModeProvider); ok {
		return smp.StepMode()
	}

	return SyncMode
}
