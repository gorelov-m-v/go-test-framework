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
	report := allure.RedisReportDTO{
		Request: allure.ToRedisRequestDTO(q.client.Addr(), q.key),
		Result:  allure.ToRedisResultDTO(result),
		Polling: allure.ToPollingSummaryDTO(pollingSummary),
	}

	redisReporter.AttachRedisReport(stepCtx, report)
}
