package dsl

import (
	"database/sql"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"

	"github.com/gorelov-m-v/go-test-framework/internal/errors"
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/typeconv"
)

var structMapper = reflectx.NewMapper("db")

func (q *Query[T]) breakIfNotFound(methodName string) bool {
	if q.expectsNotFound {
		q.stepCtx.Break(errors.ConflictingExpectations("DB", methodName, "ExpectNotFound()"))
		q.stepCtx.BrokenNow()
		return true
	}
	return false
}

func (q *Query[T]) addExpectation(exp *expect.Expectation[T]) {
	expect.AddExpectation(q.stepCtx, q.sent, &q.expectations, exp, "DB")
}

func (q *Query[T]) addExpectationAll(exp *expect.Expectation[[]T]) {
	expect.AddExpectation(q.stepCtx, q.sent, &q.expectationsAll, exp, "DB")
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

	if !v.CanAddr() {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		v = ptr.Elem()
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

func getFieldValue[T any](result T, columnName string) (any, error) {
	return getFieldValueByColumnName(result, columnName)
}

// ExpectFound checks that the query returns at least one row.
func (q *Query[T]) ExpectFound() *Query[T] {
	if q.breakIfNotFound("ExpectFound()") {
		return q
	}
	q.addExpectation(makeFoundExpectation[T]())
	return q
}

// ExpectNotFound checks that the query returns no rows (sql.ErrNoRows).
// Cannot be combined with other expectations.
func (q *Query[T]) ExpectNotFound() *Query[T] {
	if q.sent {
		q.stepCtx.Break(errors.ExpectationsAfterSend("DB"))
		q.stepCtx.BrokenNow()
		return q
	}
	if len(q.expectations) > 0 {
		q.stepCtx.Break(errors.ExpectationOrder("DB", "ExpectNotFound()", "other expectations (ExpectFound, ExpectColumnEquals, etc.)"))
		q.stepCtx.BrokenNow()
		return q
	}
	q.expectsNotFound = true
	q.expectations = []*expect.Expectation[T]{}
	q.addExpectation(makeNotFoundExpectation[T]())
	return q
}

// ExpectColumnEquals checks that a column value equals the expected value.
// Column name must match the `db` tag on the struct field. Supports numeric type coercion.
func (q *Query[T]) ExpectColumnEquals(columnName string, expectedValue any) *Query[T] {
	if q.breakIfNotFound("ExpectColumnEquals()") {
		return q
	}
	q.addExpectation(makeColumnEqualsExpectation[T](columnName, expectedValue))
	return q
}

// ExpectColumnNotEquals checks that a column value does not equal the specified value.
func (q *Query[T]) ExpectColumnNotEquals(columnName string, notExpectedValue any) *Query[T] {
	if q.breakIfNotFound("ExpectColumnNotEquals()") {
		return q
	}
	q.addExpectation(makeColumnNotEqualsExpectation[T](columnName, notExpectedValue))
	return q
}

// ExpectColumnNotEmpty checks that a column value is not empty.
func (q *Query[T]) ExpectColumnNotEmpty(columnName string) *Query[T] {
	if q.breakIfNotFound("ExpectColumnNotEmpty()") {
		return q
	}
	q.addExpectation(makeColumnNotEmptyExpectation[T](columnName))
	return q
}

// ExpectColumnIsNull checks that a column value is NULL.
func (q *Query[T]) ExpectColumnIsNull(columnName string) *Query[T] {
	if q.breakIfNotFound("ExpectColumnIsNull()") {
		return q
	}
	q.addExpectation(makeColumnIsNullExpectation[T](columnName))
	return q
}

// ExpectColumnEmpty checks that a column value is empty (zero value).
func (q *Query[T]) ExpectColumnEmpty(columnName string) *Query[T] {
	if q.breakIfNotFound("ExpectColumnEmpty()") {
		return q
	}
	q.addExpectation(makeColumnEmptyExpectation[T](columnName))
	return q
}

// ExpectColumnIsNotNull checks that a column value is not NULL.
func (q *Query[T]) ExpectColumnIsNotNull(columnName string) *Query[T] {
	if q.breakIfNotFound("ExpectColumnIsNotNull()") {
		return q
	}
	q.addExpectation(makeColumnIsNotNullExpectation[T](columnName))
	return q
}

// ExpectColumnTrue checks that a boolean column value is true.
func (q *Query[T]) ExpectColumnTrue(columnName string) *Query[T] {
	if q.breakIfNotFound("ExpectColumnTrue()") {
		return q
	}
	q.addExpectation(makeColumnTrueExpectation[T](columnName))
	return q
}

// ExpectColumnFalse checks that a boolean column value is false.
func (q *Query[T]) ExpectColumnFalse(columnName string) *Query[T] {
	if q.breakIfNotFound("ExpectColumnFalse()") {
		return q
	}
	q.addExpectation(makeColumnFalseExpectation[T](columnName))
	return q
}

// ExpectColumnJSON checks that a JSON column contains all expected key-value pairs.
func (q *Query[T]) ExpectColumnJSON(columnName string, expected map[string]interface{}) *Query[T] {
	if q.breakIfNotFound("ExpectColumnJSON()") {
		return q
	}
	q.addExpectation(makeColumnJSONEqualsExpectation[T](columnName, expected))
	return q
}

// ExpectRowEquals checks that the row exactly matches the expected struct (all fields must match).
func (q *Query[T]) ExpectRowEquals(expected T) *Query[T] {
	if q.breakIfNotFound("ExpectRowEquals()") {
		return q
	}
	q.addExpectation(makeRowExpectation[T](expected))
	return q
}

// ExpectRowPartial checks that the row contains fields from the expected struct (non-zero fields only).
func (q *Query[T]) ExpectRowPartial(expected T) *Query[T] {
	if q.breakIfNotFound("ExpectRowPartial()") {
		return q
	}
	q.addExpectation(makeRowPartialExpectation[T](expected))
	return q
}

// ExpectCountAll checks that the query returns exactly the specified number of rows. Use with SendAll().
func (q *Query[T]) ExpectCountAll(count int) *Query[T] {
	q.addExpectationAll(makeCountAllExpectation[T](count))
	return q
}

// ExpectNotEmptyAll checks that the query returns at least one row. Use with SendAll().
func (q *Query[T]) ExpectNotEmptyAll() *Query[T] {
	q.addExpectationAll(makeNotEmptyAllExpectation[T]())
	return q
}

func makeFoundExpectation[T any]() *expect.Expectation[T] {
	name := "Expect: Found"
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if stderrors.Is(err, sql.ErrNoRows) {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    "Expected query to return at least one result, but got sql.ErrNoRows",
					}
				}
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[T](name),
	)
}

func makeNotFoundExpectation[T any]() *expect.Expectation[T] {
	name := "Expect: Not Found"
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if stderrors.Is(err, sql.ErrNoRows) {
				return polling.CheckResult{Ok: true}
			}
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected sql.ErrNoRows, but got different error: %v", err),
				}
			}
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Expected sql.ErrNoRows, but query returned results",
			}
		},
		expect.StandardReport[T](name),
	)
}

func makeColumnEqualsExpectation[T any](columnName string, expectedValue any) *expect.Expectation[T] {
	return expect.BuildColumnExpectation(expect.ColumnExpectationConfig[T]{
		ColumnName: columnName,
		ExpectName: fmt.Sprintf("Expect: Column '%s' = %v", columnName, expectedValue),
		GetValue:   getFieldValue[T],
		ErrNoRows:  sql.ErrNoRows,
		Check:      expect.CheckEquals(expectedValue, equalsLoose),
	})
}

func makeColumnNotEqualsExpectation[T any](columnName string, notExpectedValue any) *expect.Expectation[T] {
	return expect.BuildColumnExpectation(expect.ColumnExpectationConfig[T]{
		ColumnName: columnName,
		ExpectName: fmt.Sprintf("Expect: Column '%s' != %v", columnName, notExpectedValue),
		GetValue:   getFieldValue[T],
		ErrNoRows:  sql.ErrNoRows,
		Check:      expect.CheckNotEquals(notExpectedValue, equalsLoose),
	})
}

func makeColumnNotEmptyExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.BuildColumnExpectation(expect.ColumnExpectationConfig[T]{
		ColumnName: columnName,
		ExpectName: fmt.Sprintf("Expect: Column '%s' not empty", columnName),
		GetValue:   getFieldValue[T],
		ErrNoRows:  sql.ErrNoRows,
		Check:      expect.CheckNotEmpty(),
	})
}

func makeColumnEmptyExpectation[T any](columnName string) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' IS EMPTY", columnName)
	return expect.BuildColumnEmptyExpectation(expect.ColumnEmptyExpectationConfig[T]{
		ColumnName:    columnName,
		ExpectName:    name,
		GetValue:      getFieldValue[T],
		ErrNoRows:     sql.ErrNoRows,
		Check:         expect.CheckEmpty(),
		ExpectedEmpty: true,
		IsEmptyFunc:   typeconv.IsEmpty,
	})
}

func makeColumnIsNullExpectation[T any](columnName string) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' IS NULL", columnName)
	return expect.BuildColumnNullExpectation(expect.ColumnNullExpectationConfig[T]{
		ColumnName:   columnName,
		ExpectName:   name,
		GetValue:     getFieldValue[T],
		ErrNoRows:    sql.ErrNoRows,
		Check:        expect.CheckIsNull(),
		ExpectedNull: true,
		IsNullFunc:   typeconv.IsNull,
	})
}

func makeColumnIsNotNullExpectation[T any](columnName string) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' IS NOT NULL", columnName)
	return expect.BuildColumnNullExpectation(expect.ColumnNullExpectationConfig[T]{
		ColumnName:   columnName,
		ExpectName:   name,
		GetValue:     getFieldValue[T],
		ErrNoRows:    sql.ErrNoRows,
		Check:        expect.CheckIsNotNull(),
		ExpectedNull: false,
		IsNullFunc:   typeconv.IsNull,
	})
}

func makeColumnTrueExpectation[T any](columnName string) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' = true", columnName)
	return expect.BuildColumnBoolExpectation(expect.ColumnBoolExpectationConfig[T]{
		ColumnName:   columnName,
		ExpectName:   name,
		GetValue:     getFieldValue[T],
		ErrNoRows:    sql.ErrNoRows,
		Check:        expect.CheckTrue(),
		ExpectedBool: true,
		ToBoolFunc:   typeconv.ToBool,
	})
}

func makeColumnFalseExpectation[T any](columnName string) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' = false", columnName)
	return expect.BuildColumnBoolExpectation(expect.ColumnBoolExpectationConfig[T]{
		ColumnName:   columnName,
		ExpectName:   name,
		GetValue:     getFieldValue[T],
		ErrNoRows:    sql.ErrNoRows,
		Check:        expect.CheckFalse(),
		ExpectedBool: false,
		ToBoolFunc:   typeconv.ToBool,
	})
}

func makeColumnJSONEqualsExpectation[T any](columnName string, expected map[string]interface{}) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' JSON = %v", columnName, expected)
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if stderrors.Is(err, sql.ErrNoRows) {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			var jsonBytes []byte
			switch v := actualValue.(type) {
			case json.RawMessage:
				jsonBytes = v
			case []byte:
				jsonBytes = v
			case string:
				jsonBytes = []byte(v)
			default:
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Column '%s' is not a JSON type, got %T", columnName, actualValue),
				}
			}

			parsed, parseErr := jsonutil.Parse(jsonBytes)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to parse JSON: %v", parseErr),
				}
			}

			for key, expectedVal := range expected {
				field := parsed.Get(key)
				if !field.Exists() {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Key '%s' not found in JSON", key),
					}
				}
				ok, msg := jsonutil.Compare(field, expectedVal)
				if !ok {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Key '%s': %s", key, msg),
					}
				}
			}

			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[T](name),
	)
}

func makeRowExpectation[T any](expected T) *expect.Expectation[T] {
	name := "Expect: Row matches (exact)"
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if stderrors.Is(err, sql.ErrNoRows) {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    "Query returned no rows",
					}
				}
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}

			ok, msg := compareStructsExact(expected, result)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[T](name),
	)
}

func makeRowPartialExpectation[T any](expected T) *expect.Expectation[T] {
	name := "Expect: Row matches (partial)"
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if stderrors.Is(err, sql.ErrNoRows) {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    "Query returned no rows",
					}
				}
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}

			ok, msg := compareStructsPartial(expected, result)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[T](name),
	)
}

func compareStructsExact[T any](expected, actual T) (bool, string) {
	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	if expVal.Kind() == reflect.Ptr {
		if expVal.IsNil() {
			return actVal.Kind() == reflect.Ptr && actVal.IsNil(), "expected nil, got non-nil"
		}
		expVal = expVal.Elem()
	}
	if actVal.Kind() == reflect.Ptr {
		if actVal.IsNil() {
			return false, "expected value, got nil"
		}
		actVal = actVal.Elem()
	}

	if expVal.Kind() != reflect.Struct || actVal.Kind() != reflect.Struct {
		return false, fmt.Sprintf("expected struct types, got %s and %s", expVal.Kind(), actVal.Kind())
	}

	expType := expVal.Type()
	for i := 0; i < expVal.NumField(); i++ {
		field := expType.Field(i)
		if !field.IsExported() {
			continue
		}

		expFieldVal := expVal.Field(i)
		actFieldVal := actVal.Field(i)

		fieldName := getDBColumnName(field)

		equal, _, reason := equalsLoose(expFieldVal.Interface(), actFieldVal.Interface())
		if !equal {
			return false, fmt.Sprintf("field '%s': %s", fieldName, reason)
		}
	}

	return true, ""
}

func compareStructsPartial[T any](expected, actual T) (bool, string) {
	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	if expVal.Kind() == reflect.Ptr {
		if expVal.IsNil() {
			return true, ""
		}
		expVal = expVal.Elem()
	}
	if actVal.Kind() == reflect.Ptr {
		if actVal.IsNil() {
			return false, "expected value, got nil"
		}
		actVal = actVal.Elem()
	}

	if expVal.Kind() != reflect.Struct || actVal.Kind() != reflect.Struct {
		return false, fmt.Sprintf("expected struct types, got %s and %s", expVal.Kind(), actVal.Kind())
	}

	expType := expVal.Type()
	for i := 0; i < expVal.NumField(); i++ {
		field := expType.Field(i)
		if !field.IsExported() {
			continue
		}

		expFieldVal := expVal.Field(i)

		if expFieldVal.IsZero() {
			continue
		}

		actFieldVal := actVal.Field(i)
		fieldName := getDBColumnName(field)

		equal, _, reason := equalsLoose(expFieldVal.Interface(), actFieldVal.Interface())
		if !equal {
			return false, fmt.Sprintf("field '%s': %s", fieldName, reason)
		}
	}

	return true, ""
}

func getDBColumnName(field reflect.StructField) string {
	if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
		return tag
	}
	return field.Name
}

func makeCountAllExpectation[T any](expectedCount int) *expect.Expectation[[]T] {
	name := fmt.Sprintf("Expect: %d rows", expectedCount)
	return expect.New(
		name,
		func(err error, results []T) polling.CheckResult {
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    err.Error(),
				}
			}
			if len(results) != expectedCount {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("expected %d rows, got %d", expectedCount, len(results)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]T](name),
	)
}

func makeNotEmptyAllExpectation[T any]() *expect.Expectation[[]T] {
	name := "Expect: rows not empty"
	return expect.New(
		name,
		func(err error, results []T) polling.CheckResult {
			if err != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    err.Error(),
				}
			}
			if len(results) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "expected non-empty result, got 0 rows",
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]T](name),
	)
}
