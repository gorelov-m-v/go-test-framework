package extension

import (
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type TExtension struct {
	provider.T
}

func NewTExtension(t provider.T) *TExtension {
	return &TExtension{T: t}
}

func (t *TExtension) WithNewStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	t.T.WithNewStep(stepName, func(sCtx provider.StepCtx) {
		syncCtx := WithSyncMode(sCtx)
		step(syncCtx)
	}, params...)
}

func (t *TExtension) WithNewAsyncStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	t.T.WithNewAsyncStep(stepName, func(sCtx provider.StepCtx) {
		asyncCtx := WithAsyncMode(sCtx)
		step(asyncCtx)
	}, params...)
}

// WithNewCleanupStep creates a step in cleanup mode.
// Uses Assert (non-fatal) so errors are logged but don't stop other cleanup steps.
func (t *TExtension) WithNewCleanupStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	t.T.WithNewStep(stepName, func(sCtx provider.StepCtx) {
		cleanupCtx := WithCleanupMode(sCtx)
		step(cleanupCtx)
	}, params...)
}
