package dsl

import (
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
)

// mockStepCtx is a minimal mock implementation of provider.StepCtx for testing
type mockStepCtx struct {
	provider.StepCtx
}

func TestWithAsyncMode(t *testing.T) {
	t.Run("wraps regular StepCtx with AsyncMode", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		wrappedCtx := WithAsyncMode(mockCtx)

		// Should implement StepModeProvider
		provider, ok := wrappedCtx.(StepModeProvider)
		assert.True(t, ok, "wrapped context should implement StepModeProvider")

		// Should return AsyncMode
		mode := provider.StepMode()
		assert.Equal(t, AsyncMode, mode, "should return AsyncMode")
	})

	t.Run("updates mode if already wrapped", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		// First wrap with SyncMode
		wrappedCtx := WithSyncMode(mockCtx)
		provider1, _ := wrappedCtx.(StepModeProvider)
		assert.Equal(t, SyncMode, provider1.StepMode())

		// Re-wrap with AsyncMode
		reWrappedCtx := WithAsyncMode(wrappedCtx)
		provider2, _ := reWrappedCtx.(StepModeProvider)
		assert.Equal(t, AsyncMode, provider2.StepMode(), "should update to AsyncMode")
	})
}

func TestWithSyncMode(t *testing.T) {
	t.Run("wraps regular StepCtx with SyncMode", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		wrappedCtx := WithSyncMode(mockCtx)

		// Should implement StepModeProvider
		provider, ok := wrappedCtx.(StepModeProvider)
		assert.True(t, ok, "wrapped context should implement StepModeProvider")

		// Should return SyncMode
		mode := provider.StepMode()
		assert.Equal(t, SyncMode, mode, "should return SyncMode")
	})

	t.Run("updates mode if already wrapped", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		// First wrap with AsyncMode
		wrappedCtx := WithAsyncMode(mockCtx)
		provider1, _ := wrappedCtx.(StepModeProvider)
		assert.Equal(t, AsyncMode, provider1.StepMode())

		// Re-wrap with SyncMode
		reWrappedCtx := WithSyncMode(wrappedCtx)
		provider2, _ := reWrappedCtx.(StepModeProvider)
		assert.Equal(t, SyncMode, provider2.StepMode(), "should update to SyncMode")
	})
}

func TestGetStepMode_WithWrapper(t *testing.T) {
	t.Run("returns AsyncMode for wrapped async context", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		asyncCtx := WithAsyncMode(mockCtx)

		mode := getStepMode(asyncCtx)
		assert.Equal(t, AsyncMode, mode)
	})

	t.Run("returns SyncMode for wrapped sync context", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		syncCtx := WithSyncMode(mockCtx)

		mode := getStepMode(syncCtx)
		assert.Equal(t, SyncMode, mode)
	})

	t.Run("returns SyncMode for unwrapped context (default)", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		mode := getStepMode(mockCtx)
		assert.Equal(t, SyncMode, mode, "should default to SyncMode for unwrapped context")
	})
}

func TestStepModeProvider_Integration(t *testing.T) {
	t.Run("wrapped context can be used with getStepMode", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		// Test async mode
		asyncCtx := WithAsyncMode(mockCtx)
		asyncMode := getStepMode(asyncCtx)
		assert.Equal(t, AsyncMode, asyncMode)

		// Test sync mode
		syncCtx := WithSyncMode(mockCtx)
		syncMode := getStepMode(syncCtx)
		assert.Equal(t, SyncMode, syncMode)

		// Test default (unwrapped)
		defaultMode := getStepMode(mockCtx)
		assert.Equal(t, SyncMode, defaultMode)
	})
}
