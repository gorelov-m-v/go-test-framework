package dsl

import (
	"testing"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
)

// mockT is a minimal mock implementation of provider.T for testing
type mockT struct {
	provider.T
	stepCalls      []string
	asyncStepCalls []string
	lastStepCtx    provider.StepCtx
}

func (m *mockT) WithNewStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	m.stepCalls = append(m.stepCalls, stepName)
	// Create a minimal context
	mockCtx := &mockStepCtx{}
	m.lastStepCtx = mockCtx
	step(mockCtx)
}

func (m *mockT) WithNewAsyncStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	m.asyncStepCalls = append(m.asyncStepCalls, stepName)
	// Create a minimal context
	mockCtx := &mockStepCtx{}
	m.lastStepCtx = mockCtx
	step(mockCtx)
}

func TestTExtension_WithNewStep(t *testing.T) {
	t.Run("wraps step context with SyncMode", func(t *testing.T) {
		mockProvider := &mockT{}
		tExt := NewTExtension(mockProvider)

		var capturedCtx provider.StepCtx
		tExt.WithNewStep("test step", func(sCtx provider.StepCtx) {
			capturedCtx = sCtx
		})

		assert.Len(t, mockProvider.stepCalls, 1)
		assert.Equal(t, "test step", mockProvider.stepCalls[0])

		// Check that context was wrapped with SyncMode
		mode := getStepMode(capturedCtx)
		assert.Equal(t, SyncMode, mode, "WithNewStep should wrap context with SyncMode")
	})
}

func TestTExtension_WithNewAsyncStep(t *testing.T) {
	t.Run("wraps step context with AsyncMode", func(t *testing.T) {
		mockProvider := &mockT{}
		tExt := NewTExtension(mockProvider)

		var capturedCtx provider.StepCtx
		tExt.WithNewAsyncStep("test async step", func(sCtx provider.StepCtx) {
			capturedCtx = sCtx
		})

		assert.Len(t, mockProvider.asyncStepCalls, 1)
		assert.Equal(t, "test async step", mockProvider.asyncStepCalls[0])

		// Check that context was wrapped with AsyncMode
		mode := getStepMode(capturedCtx)
		assert.Equal(t, AsyncMode, mode, "WithNewAsyncStep should wrap context with AsyncMode")
	})
}

func TestTExtension_Integration(t *testing.T) {
	t.Run("different modes for sync and async steps", func(t *testing.T) {
		mockProvider := &mockT{}
		tExt := NewTExtension(mockProvider)

		var syncCtx, asyncCtx provider.StepCtx

		tExt.WithNewStep("sync step", func(sCtx provider.StepCtx) {
			syncCtx = sCtx
		})

		tExt.WithNewAsyncStep("async step", func(sCtx provider.StepCtx) {
			asyncCtx = sCtx
		})

		syncMode := getStepMode(syncCtx)
		asyncMode := getStepMode(asyncCtx)

		assert.Equal(t, SyncMode, syncMode, "sync step should have SyncMode")
		assert.Equal(t, AsyncMode, asyncMode, "async step should have AsyncMode")
	})
}
