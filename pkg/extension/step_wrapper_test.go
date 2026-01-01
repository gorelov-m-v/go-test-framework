package extension

import (
	"testing"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
)

type mockStepCtx struct {
	provider.StepCtx
}

func TestWithAsyncMode(t *testing.T) {
	t.Run("wraps regular StepCtx with AsyncMode", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		wrappedCtx := WithAsyncMode(mockCtx)

		provider, ok := wrappedCtx.(StepModeProvider)
		assert.True(t, ok, "wrapped context should implement StepModeProvider")

		mode := provider.StepMode()
		assert.Equal(t, AsyncMode, mode, "should return AsyncMode")
	})

	t.Run("updates mode if already wrapped", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		wrappedCtx := WithSyncMode(mockCtx)
		provider1, _ := wrappedCtx.(StepModeProvider)
		assert.Equal(t, SyncMode, provider1.StepMode())

		reWrappedCtx := WithAsyncMode(wrappedCtx)
		provider2, _ := reWrappedCtx.(StepModeProvider)
		assert.Equal(t, AsyncMode, provider2.StepMode(), "should update to AsyncMode")
	})
}

func TestWithSyncMode(t *testing.T) {
	t.Run("wraps regular StepCtx with SyncMode", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		wrappedCtx := WithSyncMode(mockCtx)

		provider, ok := wrappedCtx.(StepModeProvider)
		assert.True(t, ok, "wrapped context should implement StepModeProvider")

		mode := provider.StepMode()
		assert.Equal(t, SyncMode, mode, "should return SyncMode")
	})

	t.Run("updates mode if already wrapped", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		wrappedCtx := WithAsyncMode(mockCtx)
		provider1, _ := wrappedCtx.(StepModeProvider)
		assert.Equal(t, AsyncMode, provider1.StepMode())

		reWrappedCtx := WithSyncMode(wrappedCtx)
		provider2, _ := reWrappedCtx.(StepModeProvider)
		assert.Equal(t, SyncMode, provider2.StepMode(), "should update to SyncMode")
	})
}

func TestGetStepMode_WithWrapper(t *testing.T) {
	t.Run("returns AsyncMode for wrapped async context", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		asyncCtx := WithAsyncMode(mockCtx)

		mode := GetStepMode(asyncCtx)
		assert.Equal(t, AsyncMode, mode)
	})

	t.Run("returns SyncMode for wrapped sync context", func(t *testing.T) {
		mockCtx := &mockStepCtx{}
		syncCtx := WithSyncMode(mockCtx)

		mode := GetStepMode(syncCtx)
		assert.Equal(t, SyncMode, mode)
	})

	t.Run("returns SyncMode for unwrapped context (default)", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		mode := GetStepMode(mockCtx)
		assert.Equal(t, SyncMode, mode, "should default to SyncMode for unwrapped context")
	})
}

func TestStepModeProvider_Integration(t *testing.T) {
	t.Run("wrapped context can be used with GetStepMode", func(t *testing.T) {
		mockCtx := &mockStepCtx{}

		asyncCtx := WithAsyncMode(mockCtx)
		asyncMode := GetStepMode(asyncCtx)
		assert.Equal(t, AsyncMode, asyncMode)

		syncCtx := WithSyncMode(mockCtx)
		syncMode := GetStepMode(syncCtx)
		assert.Equal(t, SyncMode, syncMode)

		defaultMode := GetStepMode(mockCtx)
		assert.Equal(t, SyncMode, defaultMode)
	})
}
