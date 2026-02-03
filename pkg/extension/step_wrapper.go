package extension

import (
	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type stepCtxWrapper struct {
	provider.StepCtx
	mode StepMode
}

func (w *stepCtxWrapper) StepMode() StepMode {
	return w.mode
}

func (w *stepCtxWrapper) WithNewStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	w.StepCtx.WithNewStep(stepName, func(sCtx provider.StepCtx) {
		wrappedCtx := &stepCtxWrapper{
			StepCtx: sCtx,
			mode:    w.mode,
		}
		step(wrappedCtx)
	}, params...)
}

func (w *stepCtxWrapper) WithNewAsyncStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	w.StepCtx.WithNewAsyncStep(stepName, func(sCtx provider.StepCtx) {
		wrappedCtx := &stepCtxWrapper{
			StepCtx: sCtx,
			mode:    w.mode,
		}
		step(wrappedCtx)
	}, params...)
}

func WithAsyncMode(sCtx provider.StepCtx) provider.StepCtx {
	if wrapped, ok := sCtx.(*stepCtxWrapper); ok {
		return &stepCtxWrapper{
			StepCtx: wrapped.StepCtx,
			mode:    AsyncMode,
		}
	}
	return &stepCtxWrapper{
		StepCtx: sCtx,
		mode:    AsyncMode,
	}
}

func WithSyncMode(sCtx provider.StepCtx) provider.StepCtx {
	if wrapped, ok := sCtx.(*stepCtxWrapper); ok {
		return &stepCtxWrapper{
			StepCtx: wrapped.StepCtx,
			mode:    SyncMode,
		}
	}
	return &stepCtxWrapper{
		StepCtx: sCtx,
		mode:    SyncMode,
	}
}

// WithCleanupMode wraps step context in cleanup mode.
// CleanupMode uses Assert (non-fatal assertions) so errors don't stop other cleanup steps.
func WithCleanupMode(sCtx provider.StepCtx) provider.StepCtx {
	if wrapped, ok := sCtx.(*stepCtxWrapper); ok {
		return &stepCtxWrapper{
			StepCtx: wrapped.StepCtx,
			mode:    CleanupMode,
		}
	}
	return &stepCtxWrapper{
		StepCtx: sCtx,
		mode:    CleanupMode,
	}
}
