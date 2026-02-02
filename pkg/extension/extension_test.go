package extension

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type mockStepCtx struct {
	steps           []string
	asyncSteps      []string
	brokenCalled    bool
	brokenNowCalled bool
	breakMessage    string
}

func (m *mockStepCtx) Step(step *allure.Step)                                   {}
func (m *mockStepCtx) NewStep(stepName string, parameters ...*allure.Parameter) {}
func (m *mockStepCtx) WithNewStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	m.steps = append(m.steps, stepName)
	step(m)
}
func (m *mockStepCtx) WithNewAsyncStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
	m.asyncSteps = append(m.asyncSteps, stepName)
	go step(m)
}
func (m *mockStepCtx) WithParameters(parameters ...*allure.Parameter)   {}
func (m *mockStepCtx) WithNewParameters(kv ...interface{})              {}
func (m *mockStepCtx) WithAttachments(attachment ...*allure.Attachment) {}
func (m *mockStepCtx) WithNewAttachment(name string, mimeType allure.MimeType, content []byte) {
}
func (m *mockStepCtx) Assert() provider.Asserts                    { return nil }
func (m *mockStepCtx) Require() provider.Asserts                   { return nil }
func (m *mockStepCtx) LogStep(args ...interface{})                 {}
func (m *mockStepCtx) LogfStep(format string, args ...interface{}) {}
func (m *mockStepCtx) WithStatusDetails(message, trace string)     {}
func (m *mockStepCtx) CurrentStep() *allure.Step                   { return nil }
func (m *mockStepCtx) Broken()                                     { m.brokenCalled = true }
func (m *mockStepCtx) BrokenNow()                                  { m.brokenNowCalled = true }
func (m *mockStepCtx) Fail()                                       {}
func (m *mockStepCtx) FailNow()                                    {}
func (m *mockStepCtx) Log(args ...interface{})                     {}
func (m *mockStepCtx) Logf(format string, args ...interface{})     {}
func (m *mockStepCtx) Error(args ...interface{})                   {}
func (m *mockStepCtx) Errorf(format string, args ...interface{})   {}
func (m *mockStepCtx) Break(args ...interface{}) {
	m.brokenCalled = true
	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			m.breakMessage = s
		}
	}
}
func (m *mockStepCtx) Breakf(format string, args ...interface{}) { m.brokenCalled = true }
func (m *mockStepCtx) Name() string                              { return "mock" }

// =============================================================================
// stepCtxWrapper tests
// =============================================================================

func TestStepCtxWrapper_StepMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     StepMode
		expected StepMode
	}{
		{
			name:     "returns SyncMode",
			mode:     SyncMode,
			expected: SyncMode,
		},
		{
			name:     "returns AsyncMode",
			mode:     AsyncMode,
			expected: AsyncMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := &stepCtxWrapper{
				StepCtx: &mockStepCtx{},
				mode:    tt.mode,
			}
			assert.Equal(t, tt.expected, wrapper.StepMode())
		})
	}
}

func TestStepCtxWrapper_WithNewStep_PreservesMode(t *testing.T) {
	tests := []struct {
		name string
		mode StepMode
	}{
		{name: "preserves SyncMode", mode: SyncMode},
		{name: "preserves AsyncMode", mode: AsyncMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			innerCtx := &mockStepCtx{}
			wrapper := &stepCtxWrapper{
				StepCtx: innerCtx,
				mode:    tt.mode,
			}

			var capturedMode StepMode
			wrapper.WithNewStep("nested step", func(sCtx provider.StepCtx) {
				if modeProvider, ok := sCtx.(polling.StepModeProvider); ok {
					capturedMode = modeProvider.StepMode()
				}
			})

			assert.Equal(t, tt.mode, capturedMode)
			assert.Contains(t, innerCtx.steps, "nested step")
		})
	}
}

func TestStepCtxWrapper_WithNewAsyncStep_PreservesMode(t *testing.T) {
	tests := []struct {
		name string
		mode StepMode
	}{
		{name: "preserves SyncMode in async", mode: SyncMode},
		{name: "preserves AsyncMode in async", mode: AsyncMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			innerCtx := &mockStepCtx{}
			wrapper := &stepCtxWrapper{
				StepCtx: innerCtx,
				mode:    tt.mode,
			}

			var capturedMode StepMode
			var wg sync.WaitGroup
			wg.Add(1)

			wrapper.WithNewAsyncStep("nested async step", func(sCtx provider.StepCtx) {
				defer wg.Done()
				if modeProvider, ok := sCtx.(polling.StepModeProvider); ok {
					capturedMode = modeProvider.StepMode()
				}
			})

			wg.Wait()
			assert.Equal(t, tt.mode, capturedMode)
			assert.Contains(t, innerCtx.asyncSteps, "nested async step")
		})
	}
}

func TestStepCtxWrapper_WithNewStep_WithParameters(t *testing.T) {
	innerCtx := &mockStepCtx{}
	wrapper := &stepCtxWrapper{
		StepCtx: innerCtx,
		mode:    SyncMode,
	}

	param := allure.NewParameter("key", "value")
	stepExecuted := false

	wrapper.WithNewStep("step with params", func(sCtx provider.StepCtx) {
		stepExecuted = true
	}, param)

	assert.True(t, stepExecuted)
	assert.Contains(t, innerCtx.steps, "step with params")
}

func TestStepCtxWrapper_DelegatesOtherMethods(t *testing.T) {
	innerCtx := &mockStepCtx{}
	wrapper := &stepCtxWrapper{
		StepCtx: innerCtx,
		mode:    SyncMode,
	}

	assert.Equal(t, "mock", wrapper.Name())
	assert.Nil(t, wrapper.CurrentStep())
	assert.Nil(t, wrapper.Assert())
	assert.Nil(t, wrapper.Require())
}

// =============================================================================
// WithAsyncMode / WithSyncMode tests
// =============================================================================

func TestWithAsyncMode_FromPlainCtx(t *testing.T) {
	plainCtx := &mockStepCtx{}

	result := WithAsyncMode(plainCtx)

	wrapped, ok := result.(*stepCtxWrapper)
	require.True(t, ok, "should return stepCtxWrapper")
	assert.Equal(t, AsyncMode, wrapped.mode)
	assert.Equal(t, plainCtx, wrapped.StepCtx)
}

func TestWithAsyncMode_FromWrappedCtx(t *testing.T) {
	innerCtx := &mockStepCtx{}
	existingWrapper := &stepCtxWrapper{
		StepCtx: innerCtx,
		mode:    SyncMode,
	}

	result := WithAsyncMode(existingWrapper)

	wrapped, ok := result.(*stepCtxWrapper)
	require.True(t, ok, "should return stepCtxWrapper")
	assert.Equal(t, AsyncMode, wrapped.mode)
	assert.Equal(t, innerCtx, wrapped.StepCtx, "should unwrap nested wrappers")
}

func TestWithSyncMode_FromPlainCtx(t *testing.T) {
	plainCtx := &mockStepCtx{}

	result := WithSyncMode(plainCtx)

	wrapped, ok := result.(*stepCtxWrapper)
	require.True(t, ok, "should return stepCtxWrapper")
	assert.Equal(t, SyncMode, wrapped.mode)
	assert.Equal(t, plainCtx, wrapped.StepCtx)
}

func TestWithSyncMode_FromWrappedCtx(t *testing.T) {
	innerCtx := &mockStepCtx{}
	existingWrapper := &stepCtxWrapper{
		StepCtx: innerCtx,
		mode:    AsyncMode,
	}

	result := WithSyncMode(existingWrapper)

	wrapped, ok := result.(*stepCtxWrapper)
	require.True(t, ok, "should return stepCtxWrapper")
	assert.Equal(t, SyncMode, wrapped.mode)
	assert.Equal(t, innerCtx, wrapped.StepCtx, "should unwrap nested wrappers")
}

func TestWithAsyncMode_PreservesUnderlyingCtx(t *testing.T) {
	innerCtx := &mockStepCtx{}
	wrapped := WithAsyncMode(innerCtx)

	wrapped.(*stepCtxWrapper).StepCtx.(*mockStepCtx).steps = append(
		wrapped.(*stepCtxWrapper).StepCtx.(*mockStepCtx).steps,
		"test",
	)

	assert.Contains(t, innerCtx.steps, "test")
}

func TestWithSyncMode_PreservesUnderlyingCtx(t *testing.T) {
	innerCtx := &mockStepCtx{}
	wrapped := WithSyncMode(innerCtx)

	wrapped.(*stepCtxWrapper).StepCtx.(*mockStepCtx).steps = append(
		wrapped.(*stepCtxWrapper).StepCtx.(*mockStepCtx).steps,
		"test",
	)

	assert.Contains(t, innerCtx.steps, "test")
}

// =============================================================================
// GetStepMode tests
// =============================================================================

func TestGetStepMode_WithStepModeProvider(t *testing.T) {
	wrapper := &stepCtxWrapper{
		StepCtx: &mockStepCtx{},
		mode:    AsyncMode,
	}

	mode := GetStepMode(wrapper)

	assert.Equal(t, AsyncMode, mode)
}

func TestGetStepMode_WithoutStepModeProvider(t *testing.T) {
	plainCtx := &mockStepCtx{}

	mode := GetStepMode(plainCtx)

	assert.Equal(t, SyncMode, mode, "should default to SyncMode")
}

func TestGetStepMode_SyncModeProvider(t *testing.T) {
	wrapper := &stepCtxWrapper{
		StepCtx: &mockStepCtx{},
		mode:    SyncMode,
	}

	mode := GetStepMode(wrapper)

	assert.Equal(t, SyncMode, mode)
}

// =============================================================================
// Constants tests
// =============================================================================

func TestStepModeConstants(t *testing.T) {
	assert.Equal(t, polling.SyncMode, SyncMode)
	assert.Equal(t, polling.AsyncMode, AsyncMode)
}

func TestStepModeTypeAlias(t *testing.T) {
	var mode StepMode = SyncMode
	var pollingMode polling.StepMode = mode

	assert.Equal(t, polling.SyncMode, pollingMode)
}

// =============================================================================
// BaseSuite tests (using internal testing approach)
// =============================================================================

func TestBaseSuite_AsyncWaitGroup_Behavior(t *testing.T) {
	s := &BaseSuite{}

	var order []int
	var mu sync.Mutex

	s.asyncWg.Add(1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		s.asyncWg.Done()
	}()

	s.asyncWg.Wait()
	mu.Lock()
	order = append(order, 2)
	mu.Unlock()

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, order, 2)
	assert.Equal(t, 1, order[0], "async should complete first")
	assert.Equal(t, 2, order[1], "wait should complete second")
}

func TestBaseSuite_MultipleAsyncWaitGroup(t *testing.T) {
	s := &BaseSuite{}

	var startTimes []time.Time
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		s.asyncWg.Add(1)
		go func() {
			mu.Lock()
			startTimes = append(startTimes, time.Now())
			mu.Unlock()
			time.Sleep(50 * time.Millisecond)
			s.asyncWg.Done()
		}()
	}

	s.asyncWg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, startTimes, 3)

	for i := 1; i < len(startTimes); i++ {
		diff := startTimes[i].Sub(startTimes[0])
		assert.Less(t, diff, 30*time.Millisecond, "async tasks should start nearly simultaneously")
	}
}

func TestBaseSuite_CleanupField(t *testing.T) {
	s := &BaseSuite{}

	assert.Nil(t, s.cleanup)

	cleanupCalled := false
	s.cleanup = func(t provider.T) {
		cleanupCalled = true
	}

	require.NotNil(t, s.cleanup)
	s.cleanup(nil)
	assert.True(t, cleanupCalled)
}

func TestBaseSuite_TExtField(t *testing.T) {
	s := &BaseSuite{}

	assert.Nil(t, s.tExt)

	s.tExt = &TExtension{}
	assert.NotNil(t, s.tExt)

	s.tExt = nil
	assert.Nil(t, s.tExt)
}

func TestBaseSuite_AsyncStepCompletesBeforeAfterEach(t *testing.T) {
	s := &BaseSuite{}

	var asyncCompleted int32

	s.asyncWg.Add(1)
	go func() {
		time.Sleep(30 * time.Millisecond)
		atomic.StoreInt32(&asyncCompleted, 1)
		s.asyncWg.Done()
	}()

	s.asyncWg.Wait()

	assert.Equal(t, int32(1), atomic.LoadInt32(&asyncCompleted))
}

// =============================================================================
// Nested step mode propagation tests
// =============================================================================

func TestNestedSteps_ModePropagation(t *testing.T) {
	innerCtx := &mockStepCtx{}
	wrapper := &stepCtxWrapper{
		StepCtx: innerCtx,
		mode:    AsyncMode,
	}

	var level1Mode, level2Mode StepMode

	wrapper.WithNewStep("level1", func(sCtx provider.StepCtx) {
		if mp, ok := sCtx.(polling.StepModeProvider); ok {
			level1Mode = mp.StepMode()
		}

		sCtx.WithNewStep("level2", func(sCtx2 provider.StepCtx) {
			if mp, ok := sCtx2.(polling.StepModeProvider); ok {
				level2Mode = mp.StepMode()
			}
		})
	})

	assert.Equal(t, AsyncMode, level1Mode, "level1 should have AsyncMode")
	assert.Equal(t, AsyncMode, level2Mode, "level2 should inherit AsyncMode")
}

func TestNestedAsyncSteps_ModePropagation(t *testing.T) {
	innerCtx := &mockStepCtx{}
	wrapper := &stepCtxWrapper{
		StepCtx: innerCtx,
		mode:    SyncMode,
	}

	var wg sync.WaitGroup
	var level1Mode, level2Mode StepMode
	var mu sync.Mutex

	wg.Add(1)
	wrapper.WithNewAsyncStep("level1", func(sCtx provider.StepCtx) {
		defer wg.Done()
		if mp, ok := sCtx.(polling.StepModeProvider); ok {
			mu.Lock()
			level1Mode = mp.StepMode()
			mu.Unlock()
		}

		var innerWg sync.WaitGroup
		innerWg.Add(1)
		sCtx.WithNewAsyncStep("level2", func(sCtx2 provider.StepCtx) {
			defer innerWg.Done()
			if mp, ok := sCtx2.(polling.StepModeProvider); ok {
				mu.Lock()
				level2Mode = mp.StepMode()
				mu.Unlock()
			}
		})
		innerWg.Wait()
	})

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, SyncMode, level1Mode, "level1 should have SyncMode")
	assert.Equal(t, SyncMode, level2Mode, "level2 should inherit SyncMode")
}

// =============================================================================
// Edge cases
// =============================================================================

func TestStepCtxWrapper_EmptyStepName(t *testing.T) {
	innerCtx := &mockStepCtx{}
	wrapper := &stepCtxWrapper{
		StepCtx: innerCtx,
		mode:    SyncMode,
	}

	stepExecuted := false
	wrapper.WithNewStep("", func(sCtx provider.StepCtx) {
		stepExecuted = true
	})

	assert.True(t, stepExecuted)
	assert.Contains(t, innerCtx.steps, "")
}

func TestWithAsyncMode_DoubleWrap(t *testing.T) {
	innerCtx := &mockStepCtx{}

	wrapped1 := WithAsyncMode(innerCtx)
	wrapped2 := WithAsyncMode(wrapped1)

	w2, ok := wrapped2.(*stepCtxWrapper)
	require.True(t, ok)
	assert.Equal(t, AsyncMode, w2.mode)
	assert.Equal(t, innerCtx, w2.StepCtx, "should not double-wrap")
}

func TestWithSyncMode_DoubleWrap(t *testing.T) {
	innerCtx := &mockStepCtx{}

	wrapped1 := WithSyncMode(innerCtx)
	wrapped2 := WithSyncMode(wrapped1)

	w2, ok := wrapped2.(*stepCtxWrapper)
	require.True(t, ok)
	assert.Equal(t, SyncMode, w2.mode)
	assert.Equal(t, innerCtx, w2.StepCtx, "should not double-wrap")
}

func TestModeSwitch_AsyncToSync(t *testing.T) {
	innerCtx := &mockStepCtx{}

	asyncWrapped := WithAsyncMode(innerCtx)
	syncWrapped := WithSyncMode(asyncWrapped)

	w, ok := syncWrapped.(*stepCtxWrapper)
	require.True(t, ok)
	assert.Equal(t, SyncMode, w.mode)
	assert.Equal(t, innerCtx, w.StepCtx)
}

func TestModeSwitch_SyncToAsync(t *testing.T) {
	innerCtx := &mockStepCtx{}

	syncWrapped := WithSyncMode(innerCtx)
	asyncWrapped := WithAsyncMode(syncWrapped)

	w, ok := asyncWrapped.(*stepCtxWrapper)
	require.True(t, ok)
	assert.Equal(t, AsyncMode, w.mode)
	assert.Equal(t, innerCtx, w.StepCtx)
}

// =============================================================================
// Note: provider.T interface contains a private method, making it impossible
// to create a full mock outside of allure-go package. Tests for TExtension,
// BaseSuite.T(), Step(), AsyncStep(), BeforeEach(), AfterEach(), and Cleanup()
// methods require integration testing with allure-go framework.
//
// The tests below focus on what can be unit tested without provider.T mock.
// =============================================================================

// =============================================================================
// BaseSuite internal behavior tests (without provider.T dependency)
// =============================================================================

func TestBaseSuite_CleanupField_SetAndCall(t *testing.T) {
	s := &BaseSuite{}

	assert.Nil(t, s.cleanup, "cleanup should be nil initially")

	called := false
	s.cleanup = func(t provider.T) {
		called = true
	}

	assert.NotNil(t, s.cleanup)
	s.cleanup(nil)
	assert.True(t, called, "cleanup function should be called")
}

func TestBaseSuite_TExtField_SetAndReset(t *testing.T) {
	s := &BaseSuite{}

	assert.Nil(t, s.tExt, "tExt should be nil initially")

	s.tExt = &TExtension{}
	assert.NotNil(t, s.tExt)

	s.tExt = nil
	assert.Nil(t, s.tExt, "tExt should be nil after reset")
}

func TestBaseSuite_CurrentTField(t *testing.T) {
	s := &BaseSuite{}

	assert.Nil(t, s.currentT, "currentT should be nil initially")
}
