package expect

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type OnEmptyFunc func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error)

func AssertExpectations[T any](
	stepCtx provider.StepCtx,
	expectations []*Expectation[T],
	err error,
	value T,
	onEmpty OnEmptyFunc,
) {
	mode := polling.GetStepMode(stepCtx)
	assertionMode := polling.GetAssertionModeFromStepMode(mode)

	if len(expectations) == 0 {
		if onEmpty != nil {
			onEmpty(stepCtx, assertionMode, err)
		}
		return
	}

	ReportAll(stepCtx, assertionMode, expectations, err, value)
}
