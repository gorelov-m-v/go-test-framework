package polling

import (
	"testing"
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

func TestCheckResult(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cr := CheckResult{}
		if cr.Ok {
			t.Error("Expected Ok to be false by default")
		}
		if cr.Retryable {
			t.Error("Expected Retryable to be false by default")
		}
		if cr.Reason != "" {
			t.Error("Expected Reason to be empty by default")
		}
	})

	t.Run("with values", func(t *testing.T) {
		cr := CheckResult{
			Ok:        true,
			Retryable: true,
			Reason:    "test reason",
		}
		if !cr.Ok {
			t.Error("Expected Ok to be true")
		}
		if !cr.Retryable {
			t.Error("Expected Retryable to be true")
		}
		if cr.Reason != "test reason" {
			t.Errorf("Expected Reason to be 'test reason', got '%s'", cr.Reason)
		}
	})
}

func TestPollingSummary(t *testing.T) {
	t.Run("with all fields", func(t *testing.T) {
		ps := PollingSummary{
			Attempts:      5,
			ElapsedTime:   "10s",
			Success:       false,
			LastError:     "connection timeout",
			FailedChecks:  []string{"check1 failed", "check2 failed"},
			TimeoutReason: "deadline exceeded",
		}
		if ps.Attempts != 5 {
			t.Errorf("Expected Attempts=5, got %d", ps.Attempts)
		}
		if ps.ElapsedTime != "10s" {
			t.Errorf("Expected ElapsedTime='10s', got '%s'", ps.ElapsedTime)
		}
		if ps.Success {
			t.Error("Expected Success to be false")
		}
		if len(ps.FailedChecks) != 2 {
			t.Errorf("Expected 2 failed checks, got %d", len(ps.FailedChecks))
		}
	})
}

func TestRetryContext_CalculateNextDelay(t *testing.T) {
	t.Run("no backoff no jitter", func(t *testing.T) {
		rc := &RetryContext{
			Attempt:      1,
			CurrentDelay: 100 * time.Millisecond,
			Cfg: config.AsyncConfig{
				Interval: 100 * time.Millisecond,
				Backoff: config.BackoffConfig{
					Enabled: false,
				},
				Jitter: 0,
			},
		}
		delay := rc.CalculateNextDelay()
		if delay != 100*time.Millisecond {
			t.Errorf("Expected delay=100ms, got %v", delay)
		}
	})

	t.Run("backoff enabled first attempt", func(t *testing.T) {
		rc := &RetryContext{
			Attempt:      1,
			CurrentDelay: 100 * time.Millisecond,
			Cfg: config.AsyncConfig{
				Interval: 100 * time.Millisecond,
				Backoff: config.BackoffConfig{
					Enabled:     true,
					Factor:      2.0,
					MaxInterval: 1 * time.Second,
				},
				Jitter: 0,
			},
		}
		delay := rc.CalculateNextDelay()
		if delay != 100*time.Millisecond {
			t.Errorf("Expected delay=100ms on first attempt, got %v", delay)
		}
	})

	t.Run("backoff enabled second attempt", func(t *testing.T) {
		rc := &RetryContext{
			Attempt:      2,
			CurrentDelay: 100 * time.Millisecond,
			Cfg: config.AsyncConfig{
				Interval: 100 * time.Millisecond,
				Backoff: config.BackoffConfig{
					Enabled:     true,
					Factor:      2.0,
					MaxInterval: 1 * time.Second,
				},
				Jitter: 0,
			},
		}
		delay := rc.CalculateNextDelay()
		if delay != 200*time.Millisecond {
			t.Errorf("Expected delay=200ms on second attempt with factor 2, got %v", delay)
		}
		if rc.CurrentDelay != 200*time.Millisecond {
			t.Errorf("Expected CurrentDelay updated to 200ms, got %v", rc.CurrentDelay)
		}
	})

	t.Run("backoff respects max interval", func(t *testing.T) {
		rc := &RetryContext{
			Attempt:      5,
			CurrentDelay: 800 * time.Millisecond,
			Cfg: config.AsyncConfig{
				Interval: 100 * time.Millisecond,
				Backoff: config.BackoffConfig{
					Enabled:     true,
					Factor:      2.0,
					MaxInterval: 1 * time.Second,
				},
				Jitter: 0,
			},
		}
		delay := rc.CalculateNextDelay()
		if delay != 1*time.Second {
			t.Errorf("Expected delay capped at 1s, got %v", delay)
		}
		if rc.CurrentDelay != 1*time.Second {
			t.Errorf("Expected CurrentDelay capped at 1s, got %v", rc.CurrentDelay)
		}
	})

	t.Run("jitter adds variance", func(t *testing.T) {
		rc := &RetryContext{
			Attempt:      1,
			CurrentDelay: 100 * time.Millisecond,
			Cfg: config.AsyncConfig{
				Interval: 100 * time.Millisecond,
				Backoff: config.BackoffConfig{
					Enabled: false,
				},
				Jitter: 0.5,
			},
		}

		delays := make(map[time.Duration]bool)
		for i := 0; i < 100; i++ {
			rc.CurrentDelay = 100 * time.Millisecond
			delay := rc.CalculateNextDelay()
			delays[delay] = true
			if delay < 50*time.Millisecond || delay > 150*time.Millisecond {
				t.Errorf("Delay %v outside expected jitter range [50ms, 150ms]", delay)
			}
		}
		if len(delays) < 5 {
			t.Error("Expected jitter to produce varied delays")
		}
	})

	t.Run("jitter with backoff", func(t *testing.T) {
		rc := &RetryContext{
			Attempt:      2,
			CurrentDelay: 100 * time.Millisecond,
			Cfg: config.AsyncConfig{
				Interval: 100 * time.Millisecond,
				Backoff: config.BackoffConfig{
					Enabled:     true,
					Factor:      2.0,
					MaxInterval: 1 * time.Second,
				},
				Jitter: 0.2,
			},
		}
		delay := rc.CalculateNextDelay()
		if delay < 160*time.Millisecond || delay > 240*time.Millisecond {
			t.Errorf("Delay %v outside expected range [160ms, 240ms] (200ms ± 20%%)", delay)
		}
	})

	t.Run("negative delay after jitter resets to interval", func(t *testing.T) {
		rc := &RetryContext{
			Attempt:      1,
			CurrentDelay: 1 * time.Millisecond,
			Cfg: config.AsyncConfig{
				Interval: 100 * time.Millisecond,
				Backoff: config.BackoffConfig{
					Enabled: false,
				},
				Jitter: 10.0,
			},
		}

		for i := 0; i < 100; i++ {
			rc.CurrentDelay = 1 * time.Millisecond
			delay := rc.CalculateNextDelay()
			if delay < 0 {
				t.Error("Delay should never be negative")
			}
		}
	})
}

func TestGetAssertionModeFromStepMode(t *testing.T) {
	tests := []struct {
		name     string
		stepMode StepMode
		want     AssertionMode
	}{
		{"SyncMode returns AssertionRequire", SyncMode, AssertionRequire},
		{"AsyncMode returns AssertionAssert", AsyncMode, AssertionAssert},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetAssertionModeFromStepMode(tt.stepMode)
			if got != tt.want {
				t.Errorf("GetAssertionModeFromStepMode(%v) = %v, want %v", tt.stepMode, got, tt.want)
			}
		})
	}
}

func TestAssertionModeConstants(t *testing.T) {
	if AssertionRequire != 0 {
		t.Errorf("Expected AssertionRequire=0, got %d", AssertionRequire)
	}
	if AssertionAssert != 1 {
		t.Errorf("Expected AssertionAssert=1, got %d", AssertionAssert)
	}
}

func TestStepModeConstants(t *testing.T) {
	if SyncMode != 0 {
		t.Errorf("Expected SyncMode=0, got %d", SyncMode)
	}
	if AsyncMode != 1 {
		t.Errorf("Expected AsyncMode=1, got %d", AsyncMode)
	}
}

func TestFinalFailureMessage(t *testing.T) {
	t.Run("no failed checks", func(t *testing.T) {
		summary := PollingSummary{
			Attempts:     3,
			ElapsedTime:  "5s",
			FailedChecks: []string{},
		}
		msg := FinalFailureMessage(summary)
		expected := "Expectations not met within timeout"
		if msg != expected {
			t.Errorf("Expected '%s', got '%s'", expected, msg)
		}
	})

	t.Run("with failed checks", func(t *testing.T) {
		summary := PollingSummary{
			Attempts:     3,
			ElapsedTime:  "5s",
			FailedChecks: []string{"check1 failed", "check2 failed"},
		}
		msg := FinalFailureMessage(summary)
		if !containsSubstring(msg, "3 attempts") {
			t.Errorf("Message should contain attempts count: %s", msg)
		}
		if !containsSubstring(msg, "5s") {
			t.Errorf("Message should contain elapsed time: %s", msg)
		}
		if !containsSubstring(msg, "check1 failed") {
			t.Errorf("Message should contain first check: %s", msg)
		}
		if !containsSubstring(msg, "check2 failed") {
			t.Errorf("Message should contain second check: %s", msg)
		}
		if !containsSubstring(msg, "[1]") || !containsSubstring(msg, "[2]") {
			t.Errorf("Message should contain numbered checks: %s", msg)
		}
	})

	t.Run("with last error", func(t *testing.T) {
		summary := PollingSummary{
			Attempts:     3,
			ElapsedTime:  "5s",
			FailedChecks: []string{"check failed"},
			LastError:    "connection refused",
		}
		msg := FinalFailureMessage(summary)
		if !containsSubstring(msg, "connection refused") {
			t.Errorf("Message should contain last error: %s", msg)
		}
		if !containsSubstring(msg, "Last error") {
			t.Errorf("Message should have 'Last error' prefix: %s", msg)
		}
	})

	t.Run("nil failed checks same as empty", func(t *testing.T) {
		summary := PollingSummary{
			Attempts:     1,
			ElapsedTime:  "1s",
			FailedChecks: nil,
		}
		msg := FinalFailureMessage(summary)
		expected := "Expectations not met within timeout"
		if msg != expected {
			t.Errorf("Expected '%s', got '%s'", expected, msg)
		}
	})
}

func TestRetryContextFields(t *testing.T) {
	now := time.Now()
	rc := RetryContext{
		Attempt:       5,
		Cfg:           config.AsyncConfig{Timeout: 10 * time.Second},
		Deadline:      now.Add(10 * time.Second),
		CurrentDelay:  200 * time.Millisecond,
		LastErr:       nil,
		LastResult:    "some result",
		FailedReasons: []string{"reason1", "reason2"},
	}

	if rc.Attempt != 5 {
		t.Errorf("Expected Attempt=5, got %d", rc.Attempt)
	}
	if rc.Cfg.Timeout != 10*time.Second {
		t.Errorf("Expected Cfg.Timeout=10s, got %v", rc.Cfg.Timeout)
	}
	if rc.CurrentDelay != 200*time.Millisecond {
		t.Errorf("Expected CurrentDelay=200ms, got %v", rc.CurrentDelay)
	}
	if rc.LastResult != "some result" {
		t.Errorf("Expected LastResult='some result', got %v", rc.LastResult)
	}
	if len(rc.FailedReasons) != 2 {
		t.Errorf("Expected 2 FailedReasons, got %d", len(rc.FailedReasons))
	}
}

func TestCalculateNextDelay_MultipleAttempts(t *testing.T) {
	rc := &RetryContext{
		Attempt:      1,
		CurrentDelay: 100 * time.Millisecond,
		Cfg: config.AsyncConfig{
			Interval: 100 * time.Millisecond,
			Backoff: config.BackoffConfig{
				Enabled:     true,
				Factor:      1.5,
				MaxInterval: 500 * time.Millisecond,
			},
			Jitter: 0,
		},
	}

	expectedDelays := []time.Duration{
		100 * time.Millisecond,
		150 * time.Millisecond,
		225 * time.Millisecond,
		337500 * time.Microsecond,
		500 * time.Millisecond,
		500 * time.Millisecond,
	}

	for i, expected := range expectedDelays {
		rc.Attempt = i + 1
		delay := rc.CalculateNextDelay()

		tolerance := time.Millisecond
		if diff := delay - expected; diff > tolerance || diff < -tolerance {
			t.Errorf("Attempt %d: expected delay ~%v, got %v", i+1, expected, delay)
		}
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
