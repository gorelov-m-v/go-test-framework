package dsl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/config"
	"go-test-framework/pkg/database/client"
	"go-test-framework/pkg/extension"
)

type Query[T any] struct {
	sCtx            provider.StepCtx
	client          *client.Client
	ctx             context.Context
	sql             string
	args            []any
	expectations    []*expectation
	expectsNotFound bool
	scannedResult   T
	sqlResult       sql.Result
	asyncCfg        config.AsyncConfig
}

func NewQuery[T any](sCtx provider.StepCtx, dbClient *client.Client) *Query[T] {
	return &Query[T]{
		sCtx:   sCtx,
		client: dbClient,
		ctx:    context.Background(),
		asyncCfg: config.AsyncConfig{
			Enabled:  dbClient.AsyncConfig.Enabled,
			Timeout:  dbClient.AsyncConfig.Timeout,
			Interval: dbClient.AsyncConfig.Interval,
			Backoff: config.BackoffConfig{
				Enabled:     dbClient.AsyncConfig.Backoff.Enabled,
				Factor:      dbClient.AsyncConfig.Backoff.Factor,
				MaxInterval: dbClient.AsyncConfig.Backoff.MaxInterval,
			},
			Jitter: dbClient.AsyncConfig.Jitter,
		},
	}
}

func (q *Query[T]) SQL(query string, args ...any) *Query[T] {
	q.sql = query
	q.args = args
	return q
}

func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
	q.ctx = ctx
	return q
}

func (q *Query[T]) MustFetch() T {
	tableName := extractTableName(q.sql)
	stepName := fmt.Sprintf("SELECT %s", tableName)

	q.sCtx.WithNewStep(stepName, func(stepCtx provider.StepCtx) {
		attachQuery(stepCtx, q.sql, q.args)

		mode := extension.GetStepMode(stepCtx)
		var result T
		var err error
		var summary extension.PollingSummary

		if mode == extension.AsyncMode {
			result, err, summary = q.executeWithRetry(stepCtx, q.expectations)
			extension.AttachPollingSummary(stepCtx, summary)
		} else {
			result, err, summary = q.executeSingle()
		}

		q.scannedResult = result

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				stepCtx.Logf("Query returned no rows")
			}
			attachResult(stepCtx, q.client, nil, err)
		} else {
			attachResult(stepCtx, q.client, result, nil)
		}

		assertMd := extension.GetAssertionModeFromStepMode(mode)

		if len(q.expectations) > 0 {
			reportExpectations(stepCtx, assertMd, q.expectations, err, result)
		}

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if !q.expectsNotFound {
					extension.NoError(stepCtx, assertMd, err, "Expected row to exist, but got sql.ErrNoRows%s. Use ExpectNotFound() if 'not found' is expected", func() string {
						if mode == extension.AsyncMode {
							return " after retry"
						}
						return ""
					}())
				}
			} else {
				msg := "DB query failed"
				if mode == extension.AsyncMode {
					msg = extension.FinalFailureMessage(summary)
				}
				extension.NoError(stepCtx, assertMd, err, msg)
			}
		} else if !summary.Success {
			msg := "DB query expectations not met"
			if mode == extension.AsyncMode {
				msg = extension.FinalFailureMessage(summary)
			}
			extension.True(stepCtx, assertMd, false, msg)
		}
	})

	return q.scannedResult
}

func (q *Query[T]) MustExec() sql.Result {
	tableName := extractTableName(q.sql)
	operation := extractOperation(q.sql)
	stepName := fmt.Sprintf("%s %s", operation, tableName)

	q.sCtx.WithNewStep(stepName, func(stepCtx provider.StepCtx) {
		attachQuery(stepCtx, q.sql, q.args)

		mode := extension.GetStepMode(stepCtx)
		assertMd := extension.GetAssertionModeFromStepMode(mode)

		if len(q.expectations) > 0 {
			extension.True(stepCtx, assertMd, false, "MustExec() cannot be used with expectations (ExpectColumn*, ExpectFound, ExpectNotFound). Expectations are only valid for MustFetch()")
			return
		}

		res, err := q.client.DB.ExecContext(q.ctx, q.sql, q.args...)
		q.sqlResult = res

		attachExecResult(stepCtx, res, err)

		if err != nil {
			extension.NoError(stepCtx, assertMd, err, "DB exec failed")
		}
	})

	return q.sqlResult
}

func getFieldValueByColumnName(target any, columnName string) (any, error) {
	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, fmt.Errorf("target pointer is nil")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target is not a struct")
	}

	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("db")

		tagParts := strings.SplitN(tag, ",", 2)
		if tagParts[0] == columnName {
			fieldVal := val.Field(i)
			if !fieldVal.CanInterface() {
				return nil, fmt.Errorf("cannot interface field %s", field.Name)
			}
			return fieldVal.Interface(), nil
		}
	}

	return nil, fmt.Errorf("no field with db tag '%s' found in struct %T", columnName, target)
}

func extractTableName(query string) string {
	query = strings.TrimSpace(query)
	upper := strings.ToUpper(query)

	fromIdx := strings.Index(upper, " FROM ")
	if fromIdx != -1 {
		afterFrom := query[fromIdx+6:]
		words := strings.Fields(afterFrom)
		if len(words) > 0 {
			tableName := words[0]
			tableName = strings.Trim(tableName, "`'\"")
			if dotIdx := strings.LastIndex(tableName, "."); dotIdx != -1 {
				tableName = tableName[dotIdx+1:]
			}
			return tableName
		}
	}

	intoIdx := strings.Index(upper, " INTO ")
	if intoIdx != -1 {
		afterInto := query[intoIdx+6:]
		words := strings.Fields(afterInto)
		if len(words) > 0 {
			tableName := words[0]
			tableName = strings.Trim(tableName, "`'\"")
			if dotIdx := strings.LastIndex(tableName, "."); dotIdx != -1 {
				tableName = tableName[dotIdx+1:]
			}
			return tableName
		}
	}

	updateIdx := strings.Index(upper, "UPDATE ")
	if updateIdx != -1 {
		afterUpdate := query[updateIdx+7:]
		words := strings.Fields(afterUpdate)
		if len(words) > 0 {
			tableName := words[0]
			tableName = strings.Trim(tableName, "`'\"")
			if dotIdx := strings.LastIndex(tableName, "."); dotIdx != -1 {
				tableName = tableName[dotIdx+1:]
			}
			return tableName
		}
	}

	return "query"
}

func extractOperation(query string) string {
	query = strings.TrimSpace(strings.ToUpper(query))

	if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	}
	if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	}
	if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	}
	if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	}

	return "EXEC"
}
