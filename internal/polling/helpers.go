package polling

import (
	"encoding/json"
	"fmt"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func AttachPollingSummary(stepCtx provider.StepCtx, summary PollingSummary) {
	summaryJSON, _ := json.MarshalIndent(summary, "", "  ")
	stepCtx.WithNewAttachment("Polling Summary", allure.JSON, summaryJSON)
}

func AttachIfAsync(stepCtx provider.StepCtx, summary PollingSummary) {
	if GetStepMode(stepCtx) == AsyncMode {
		AttachPollingSummary(stepCtx, summary)
	}
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
