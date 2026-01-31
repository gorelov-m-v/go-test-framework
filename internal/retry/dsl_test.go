package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

type mockStepCtx struct {
	mode polling.StepMode
}

func (m *mockStepCtx) StepMode() polling.StepMode                               { return m.mode }
func (m *mockStepCtx) Step(step *allure.Step)                                   {}
func (m *mockStepCtx) NewStep(stepName string, parameters ...*allure.Parameter) {}
func (m *mockStepCtx) WithNewStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
}
func (m *mockStepCtx) WithNewAsyncStep(stepName string, step func(sCtx provider.StepCtx), params ...*allure.Parameter) {
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
func (m *mockStepCtx) Broken()                                     {}
func (m *mockStepCtx) BrokenNow()                                  {}
func (m *mockStepCtx) Fail()                                       {}
func (m *mockStepCtx) FailNow()                                    {}
func (m *mockStepCtx) Log(args ...interface{})                     {}
func (m *mockStepCtx) Logf(format string, args ...interface{})     {}
func (m *mockStepCtx) Error(args ...interface{})                   {}
func (m *mockStepCtx) Errorf(format string, args ...interface{})   {}
func (m *mockStepCtx) Break(args ...interface{})                   {}
func (m *mockStepCtx) Breakf(format string, args ...interface{})   {}
func (m *mockStepCtx) Name() string                                { return "mock" }

var _ polling.StepModeProvider = (*mockStepCtx)(nil)

func newAsyncConfig() config.AsyncConfig {
	return config.AsyncConfig{
		Enabled:  true,
		Timeout:  100 * time.Millisecond,
		Interval: 10 * time.Millisecond,
	}
}

func TestExecuteDSL_SyncMode_NoExpectations(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.SyncMode}
	executorCalls := 0

	result, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) (string, error) {
			executorCalls++
			return "result", nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, 1, executorCalls)
	assert.Equal(t, 1, summary.Attempts)
	assert.True(t, summary.Success)
}

func TestExecuteDSL_SyncMode_WithExpectations(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.SyncMode}
	executorCalls := 0

	exp := &expect.Expectation[string]{
		Name: "check",
		Check: func(err error, result string) polling.CheckResult {
			return polling.CheckResult{Ok: true}
		},
	}

	result, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:          context.Background(),
		StepCtx:      stepCtx,
		AsyncConfig:  newAsyncConfig(),
		Expectations: []*expect.Expectation[string]{exp},
		Executor: func(ctx context.Context) (string, error) {
			executorCalls++
			return "result", nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, 1, executorCalls)
	assert.Equal(t, 1, summary.Attempts)
}

func TestExecuteDSL_AsyncMode_Disabled(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.AsyncMode}
	executorCalls := 0

	exp := &expect.Expectation[string]{
		Name: "check",
		Check: func(err error, result string) polling.CheckResult {
			return polling.CheckResult{Ok: true}
		},
	}

	cfg := newAsyncConfig()
	cfg.Enabled = false

	result, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:          context.Background(),
		StepCtx:      stepCtx,
		AsyncConfig:  cfg,
		Expectations: []*expect.Expectation[string]{exp},
		Executor: func(ctx context.Context) (string, error) {
			executorCalls++
			return "result", nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, 1, executorCalls)
	assert.Equal(t, 1, summary.Attempts)
}

func TestExecuteDSL_AsyncMode_RetryUntilSuccess(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.AsyncMode}
	executorCalls := 0

	exp := &expect.Expectation[string]{
		Name: "check",
		Check: func(err error, result string) polling.CheckResult {
			if result == "success" {
				return polling.CheckResult{Ok: true}
			}
			return polling.CheckResult{Ok: false, Retryable: true, Reason: "not ready"}
		},
	}

	result, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:          context.Background(),
		StepCtx:      stepCtx,
		AsyncConfig:  newAsyncConfig(),
		Expectations: []*expect.Expectation[string]{exp},
		Executor: func(ctx context.Context) (string, error) {
			executorCalls++
			if executorCalls >= 3 {
				return "success", nil
			}
			return "pending", nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.GreaterOrEqual(t, executorCalls, 3)
	assert.True(t, summary.Success)
}

func TestExecuteDSL_AsyncMode_Timeout(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.AsyncMode}

	exp := &expect.Expectation[string]{
		Name: "check",
		Check: func(err error, result string) polling.CheckResult {
			return polling.CheckResult{Ok: false, Retryable: true, Reason: "never ready"}
		},
	}

	cfg := config.AsyncConfig{
		Enabled:  true,
		Timeout:  50 * time.Millisecond,
		Interval: 10 * time.Millisecond,
	}

	_, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:          context.Background(),
		StepCtx:      stepCtx,
		AsyncConfig:  cfg,
		Expectations: []*expect.Expectation[string]{exp},
		Executor: func(ctx context.Context) (string, error) {
			return "pending", nil
		},
	})

	assert.Error(t, err)
	assert.False(t, summary.Success)
	assert.NotEmpty(t, summary.TimeoutReason)
}

func TestExecuteDSL_NilResultHandling(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.SyncMode}

	type Result struct {
		Value string
	}

	result, err, _ := ExecuteDSL(DSLConfig[*Result, *Result]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) (*Result, error) {
			return nil, nil
		},
		NilResultFactory: func(err error) *Result {
			return &Result{Value: "default"}
		},
	})

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "default", result.Value)
}

func TestExecuteDSL_NilResultWithError(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.SyncMode}

	type Result struct {
		Value string
	}

	expectedErr := errors.New("test error")

	result, err, _ := ExecuteDSL(DSLConfig[*Result, *Result]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) (*Result, error) {
			return nil, expectedErr
		},
		NilResultFactory: func(err error) *Result {
			return &Result{Value: "error:" + err.Error()}
		},
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NotNil(t, result)
	assert.Equal(t, "error:test error", result.Value)
}

func TestExecuteDSL_PostProcess(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.SyncMode}
	postProcessCalled := false

	type Result struct {
		HasError bool
	}

	result, _, summary := ExecuteDSL(DSLConfig[*Result, *Result]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) (*Result, error) {
			return &Result{HasError: true}, nil
		},
		PostProcess: func(result *Result, err error, summary *polling.PollingSummary) {
			postProcessCalled = true
			if result != nil && result.HasError {
				summary.Success = false
				summary.LastError = "internal error"
			}
		},
	})

	assert.True(t, postProcessCalled)
	assert.NotNil(t, result)
	assert.False(t, summary.Success)
	assert.Equal(t, "internal error", summary.LastError)
}

func TestExecuteDSL_CustomChecker(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.AsyncMode}
	checkerCalls := 0

	result, err, summary := ExecuteDSL(DSLConfig[[]byte, []byte]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) ([]byte, error) {
			return []byte("data"), nil
		},
		Checker: func(result []byte, err error) []polling.CheckResult {
			checkerCalls++
			if string(result) == "data" {
				return []polling.CheckResult{{Ok: true}}
			}
			return []polling.CheckResult{{Ok: false, Retryable: true, Reason: "wrong data"}}
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, []byte("data"), result)
	assert.Equal(t, 1, checkerCalls)
	assert.True(t, summary.Success)
}

func TestExecuteDSL_WithConvert(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.AsyncMode}

	type TypedResult struct {
		Value int
	}

	type AnyResult struct {
		Value interface{}
	}

	exp := &expect.Expectation[*AnyResult]{
		Name: "check",
		Check: func(err error, result *AnyResult) polling.CheckResult {
			if result != nil && result.Value == 42 {
				return polling.CheckResult{Ok: true}
			}
			return polling.CheckResult{Ok: false, Retryable: true, Reason: "wrong value"}
		},
	}

	result, err, summary := ExecuteDSL(DSLConfig[*TypedResult, *AnyResult]{
		Ctx:          context.Background(),
		StepCtx:      stepCtx,
		AsyncConfig:  newAsyncConfig(),
		Expectations: []*expect.Expectation[*AnyResult]{exp},
		Executor: func(ctx context.Context) (*TypedResult, error) {
			return &TypedResult{Value: 42}, nil
		},
		Convert: func(result *TypedResult) *AnyResult {
			if result == nil {
				return nil
			}
			return &AnyResult{Value: result.Value}
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 42, result.Value)
	assert.True(t, summary.Success)
}

func TestExecuteDSLSimple(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.SyncMode}

	result, err, summary := ExecuteDSLSimple(DSLConfig[string, string]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) (string, error) {
			return "simple", nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "simple", result)
	assert.True(t, summary.Success)
}

func TestIsNilResult(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil pointer", (*string)(nil), true},
		{"nil slice", []string(nil), true},
		{"nil map", map[string]string(nil), true},
		{"nil interface", (interface{})(nil), true},
		{"empty slice", []string{}, false},
		{"empty map", map[string]string{}, false},
		{"non-nil pointer", new(string), false},
		{"string value", "test", false},
		{"int value", 42, false},
		{"zero int", 0, false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNilResult(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExecuteDSL_ContextCancellation(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.AsyncMode}

	ctx, cancel := context.WithCancel(context.Background())

	exp := &expect.Expectation[string]{
		Name: "check",
		Check: func(err error, result string) polling.CheckResult {
			return polling.CheckResult{Ok: false, Retryable: true, Reason: "never ready"}
		},
	}

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	_, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:          ctx,
		StepCtx:      stepCtx,
		AsyncConfig:  newAsyncConfig(),
		Expectations: []*expect.Expectation[string]{exp},
		Executor: func(ctx context.Context) (string, error) {
			return "pending", nil
		},
	})

	assert.Error(t, err)
	assert.False(t, summary.Success)
}

func TestExecuteDSL_ExecutorError(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.SyncMode}
	expectedErr := errors.New("executor error")

	_, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) (string, error) {
			return "", expectedErr
		},
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.False(t, summary.Success)
	assert.Equal(t, expectedErr.Error(), summary.LastError)
}

func TestExecuteDSL_NoExpectationsNoChecker_SingleExecution(t *testing.T) {
	stepCtx := &mockStepCtx{mode: polling.AsyncMode}
	executorCalls := 0

	result, err, summary := ExecuteDSL(DSLConfig[string, string]{
		Ctx:         context.Background(),
		StepCtx:     stepCtx,
		AsyncConfig: newAsyncConfig(),
		Executor: func(ctx context.Context) (string, error) {
			executorCalls++
			return "result", nil
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "result", result)
	assert.Equal(t, 1, executorCalls)
	assert.True(t, summary.Success)
}

func TestBuildCheckerFromConfig_WithExpectations(t *testing.T) {
	exp := &expect.Expectation[string]{
		Name: "check",
		Check: func(err error, result string) polling.CheckResult {
			if result == "good" {
				return polling.CheckResult{Ok: true}
			}
			return polling.CheckResult{Ok: false, Reason: "bad result"}
		},
	}

	cfg := DSLConfig[string, string]{
		Expectations: []*expect.Expectation[string]{exp},
	}

	checker := buildCheckerFromConfig(cfg)
	results := checker("good", nil)

	assert.Len(t, results, 1)
	assert.True(t, results[0].Ok)
}

func TestBuildCheckerFromConfig_WithCustomChecker(t *testing.T) {
	customChecker := func(result string, err error) []polling.CheckResult {
		return []polling.CheckResult{{Ok: true, Reason: "custom"}}
	}

	cfg := DSLConfig[string, string]{
		Checker: customChecker,
		Expectations: []*expect.Expectation[string]{{
			Name: "ignored",
			Check: func(err error, result string) polling.CheckResult {
				return polling.CheckResult{Ok: false}
			},
		}},
	}

	checker := buildCheckerFromConfig(cfg)
	results := checker("any", nil)

	assert.Len(t, results, 1)
	assert.True(t, results[0].Ok)
	assert.Equal(t, "custom", results[0].Reason)
}
