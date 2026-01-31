package dsl

import (
	"database/sql"
	"errors"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/allure"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/database/client"
)

var sqlReporter = allure.NewDefaultReporter()

func attachSQLReport(
	stepCtx provider.StepCtx,
	dbClient *client.Client,
	sqlQuery string,
	args []any,
	result any,
	rowCount int,
	duration time.Duration,
	err error,
	pollingSummary polling.PollingSummary,
) {
	reqDTO := allure.SQLRequestDTO{
		Query: sqlQuery,
		Args:  args,
	}

	resultDTO := allure.SQLResultDTO{
		Duration: duration,
		Error:    err,
	}

	if err == nil {
		resultDTO.Found = true
		resultDTO.RowCount = rowCount
		resultDTO.Data = result
	} else if errors.Is(err, sql.ErrNoRows) {
		resultDTO.Found = false
	}

	report := allure.SQLReportDTO{
		Request: reqDTO,
		Result:  resultDTO,
		Polling: allure.ToPollingSummaryDTO(pollingSummary),
	}

	sqlReporter.AttachSQLReport(stepCtx, dbClient, report)
}
