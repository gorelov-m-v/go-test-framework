package dsl

import (
	"runtime"
	"strings"

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

func isInAsyncGoroutine() bool {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	hasAsyncStepMarker := strings.Contains(stack, "WithNewAsyncStep") ||
		strings.Contains(stack, "allure-go/pkg/framework/core/common.(*CommonT).newAsyncStep")

	notInMainRunner := !strings.Contains(stack, "testing.tRunner")

	return hasAsyncStepMarker || (notInMainRunner && strings.Contains(stack, "goroutine"))
}

func getStepMode(sCtx provider.StepCtx) StepMode {
	if smp, ok := sCtx.(StepModeProvider); ok {
		return smp.StepMode()
	}

	if isInAsyncGoroutine() {
		return AsyncMode
	}

	return SyncMode
}
