package dsl

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type assertMode int

const (
	requireMode assertMode = iota
	assertModeValue
)

type checkResult struct {
	ok        bool
	retryable bool
	reason    string
}

type expectation struct {
	name   string
	check  func(err error, result any) checkResult
	report func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult)
}

func (q *Query[T]) ExpectFound() *Query[T] {
	q.expectations = append(q.expectations, makeFoundExpectation())
	return q
}

func (q *Query[T]) ExpectNotFound() *Query[T] {
	q.expectsNotFound = true
	q.expectations = []*expectation{}
	q.expectations = append(q.expectations, makeNotFoundExpectation())
	return q
}

func (q *Query[T]) ExpectColumnEquals(columnName string, expectedValue any) *Query[T] {
	if q.expectsNotFound {
		panic("ExpectColumnEquals cannot be used with ExpectNotFound()")
	}
	q.expectations = append(q.expectations, makeColumnEqualsExpectation(columnName, expectedValue))
	return q
}

func (q *Query[T]) ExpectColumnNotEmpty(columnName string) *Query[T] {
	if q.expectsNotFound {
		panic("ExpectColumnNotEmpty cannot be used with ExpectNotFound()")
	}
	q.expectations = append(q.expectations, makeColumnNotEmptyExpectation(columnName))
	return q
}

func (q *Query[T]) ExpectColumnIsNull(columnName string) *Query[T] {
	if q.expectsNotFound {
		panic("ExpectColumnIsNull cannot be used with ExpectNotFound()")
	}
	q.expectations = append(q.expectations, makeColumnIsNullExpectation(columnName))
	return q
}

func (q *Query[T]) ExpectColumnIsNotNull(columnName string) *Query[T] {
	if q.expectsNotFound {
		panic("ExpectColumnIsNotNull cannot be used with ExpectNotFound()")
	}
	q.expectations = append(q.expectations, makeColumnIsNotNullExpectation(columnName))
	return q
}

func (q *Query[T]) ExpectColumnTrue(columnName string) *Query[T] {
	if q.expectsNotFound {
		panic("ExpectColumnTrue cannot be used with ExpectNotFound()")
	}
	q.expectations = append(q.expectations, makeColumnTrueExpectation(columnName))
	return q
}

func (q *Query[T]) ExpectColumnFalse(columnName string) *Query[T] {
	if q.expectsNotFound {
		panic("ExpectColumnFalse cannot be used with ExpectNotFound()")
	}
	q.expectations = append(q.expectations, makeColumnFalseExpectation(columnName))
	return q
}

func newExpectation(
	name string,
	checkFn func(err error, result any) checkResult,
	reportFn func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult),
) *expectation {
	return &expectation{
		name:   name,
		check:  checkFn,
		report: reportFn,
	}
}

func reportWithMode(stepCtx provider.StepCtx, mode assertMode) provider.Asserts {
	if mode == requireMode {
		return stepCtx.Require()
	}
	return stepCtx.Assert()
}

func isNilAny(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.Chan:
		return rv.IsNil()
	default:
		return false
	}
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

func isEmptyValue(actualValue any) bool {
	if isNilAny(actualValue) {
		return true
	}

	if str, ok := actualValue.(string); ok {
		return strings.TrimSpace(str) == ""
	}

	val := reflect.ValueOf(actualValue)
	if val.Kind() == reflect.Ptr {
		return val.IsNil()
	}

	switch v := actualValue.(type) {
	case sql.NullString:
		return !v.Valid || (v.Valid && strings.TrimSpace(v.String) == "")
	case sql.NullInt64:
		return !v.Valid
	case sql.NullInt32:
		return !v.Valid
	case sql.NullInt16:
		return !v.Valid
	case sql.NullByte:
		return !v.Valid
	case sql.NullFloat64:
		return !v.Valid
	case sql.NullBool:
		return !v.Valid
	case sql.NullTime:
		return !v.Valid
	case *sql.NullString:
		return v == nil || !v.Valid || (v.Valid && strings.TrimSpace(v.String) == "")
	case *sql.NullInt64:
		return v == nil || !v.Valid
	case *sql.NullInt32:
		return v == nil || !v.Valid
	case *sql.NullInt16:
		return v == nil || !v.Valid
	case *sql.NullByte:
		return v == nil || !v.Valid
	case *sql.NullFloat64:
		return v == nil || !v.Valid
	case *sql.NullBool:
		return v == nil || !v.Valid
	case *sql.NullTime:
		return v == nil || !v.Valid
	default:
		switch val.Kind() {
		case reflect.Slice, reflect.Map, reflect.Array:
			return val.Len() == 0
		case reflect.String:
			return strings.TrimSpace(val.String()) == ""
		default:
			return false
		}
	}
}

func isValueNull(actualValue any) bool {
	if isNilAny(actualValue) {
		return true
	}

	val := reflect.ValueOf(actualValue)
	if val.Kind() == reflect.Ptr {
		return val.IsNil()
	}

	switch v := actualValue.(type) {
	case sql.NullString, sql.NullInt64, sql.NullInt32, sql.NullInt16, sql.NullByte, sql.NullFloat64, sql.NullBool, sql.NullTime:
		return !reflect.ValueOf(v).FieldByName("Valid").Bool()
	case *sql.NullString, *sql.NullInt64, *sql.NullInt32, *sql.NullInt16, *sql.NullByte, *sql.NullFloat64, *sql.NullBool, *sql.NullTime:
		if val.IsNil() {
			return true
		}
		return !val.Elem().FieldByName("Valid").Bool()
	default:
		return false
	}
}

func reportNullCheck(a provider.Asserts, actualValue any, columnName string, expectNull bool) {
	if isNilAny(actualValue) {
		if expectNull {
			return
		}
		a.True(false, "Expected column '%s' to be NOT NULL", columnName)
		return
	}

	val := reflect.ValueOf(actualValue)
	if val.Kind() == reflect.Ptr {
		if expectNull {
			a.True(val.IsNil(), "Expected pointer for column '%s' to be nil", columnName)
		} else {
			a.False(val.IsNil(), "Expected pointer for column '%s' to not be nil", columnName)
		}
		return
	}

	switch v := actualValue.(type) {
	case sql.NullString:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case sql.NullInt64:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case sql.NullInt32:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case sql.NullInt16:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case sql.NullByte:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case sql.NullFloat64:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case sql.NullBool:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case sql.NullTime:
		a.Equal(expectNull, !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullString:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullInt64:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullInt32:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullInt16:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullByte:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullFloat64:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullBool:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	case *sql.NullTime:
		a.Equal(expectNull, v == nil || !v.Valid, "Expected column '%s' NULL=%v", columnName, expectNull)
	default:
		if !expectNull {
			return
		}
		a.True(false, "Column '%s' type %T does not support NULL check", columnName, actualValue)
	}
}

func makeFoundExpectation() *expectation {
	return newExpectation(
		"Expect: Found",
		func(err error, result any) checkResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return checkResult{
						ok:        false,
						retryable: true,
						reason:    "Expected query to return at least one result, but got sql.ErrNoRows",
					}
				}
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Query failed: %v", err),
				}
			}
			return checkResult{ok: true}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep("Expect: Found", func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				a.NoError(err, checkRes.reason)
			})
		},
	)
}

func makeNotFoundExpectation() *expectation {
	return newExpectation(
		"Expect: Not Found",
		func(err error, result any) checkResult {
			if errors.Is(err, sql.ErrNoRows) {
				return checkResult{ok: true}
			}
			if err != nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Expected sql.ErrNoRows, but got different error: %v", err),
				}
			}
			return checkResult{
				ok:        false,
				retryable: true,
				reason:    "Expected sql.ErrNoRows, but query returned results",
			}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep("Expect: Not Found", func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				if !checkRes.ok {
					a.True(false, checkRes.reason)
				} else {
					a.ErrorIs(err, sql.ErrNoRows, "Expected sql.ErrNoRows")
				}
			})
		},
	)
}

func makeColumnEqualsExpectation(columnName string, expectedValue any) *expectation {
	return newExpectation(
		fmt.Sprintf("Expect: Column '%s' = %v", columnName, expectedValue),
		func(err error, result any) checkResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return checkResult{
						ok:        false,
						retryable: true,
						reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			if result == nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    "This expectation can only be used with MustFetch()",
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			equal, retryable, reason := equalsLoose(expectedValue, actualValue)
			if !equal {
				return checkResult{
					ok:        false,
					retryable: retryable,
					reason:    fmt.Sprintf("Column '%s': %s", columnName, reason),
				}
			}
			return checkResult{ok: true}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep(fmt.Sprintf("Expect: Column '%s' = %v", columnName, expectedValue), func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				if !checkRes.ok {
					a.True(false, checkRes.reason)
					return
				}

				actualValue, _ := getFieldValueByColumnName(result, columnName)
				if expectedBool, ok := expectedValue.(bool); ok {
					actualBool, _ := asBool(actualValue)
					a.Equal(expectedBool, actualBool, "Expected column '%s' = %v", columnName, expectedValue)
				} else {
					a.Equal(expectedValue, actualValue, "Expected column '%s' = %v", columnName, expectedValue)
				}
			})
		},
	)
}

func makeColumnNotEmptyExpectation(columnName string) *expectation {
	return newExpectation(
		fmt.Sprintf("Expect: Column '%s' not empty", columnName),
		func(err error, result any) checkResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return checkResult{
						ok:        false,
						retryable: true,
						reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			if result == nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    "This expectation can only be used with MustFetch()",
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			isEmpty := isEmptyValue(actualValue)
			if isEmpty {
				return checkResult{
					ok:        false,
					retryable: true,
					reason:    fmt.Sprintf("Expected column '%s' to not be empty, but it is", columnName),
				}
			}
			return checkResult{ok: true}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep(fmt.Sprintf("Expect: Column '%s' not empty", columnName), func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				if !checkRes.ok {
					a.True(false, checkRes.reason)
					return
				}

				actualValue, _ := getFieldValueByColumnName(result, columnName)
				isEmpty := isEmptyValue(actualValue)
				a.False(isEmpty, "Expected column '%s' to not be empty", columnName)
			})
		},
	)
}

func makeColumnIsNullExpectation(columnName string) *expectation {
	return newExpectation(
		fmt.Sprintf("Expect: Column '%s' IS NULL", columnName),
		func(err error, result any) checkResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return checkResult{
						ok:        false,
						retryable: true,
						reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			if result == nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    "This expectation can only be used with MustFetch()",
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			isNull := isValueNull(actualValue)
			if !isNull {
				return checkResult{
					ok:        false,
					retryable: true,
					reason:    fmt.Sprintf("Expected column '%s' to be NULL, but it has a value", columnName),
				}
			}
			return checkResult{ok: true}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep(fmt.Sprintf("Expect: Column '%s' IS NULL", columnName), func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				if !checkRes.ok {
					a.True(false, checkRes.reason)
					return
				}

				actualValue, _ := getFieldValueByColumnName(result, columnName)
				reportNullCheck(a, actualValue, columnName, true)
			})
		},
	)
}

func makeColumnIsNotNullExpectation(columnName string) *expectation {
	return newExpectation(
		fmt.Sprintf("Expect: Column '%s' IS NOT NULL", columnName),
		func(err error, result any) checkResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return checkResult{
						ok:        false,
						retryable: true,
						reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			if result == nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    "This expectation can only be used with MustFetch()",
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			isNull := isValueNull(actualValue)
			if isNull {
				return checkResult{
					ok:        false,
					retryable: true,
					reason:    fmt.Sprintf("Expected column '%s' to be NOT NULL, but it is NULL", columnName),
				}
			}
			return checkResult{ok: true}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep(fmt.Sprintf("Expect: Column '%s' IS NOT NULL", columnName), func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				if !checkRes.ok {
					a.True(false, checkRes.reason)
					return
				}

				actualValue, _ := getFieldValueByColumnName(result, columnName)
				reportNullCheck(a, actualValue, columnName, false)
			})
		},
	)
}

func makeColumnTrueExpectation(columnName string) *expectation {
	return newExpectation(
		fmt.Sprintf("Expect: Column '%s' = true", columnName),
		func(err error, result any) checkResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return checkResult{
						ok:        false,
						retryable: true,
						reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			if result == nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    "This expectation can only be used with MustFetch()",
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			if isValueNull(actualValue) {
				return checkResult{
					ok:        false,
					retryable: true,
					reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
				}
			}

			b, ok := asBool(actualValue)
			if !ok {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Column '%s' is not a boolean/numeric(0/1) type", columnName),
				}
			}

			if !b {
				return checkResult{
					ok:        false,
					retryable: true,
					reason:    fmt.Sprintf("Expected column '%s' to be true, but got false", columnName),
				}
			}

			return checkResult{ok: true}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep(fmt.Sprintf("Expect: Column '%s' = true", columnName), func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				if !checkRes.ok {
					a.True(false, checkRes.reason)
					return
				}

				actualValue, _ := getFieldValueByColumnName(result, columnName)
				b, _ := asBool(actualValue)
				a.True(b, "Expected column '%s' to be true", columnName)
			})
		},
	)
}

func makeColumnFalseExpectation(columnName string) *expectation {
	return newExpectation(
		fmt.Sprintf("Expect: Column '%s' = false", columnName),
		func(err error, result any) checkResult {
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return checkResult{
						ok:        false,
						retryable: true,
						reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
					}
				}
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
				}
			}

			if result == nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    "This expectation can only be used with MustFetch()",
				}
			}

			actualValue, getErr := getFieldValueByColumnName(result, columnName)
			if getErr != nil {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
				}
			}

			if isValueNull(actualValue) {
				return checkResult{
					ok:        false,
					retryable: true,
					reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
				}
			}

			b, ok := asBool(actualValue)
			if !ok {
				return checkResult{
					ok:        false,
					retryable: false,
					reason:    fmt.Sprintf("Column '%s' is not a boolean/numeric(0/1) type", columnName),
				}
			}

			if b {
				return checkResult{
					ok:        false,
					retryable: true,
					reason:    fmt.Sprintf("Expected column '%s' to be false, but got true", columnName),
				}
			}

			return checkResult{ok: true}
		},
		func(stepCtx provider.StepCtx, mode assertMode, err error, result any, checkRes checkResult) {
			stepCtx.WithNewStep(fmt.Sprintf("Expect: Column '%s' = false", columnName), func(sCtx provider.StepCtx) {
				a := reportWithMode(sCtx, mode)
				if !checkRes.ok {
					a.True(false, checkRes.reason)
					return
				}

				actualValue, _ := getFieldValueByColumnName(result, columnName)
				b, _ := asBool(actualValue)
				a.False(b, "Expected column '%s' to be false", columnName)
			})
		},
	)
}
