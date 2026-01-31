package dsl

import (
	"encoding/json"

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
	searchDTO := allure.KafkaSearchDTO{
		Topic:   q.topicName,
		Filters: q.filters,
		Timeout: q.client.GetDefaultTimeout(),
		Unique:  q.unique,
	}

	resultDTO := buildKafkaResultDTO(q)

	report := allure.KafkaReportDTO{
		Search:  searchDTO,
		Result:  resultDTO,
		Polling: allure.ToPollingSummaryDTO(pollingSummary),
	}

	kafkaReporter.AttachKafkaReport(stepCtx, report)
}

func buildKafkaResultDTO[T any](q *Query[T]) allure.KafkaResultDTO {
	if !q.found {
		return allure.KafkaResultDTO{
			Found: false,
		}
	}

	if q.expectedCount > 0 && len(q.allMatchingMsgs) > 0 {
		return allure.KafkaResultDTO{
			Found:      true,
			MatchCount: len(q.allMatchingMsgs),
			Message:    parseMessagesToAny(q.allMatchingMsgs),
		}
	}

	var msgAny any
	if err := json.Unmarshal(q.messageBytes, &msgAny); err != nil {
		msgAny = string(q.messageBytes)
	}

	return allure.KafkaResultDTO{
		Found:      true,
		MatchCount: 1,
		Message:    msgAny,
		RawMessage: q.messageBytes,
	}
}

func parseMessagesToAny(messages [][]byte) any {
	if len(messages) == 0 {
		return nil
	}

	var parsed []any
	for _, msgBytes := range messages {
		var msgMap map[string]interface{}
		if err := json.Unmarshal(msgBytes, &msgMap); err == nil {
			parsed = append(parsed, msgMap)
		}
	}
	return parsed
}
