package expect

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type ReportFunc[T any] func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, value T, res polling.CheckResult)

func StandardReport[T any](name string) ReportFunc[T] {
	return func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, value T, res polling.CheckResult) {
		a := polling.PickAsserter(stepCtx, mode)
		if !res.Ok {
			a.True(false, "[%s] %s", name, res.Reason)
		} else {
			a.True(true, "[%s]", name)
		}
	}
}

func StandardReportWithActual[T any](name string, getActual func(value T) string) ReportFunc[T] {
	return func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, value T, res polling.CheckResult) {
		a := polling.PickAsserter(stepCtx, mode)
		if !res.Ok {
			a.True(false, "[%s] %s", name, res.Reason)
		} else {
			if getActual != nil {
				actual := getActual(value)
				a.True(true, "[%s] actual: %s", name, actual)
			} else {
				a.True(true, "[%s]", name)
			}
		}
	}
}
