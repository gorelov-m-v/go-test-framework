package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/allure"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

var redisReporter = allure.NewDefaultReporter()

func attachRedisReport(
	stepCtx provider.StepCtx,
	q *Query,
	result *client.Result,
	pollingSummary polling.PollingSummary,
) {
	reqDTO := allure.RedisRequestDTO{
		Server: q.client.Addr(),
		Key:    q.key,
	}

	resultDTO := allure.RedisResultDTO{}
	if result != nil {
		resultDTO = allure.RedisResultDTO{
			Key:      result.Key,
			Exists:   result.Exists,
			Value:    result.Value,
			TTL:      result.TTL,
			Duration: result.Duration,
			Error:    result.Error,
		}
	}

	report := allure.RedisReportDTO{
		Request: reqDTO,
		Result:  resultDTO,
		Polling: allure.ToPollingSummaryDTO(pollingSummary),
	}

	redisReporter.AttachRedisReport(stepCtx, report)
}
