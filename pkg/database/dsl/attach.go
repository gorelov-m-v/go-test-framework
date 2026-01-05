package dsl

import (
	"database/sql"

	"go-test-framework/pkg/allure"
	"go-test-framework/pkg/database/client"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

var sqlReporter = allure.NewDefaultReporter()

func attachQuery(sCtx provider.StepCtx, sqlQuery string, args []any) {
	sqlReporter.AttachSQLQuery(sCtx, sqlQuery, args)
}

func attachResult(sCtx provider.StepCtx, dbClient *client.Client, result any, err error) {
	sqlReporter.AttachSQLResult(sCtx, dbClient, result, err)
}

func attachExecResult(sCtx provider.StepCtx, res sql.Result, err error) {
	sqlReporter.AttachSQLExecResult(sCtx, res, err)
}
