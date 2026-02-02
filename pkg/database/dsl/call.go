package dsl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/allure"
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/validation"
	"github.com/gorelov-m-v/go-test-framework/pkg/database/client"
)

// Query represents a database query builder with fluent interface.
// It supports SQL queries, expectations on columns and rows, and automatic retry in async mode.
//
// Type parameter T must be a struct with `db` tags for column mapping.
//
// Example:
//
//	dsl.NewQuery[models.User](sCtx, dbClient).
//	    SQL("SELECT * FROM users WHERE id = ?", userID).
//	    ExpectFound().
//	    ExpectColumnEquals("status", "active").
//	    Send()
type Query[T any] struct {
	sCtx            provider.StepCtx
	client          *client.Client
	ctx             context.Context
	sql             string
	args            []any
	expectations    []*expect.Expectation[T]
	expectationsAll []*expect.Expectation[[]T]
	expectsNotFound bool
	scannedResult   T
	scannedResults  []T
	lastError       error
}

// NewQuery creates a new database query builder.
// The type parameter T must be a struct with `db` tags for column mapping.
//
// Parameters:
//   - sCtx: Allure step context for test reporting
//   - dbClient: Database client with connection pool
//
// Returns a Query builder that can be configured with SQL and expectations.
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

// SQL sets the SQL query and its arguments.
// Use ? for MySQL placeholders or $1, $2 for PostgreSQL.
func (q *Query[T]) SQL(query string, args ...any) *Query[T] {
	q.sql = query
	q.args = args
	return q
}

func (q *Query[T]) validate() {
	v := validation.New(q.sCtx, "DB")
	v.RequireNotNil(q.client, "Database client")
	v.RequireNotEmptyWithHint(q.sql, "SQL query", "Use .SQL(\"SELECT...\", args...).")
}

// Send executes the query expecting a single row result.
// In async mode (AsyncStep), automatically retries with backoff until expectations pass.
// Returns the scanned struct. Use ExpectNotFound if no rows is the expected outcome.
func (q *Query[T]) Send() T {
	q.validate()

	q.sCtx.WithNewStep(q.stepName(), func(stepCtx provider.StepCtx) {
		result, duration, err, summary := q.execute(stepCtx, q.expectations)
		q.scannedResult = result
		q.lastError = err

		rowCount := 0
		if err == nil {
			rowCount = 1
		}
		attachSQLReport(stepCtx, q.client, allure.SQLAttachParams{
			Query:    q.sql,
			Args:     q.args,
			Result:   q.scannedResult,
			RowCount: rowCount,
			Duration: duration,
			Error:    err,
		}, summary)
		q.assertResults(stepCtx, err)
	})

	return q.scannedResult
}

func (q *Query[T]) stepName() string {
	tableName := extractTableName(q.sql)
	return fmt.Sprintf("SELECT %s", tableName)
}

func (q *Query[T]) assertResults(stepCtx provider.StepCtx, err error) {
	expect.AssertExpectations(stepCtx, q.expectations, err, q.scannedResult, q.assertNoExpectations)
}

func (q *Query[T]) assertNoExpectations(stepCtx provider.StepCtx, mode polling.AssertionMode, err error) {
	if err == nil {
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		if !q.expectsNotFound {
			polling.NoError(stepCtx, mode, err, "Expected row to exist, but got sql.ErrNoRows. Use ExpectNotFound() if 'not found' is expected")
		}
		return
	}

	polling.NoError(stepCtx, mode, err, "DB query failed: %v", err)
}

// SendAll executes the query expecting multiple row results.
// In async mode (AsyncStep), automatically retries with backoff until expectations pass.
// Returns a slice of scanned structs.
func (q *Query[T]) SendAll() []T {
	q.validate()

	q.sCtx.WithNewStep(q.stepNameAll(), func(stepCtx provider.StepCtx) {
		results, duration, err, summary := q.executeAll(stepCtx)
		q.scannedResults = results
		q.lastError = err

		attachSQLReport(stepCtx, q.client, allure.SQLAttachParams{
			Query:    q.sql,
			Args:     q.args,
			Result:   q.scannedResults,
			RowCount: len(q.scannedResults),
			Duration: duration,
			Error:    err,
		}, summary)
		q.assertResultsAll(stepCtx, err)
	})

	return q.scannedResults
}

func (q *Query[T]) stepNameAll() string {
	tableName := extractTableName(q.sql)
	return fmt.Sprintf("SELECT %s (all)", tableName)
}

func (q *Query[T]) assertResultsAll(stepCtx provider.StepCtx, err error) {
	expect.AssertExpectations(stepCtx, q.expectationsAll, err, q.scannedResults, q.assertNoExpectationsAll)
}

func (q *Query[T]) assertNoExpectationsAll(stepCtx provider.StepCtx, mode polling.AssertionMode, err error) {
	if err != nil {
		polling.NoError(stepCtx, mode, err, "DB query failed: %v", err)
	}
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
