package extension

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type StepMode = polling.StepMode

const (
	SyncMode    = polling.SyncMode
	AsyncMode   = polling.AsyncMode
	CleanupMode = polling.CleanupMode
)

type StepModeProvider = polling.StepModeProvider

func GetStepMode(sCtx provider.StepCtx) StepMode {
	return polling.GetStepMode(sCtx)
}
