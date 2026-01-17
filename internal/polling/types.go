package polling

import (
	"math/rand"
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

type CheckResult struct {
	Ok        bool
	Retryable bool
	Reason    string
}

type PollingSummary struct {
	Attempts      int      `json:"attempts"`
	ElapsedTime   string   `json:"elapsed_time"`
	Success       bool     `json:"success"`
	LastError     string   `json:"last_error,omitempty"`
	FailedChecks  []string `json:"failed_checks,omitempty"`
	TimeoutReason string   `json:"timeout_reason,omitempty"`
}

type RetryContext struct {
	Attempt       int
	Cfg           config.AsyncConfig
	Deadline      time.Time
	CurrentDelay  time.Duration
	LastErr       error
	LastResult    any
	FailedReasons []string
}

func (rc *RetryContext) CalculateNextDelay() time.Duration {
	delay := rc.CurrentDelay

	if rc.Cfg.Backoff.Enabled && rc.Attempt > 1 {
		delay = time.Duration(float64(delay) * rc.Cfg.Backoff.Factor)
		if delay > rc.Cfg.Backoff.MaxInterval {
			delay = rc.Cfg.Backoff.MaxInterval
		}
		rc.CurrentDelay = delay
	}

	if rc.Cfg.Jitter > 0 {
		jitterAmount := float64(delay) * rc.Cfg.Jitter
		jitterDelta := (rand.Float64()*2 - 1) * jitterAmount
		delay = time.Duration(float64(delay) + jitterDelta)
		if delay < 0 {
			delay = rc.Cfg.Interval
		}
	}

	return delay
}
