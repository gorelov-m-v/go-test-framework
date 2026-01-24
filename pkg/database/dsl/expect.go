package dsl

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/typeconv"
)

var structMapper = reflectx.NewMapper("db")

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

func (q *Query[T]) ExpectFound() *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectFound() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeFoundExpectation[T]())
	return q
}

func (q *Query[T]) ExpectNotFound() *Query[T] {
	if len(q.expectations) > 0 {
		q.sCtx.Break("DB DSL Error: ExpectNotFound() cannot be used after other expectations (ExpectFound, ExpectColumnEquals, etc.)")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectsNotFound = true
	q.expectations = []*expect.Expectation[T]{}
	q.expectations = append(q.expectations, makeNotFoundExpectation[T]())
	return q
}

func (q *Query[T]) ExpectColumnEquals(columnName string, expectedValue any) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnEquals() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnEqualsExpectation[T](columnName, expectedValue))
	return q
}

func (q *Query[T]) ExpectColumnNotEquals(columnName string, notExpectedValue any) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnNotEquals() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnNotEqualsExpectation[T](columnName, notExpectedValue))
	return q
}

func (q *Query[T]) ExpectColumnNotEmpty(columnName string) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnNotEmpty() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnNotEmptyExpectation[T](columnName))
	return q
}

func (q *Query[T]) ExpectColumnIsNull(columnName string) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnIsNull() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnIsNullExpectation[T](columnName))
	return q
}

func (q *Query[T]) ExpectColumnEmpty(columnName string) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnEmpty() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnEmptyExpectation[T](columnName))
	return q
}

func (q *Query[T]) ExpectColumnIsNotNull(columnName string) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnIsNotNull() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnIsNotNullExpectation[T](columnName))
	return q
}

func (q *Query[T]) ExpectColumnTrue(columnName string) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnTrue() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnTrueExpectation[T](columnName))
	return q
}

func (q *Query[T]) ExpectColumnFalse(columnName string) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnFalse() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnFalseExpectation[T](columnName))
	return q
}

func (q *Query[T]) ExpectColumnJSONEquals(columnName string, expected map[string]interface{}) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectColumnJSONEquals() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeColumnJSONEqualsExpectation[T](columnName, expected))
	return q
}

func (q *Query[T]) ExpectColumnJsonEquals(columnName string, expected map[string]interface{}) *Query[T] {
	return q.ExpectColumnJSONEquals(columnName, expected)
}

func (q *Query[T]) ExpectRow(expected T) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectRow() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeRowExpectation[T](expected))
	return q
}

func (q *Query[T]) ExpectRowPartial(expected T) *Query[T] {
	if q.expectsNotFound {
		q.sCtx.Break("DB DSL Error: ExpectRowPartial() cannot be used with ExpectNotFound()")
		q.sCtx.BrokenNow()
		return q
	}
	q.expectations = append(q.expectations, makeRowPartialExpectation[T](expected))
	return q
}

func makeFoundExpectation[T any]() *expect.Expectation[T] {
	name := "Expect: Found"
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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
			if errors.Is(err, sql.ErrNoRows) {
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
	name := fmt.Sprintf("Expect: Column '%s' = %v", columnName, expectedValue)
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			equal, retryable, reason := equalsLoose(expectedValue, actualValue)
			if !equal {
				return polling.CheckResult{
					Ok:        false,
					Retryable: retryable,
					Reason:    fmt.Sprintf("Column '%s': %s", columnName, reason),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[T](name),
	)
}

func makeColumnNotEmptyExpectation[T any](columnName string) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' not empty", columnName)
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			isEmpty := typeconv.IsEmpty(actualValue)
			if isEmpty {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to not be empty, but it is", columnName),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[T](name),
	)
}

func makeColumnIsNullExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' IS NULL", columnName),
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			isNull := typeconv.IsNull(actualValue)
			if !isNull {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be NULL, but it has a value", columnName),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' IS NULL] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			isNull := typeconv.IsNull(actualValue)
			a.Equal(true, isNull, "[Expect: Column '%s' IS NULL]", columnName)
		},
	)
}

func makeColumnEmptyExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' IS EMPTY", columnName),
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			isEmpty := typeconv.IsEmpty(actualValue)
			if !isEmpty {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be empty, but got: %v", columnName, actualValue),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' IS EMPTY] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			isEmpty := typeconv.IsEmpty(actualValue)
			a.Equal(true, isEmpty, "[Expect: Column '%s' IS EMPTY]", columnName)
		},
	)
}

func makeColumnIsNotNullExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' IS NOT NULL", columnName),
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			isNull := typeconv.IsNull(actualValue)
			if isNull {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be NOT NULL, but it is NULL", columnName),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' IS NOT NULL] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			isNull := typeconv.IsNull(actualValue)
			a.Equal(false, isNull, "[Expect: Column '%s' IS NOT NULL]", columnName)
		},
	)
}

func makeColumnTrueExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' = true", columnName),
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			if typeconv.IsNull(actualValue) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
				}
			}

			b, ok := typeconv.ToBool(actualValue)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Column '%s' is not a boolean/numeric(0/1) type", columnName),
				}
			}

			if !b {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be true, but got false", columnName),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' = true] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			b, _ := typeconv.ToBool(actualValue)
			a.Equal(true, b, "[Expect: Column '%s' = true]", columnName)
		},
	)
}

func makeColumnFalseExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' = false", columnName),
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			if typeconv.IsNull(actualValue) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
				}
			}

			b, ok := typeconv.ToBool(actualValue)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Column '%s' is not a boolean/numeric(0/1) type", columnName),
				}
			}

			if b {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be false, but got true", columnName),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' = false] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			b, _ := typeconv.ToBool(actualValue)
			a.Equal(false, b, "[Expect: Column '%s' = false]", columnName)
		},
	)
}

func makeColumnNotEqualsExpectation[T any](columnName string, notExpectedValue any) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' != %v", columnName, notExpectedValue)
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			equal, _, _ := equalsLoose(notExpectedValue, actualValue)

			if equal {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Column '%s' equals %v, but expected NOT to equal", columnName, actualValue),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[T](name),
	)
}

func makeColumnJSONEqualsExpectation[T any](columnName string, expected map[string]interface{}) *expect.Expectation[T] {
	name := fmt.Sprintf("Expect: Column '%s' JSON = %v", columnName, expected)
	return expect.New(
		name,
		func(err error, result T) polling.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
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

			var actualMap map[string]interface{}
			switch v := actualValue.(type) {
			case json.RawMessage:
				if err := json.Unmarshal(v, &actualMap); err != nil {
					return polling.CheckResult{
						Ok:        false,
						Retryable: false,
						Reason:    fmt.Sprintf("Failed to unmarshal JSON: %v", err),
					}
				}
			case []byte:
				if err := json.Unmarshal(v, &actualMap); err != nil {
					return polling.CheckResult{
						Ok:        false,
						Retryable: false,
						Reason:    fmt.Sprintf("Failed to unmarshal JSON: %v", err),
					}
				}
			case string:
				if err := json.Unmarshal([]byte(v), &actualMap); err != nil {
					return polling.CheckResult{
						Ok:        false,
						Retryable: false,
						Reason:    fmt.Sprintf("Failed to unmarshal JSON string: %v", err),
					}
				}
			default:
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Column '%s' is not a JSON type, got %T", columnName, actualValue),
				}
			}

			for key, expectedVal := range expected {
				actualVal, exists := actualMap[key]
				if !exists {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Key '%s' not found in JSON", key),
					}
				}
				if fmt.Sprintf("%v", expectedVal) != fmt.Sprintf("%v", actualVal) {
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Key '%s': expected '%v', got '%v'", key, expectedVal, actualVal),
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
				if errors.Is(err, sql.ErrNoRows) {
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
				if errors.Is(err, sql.ErrNoRows) {
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
