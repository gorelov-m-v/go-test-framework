package dsl

import (
	"go-test-framework/internal/allure"
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
