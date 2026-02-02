package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/allure"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

var kafkaReporter = allure.NewDefaultReporter()

func attachKafkaReport[T any](
	stepCtx provider.StepCtx,
	q *Query[T],
	pollingSummary polling.PollingSummary,
) {
	report := allure.KafkaReportDTO{
		Search: allure.ToKafkaSearchDTO(q.topicName, q.filters, q.client.GetDefaultTimeout(), q.unique),
		Result: allure.ToKafkaResultDTO(allure.KafkaResultParams{
			Found:           q.found,
			MessageBytes:    q.messageBytes,
			AllMatchingMsgs: q.allMatchingMsgs,
			ExpectedCount:   q.expectedCount,
		}),
		Polling: allure.ToPollingSummaryDTO(pollingSummary),
	}

	kafkaReporter.AttachKafkaReport(stepCtx, report)
}
