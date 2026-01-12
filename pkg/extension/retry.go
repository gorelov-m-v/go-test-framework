package extension

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

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

func AttachPollingSummary(stepCtx provider.StepCtx, summary PollingSummary) {
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	stepCtx.WithNewAttachment("Polling Summary", allure.JSON, summaryJSON)
}

func FinalFailureMessage(summary PollingSummary) string {
	if len(summary.FailedChecks) == 0 {
		return "Expectations not met within timeout"
	}

	msg := fmt.Sprintf("Expectations not met after %d attempts (%s):\n", summary.Attempts, summary.ElapsedTime)
	for i, reason := range summary.FailedChecks {
		msg += fmt.Sprintf("  [%d] %s\n", i+1, reason)
	}
	if summary.LastError != "" {
		msg += fmt.Sprintf("Last error: %s", summary.LastError)
	}
	return msg
}
