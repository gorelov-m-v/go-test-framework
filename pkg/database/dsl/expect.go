package dsl

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/internal/expect"
	"go-test-framework/pkg/extension"
)

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

func asBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case int:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int8:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int16:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int32:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case int64:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint8:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint16:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint32:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case uint64:
		if x == 0 {
			return false, true
		} else if x == 1 {
			return true, true
		}
	case sql.NullBool:
		if !x.Valid {
			return false, false
		}
		return x.Bool, true
	case *sql.NullBool:
		if x == nil || !x.Valid {
			return false, false
		}
		return x.Bool, true
	case sql.NullInt64:
		if !x.Valid {
			return false, false
		}
		if x.Int64 == 0 {
			return false, true
		} else if x.Int64 == 1 {
			return true, true
		}
	case *sql.NullInt64:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Int64 == 0 {
			return false, true
		} else if x.Int64 == 1 {
			return true, true
		}
	case sql.NullInt32:
		if !x.Valid {
			return false, false
		}
		if x.Int32 == 0 {
			return false, true
		} else if x.Int32 == 1 {
			return true, true
		}
	case *sql.NullInt32:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Int32 == 0 {
			return false, true
		} else if x.Int32 == 1 {
			return true, true
		}
	case sql.NullInt16:
		if !x.Valid {
			return false, false
		}
		if x.Int16 == 0 {
			return false, true
		} else if x.Int16 == 1 {
			return true, true
		}
	case *sql.NullInt16:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Int16 == 0 {
			return false, true
		} else if x.Int16 == 1 {
			return true, true
		}
	case sql.NullByte:
		if !x.Valid {
			return false, false
		}
		if x.Byte == 0 {
			return false, true
		} else if x.Byte == 1 {
			return true, true
		}
	case *sql.NullByte:
		if x == nil || !x.Valid {
			return false, false
		}
		if x.Byte == 0 {
			return false, true
		} else if x.Byte == 1 {
			return true, true
		}
	}
	return false, false
}

func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}

	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x) == ""

	case sql.NullString:
		return !x.Valid || strings.TrimSpace(x.String) == ""
	case *sql.NullString:
		return x == nil || !x.Valid || strings.TrimSpace(x.String) == ""

	case sql.NullInt64:
		return !x.Valid
	case *sql.NullInt64:
		return x == nil || !x.Valid

	case sql.NullInt32:
		return !x.Valid
	case *sql.NullInt32:
		return x == nil || !x.Valid

	case sql.NullInt16:
		return !x.Valid
	case *sql.NullInt16:
		return x == nil || !x.Valid

	case sql.NullByte:
		return !x.Valid
	case *sql.NullByte:
		return x == nil || !x.Valid

	case sql.NullFloat64:
		return !x.Valid
	case *sql.NullFloat64:
		return x == nil || !x.Valid

	case sql.NullBool:
		return !x.Valid
	case *sql.NullBool:
		return x == nil || !x.Valid

	case sql.NullTime:
		return !x.Valid
	case *sql.NullTime:
		return x == nil || !x.Valid

	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
			return rv.IsNil()
		}
		switch rv.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array:
			return rv.Len() == 0
		case reflect.String:
			return strings.TrimSpace(rv.String()) == ""
		default:
			return false
		}
	}
}

func isValueNull(v any) bool {
	if v == nil {
		return true
	}

	switch x := v.(type) {
	case sql.NullString:
		return !x.Valid
	case *sql.NullString:
		return x == nil || !x.Valid

	case sql.NullInt64:
		return !x.Valid
	case *sql.NullInt64:
		return x == nil || !x.Valid

	case sql.NullInt32:
		return !x.Valid
	case *sql.NullInt32:
		return x == nil || !x.Valid

	case sql.NullInt16:
		return !x.Valid
	case *sql.NullInt16:
		return x == nil || !x.Valid

	case sql.NullByte:
		return !x.Valid
	case *sql.NullByte:
		return x == nil || !x.Valid

	case sql.NullFloat64:
		return !x.Valid
	case *sql.NullFloat64:
		return x == nil || !x.Valid

	case sql.NullBool:
		return !x.Valid
	case *sql.NullBool:
		return x == nil || !x.Valid

	case sql.NullTime:
		return !x.Valid
	case *sql.NullTime:
		return x == nil || !x.Valid

	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
			return rv.IsNil()
		default:
			return false
		}
	}
}

func makeFoundExpectation[T any]() *expect.Expectation[T] {
	return expect.New(
		"Expect: Found",
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    "Expected query to return at least one result, but got sql.ErrNoRows",
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Found] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: Found]")
			}
		},
	)
}

func makeNotFoundExpectation[T any]() *expect.Expectation[T] {
	return expect.New(
		"Expect: Not Found",
		func(err error, result T) expect.CheckResult {
			if errors.Is(err, sql.ErrNoRows) {
				return expect.CheckResult{Ok: true}
			}
			if err != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected sql.ErrNoRows, but got different error: %v", err),
				}
			}
			return expect.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Expected sql.ErrNoRows, but query returned results",
			}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Not Found] %s", checkRes.Reason)
			} else {
				a.True(true, "[Expect: Not Found]")
			}
		},
	)
}

func makeColumnEqualsExpectation[T any](columnName string, expectedValue any) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' = %v", columnName, expectedValue),
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			equal, retryable, reason := equalsLoose(expectedValue, actualValue)
			if !equal {
				return expect.CheckResult{
					Ok:        false,
					Retryable: retryable,
					Reason:    fmt.Sprintf("Column '%s': %s", columnName, reason),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' = %v] %s", columnName, expectedValue, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Column '%s' = %v]", columnName, expectedValue)
			}
		},
	)
}

func makeColumnNotEmptyExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' not empty", columnName),
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			isEmpty := isEmptyValue(actualValue)
			if isEmpty {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to not be empty, but it is", columnName),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' not empty] %s", columnName, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Column '%s' not empty]", columnName)
			}
		},
	)
}

func makeColumnIsNullExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' IS NULL", columnName),
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			isNull := isValueNull(actualValue)
			if !isNull {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be NULL, but it has a value", columnName),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' IS NULL] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			isNull := isValueNull(actualValue)
			a.Equal(true, isNull, "[Expect: Column '%s' IS NULL]", columnName)
		},
	)
}

func makeColumnIsNotNullExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' IS NOT NULL", columnName),
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			isNull := isValueNull(actualValue)
			if isNull {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be NOT NULL, but it is NULL", columnName),
				}
			}
			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' IS NOT NULL] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			isNull := isValueNull(actualValue)
			a.Equal(false, isNull, "[Expect: Column '%s' IS NOT NULL]", columnName)
		},
	)
}

func makeColumnTrueExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' = true", columnName),
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			if isValueNull(actualValue) {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
				}
			}

			b, ok := asBool(actualValue)
			if !ok {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Column '%s' is not a boolean/numeric(0/1) type", columnName),
				}
			}

			if !b {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be true, but got false", columnName),
				}
			}

			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' = true] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			b, _ := asBool(actualValue)
			a.Equal(true, b, "[Expect: Column '%s' = true]", columnName)
		},
	)
}

func makeColumnFalseExpectation[T any](columnName string) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' = false", columnName),
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			if isValueNull(actualValue) {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
				}
			}

			b, ok := asBool(actualValue)
			if !ok {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Column '%s' is not a boolean/numeric(0/1) type", columnName),
				}
			}

			if b {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected column '%s' to be false, but got true", columnName),
				}
			}

			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' = false] %s", columnName, checkRes.Reason)
				return
			}

			actualValue, _ := getFieldValueByColumnName(result, columnName)
			b, _ := asBool(actualValue)
			a.Equal(false, b, "[Expect: Column '%s' = false]", columnName)
		},
	)
}

func makeColumnNotEqualsExpectation[T any](columnName string, notExpectedValue any) *expect.Expectation[T] {
	return expect.New(
		fmt.Sprintf("Expect: Column '%s' != %v", columnName, notExpectedValue),
		func(err error, result T) expect.CheckResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return expect.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return expect.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			equal, _, _ := equalsLoose(notExpectedValue, actualValue)

			if equal {
				return expect.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Column '%s' equals %v, but expected NOT to equal", columnName, actualValue),
				}
			}

			return expect.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, result T, checkRes expect.CheckResult) {
			a := extension.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect: Column '%s' != %v] %s", columnName, notExpectedValue, checkRes.Reason)
			} else {
				a.True(true, "[Expect: Column '%s' != %v]", columnName, notExpectedValue)
			}
		},
	)
}
