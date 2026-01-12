package dsl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/database/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
)

var structMapper = reflectx.NewMapper("db")

type Query[T any] struct {
	sCtx            provider.StepCtx
	client          *client.Client
	ctx             context.Context
	sql             string
	args            []any
	expectations    []*expect.Expectation[T]
	expectsNotFound bool
	scannedResult   T
}

func NewQuery[T any](sCtx provider.StepCtx, dbClient *client.Client) *Query[T] {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil || t.Kind() != reflect.Struct {
		sCtx.Break(fmt.Sprintf("DB DSL Error: Query type parameter must be a struct, got %T. Check your NewQuery[T] generic type.", zero))
		sCtx.BrokenNow()
		return nil
	}

	return &Query[T]{
		sCtx:   sCtx,
		client: dbClient,
		ctx:    context.Background(),
	}
}

func (q *Query[T]) SQL(query string, args ...any) *Query[T] {
	q.sql = query
	q.args = args
	return q
}

func (q *Query[T]) WithContext(ctx context.Context) *Query[T] {
	if ctx != nil {
		q.ctx = ctx
	}
	return q
}

func (q *Query[T]) Send() T {
	tableName := extractTableName(q.sql)
	stepName := fmt.Sprintf("SELECT %s", tableName)

	q.sCtx.WithNewStep(stepName, func(stepCtx provider.StepCtx) {
		attachQuery(stepCtx, q.sql, q.args)

		mode := extension.GetStepMode(stepCtx)
		var result T
		var err error
		var summary extension.PollingSummary

		useRetry := mode == extension.AsyncMode && len(q.expectations) > 0

		if useRetry {
			result, err, summary = q.executeWithRetry(stepCtx, q.expectations)
		} else {
			result, err, summary = q.executeSingle()
		}

		if mode == extension.AsyncMode {
			extension.AttachPollingSummary(stepCtx, summary)
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
			expect.ReportAll(stepCtx, assertMd, q.expectations, err, result)
		} else {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					if !q.expectsNotFound {
						suffix := ""
						if useRetry {
							suffix = " after retry"
						}
						extension.NoError(stepCtx, assertMd, err, "Expected row to exist, but got sql.ErrNoRows%s. Use ExpectNotFound() if 'not found' is expected", suffix)
					}
				} else {
					msg := "DB query failed"
					if mode == extension.AsyncMode {
						msg = extension.FinalFailureMessage(summary)
					}
					extension.NoError(stepCtx, assertMd, err, msg)
				}
			}
		}
	})

	return q.scannedResult
}

func getFieldValueByColumnName(target any, columnName string) (any, error) {
	columnName = strings.TrimSpace(columnName)
	if columnName == "" {
		return nil, fmt.Errorf("columnName cannot be empty")
	}

	if target == nil {
		return nil, fmt.Errorf("target is nil")
	}

	v := reflect.ValueOf(target)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("target pointer is nil")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("target is not a struct, got %s", v.Kind())
	}

	fieldMap := structMapper.FieldMap(v)

	fieldValue, found := fieldMap[columnName]
	if !found {
		return nil, fmt.Errorf("column '%s' not found in struct %T (check 'db' tags)", columnName, target)
	}

	if !fieldValue.CanInterface() {
		return nil, fmt.Errorf("field for column '%s' is unexported", columnName)
	}

	return fieldValue.Interface(), nil
}

func extractTableName(query string) string {
	query = strings.TrimSpace(query)
	upper := strings.ToUpper(query)

	if strings.HasPrefix(upper, "WITH") {
		parenDepth := 0
		for i, char := range upper {
			switch char {
			case '(':
				parenDepth++
			case ')':
				parenDepth--
			}

			if parenDepth == 0 && i > 0 {
				remaining := upper[i:]
				if strings.HasPrefix(remaining, "SELECT") {
					return extractTableName(query[i:])
				}
			}
		}
	}

	if tableName := extractTableFromKeyword(query, upper, "FROM"); tableName != "" {
		return tableName
	}

	return "query"
}

func extractTableFromKeyword(query, upper, keyword string) string {
	tokens := strings.Fields(upper)

	for i, token := range tokens {
		if token == keyword && i+1 < len(tokens) {
			keywordPos := strings.Index(upper, keyword)
			if keywordPos == -1 {
				continue
			}

			afterKeyword := query[keywordPos+len(keyword):]
			afterKeyword = strings.TrimSpace(afterKeyword)

			words := strings.Fields(afterKeyword)
			if len(words) > 0 {
				return cleanTableName(words[0])
			}
		}
	}

	return ""
}

func cleanTableName(tableName string) string {
	tableName = strings.Trim(tableName, "`'\"")

	if dotIdx := strings.LastIndex(tableName, "."); dotIdx != -1 {
		tableName = tableName[dotIdx+1:]
		tableName = strings.Trim(tableName, "`'\"")
	}

	tableName = strings.TrimRight(tableName, ",;()")

	return tableName
}
