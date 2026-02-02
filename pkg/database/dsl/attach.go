package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/allure"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/database/client"
)

var sqlReporter = allure.NewDefaultReporter()

func attachSQLReport(
	stepCtx provider.StepCtx,
	dbClient *client.Client,
	params allure.SQLAttachParams,
	pollingSummary polling.PollingSummary,
) {
	report := allure.SQLReportDTO{
		Request: allure.ToSQLRequestDTO(params.Query, params.Args),
		Result: allure.ToSQLResultDTO(allure.SQLResultParams{
			Result:   params.Result,
			RowCount: params.RowCount,
			Duration: params.Duration,
			Error:    params.Error,
		}),
		Polling: allure.ToPollingSummaryDTO(pollingSummary),
	}

	sqlReporter.AttachSQLReport(stepCtx, dbClient, report)
}
