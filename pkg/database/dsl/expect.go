package dsl

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func ensureQuerySuccessSilent(a provider.Asserts, err error) bool {
	if err != nil {
		a.NoError(err, "Cannot check expectation when query failed")
		return false
	}
	return true
}

func ensureFetchResultSilent(a provider.Asserts, scannedResult any) bool {
	if scannedResult == nil {
		a.True(false, "This expectation can only be used with MustFetch()")
		return false
	}
	return true
}

func getColumnValueSilent(a provider.Asserts, scannedResult any, columnName string) (any, bool) {
	if !ensureFetchResultSilent(a, scannedResult) {
		return nil, false
	}

	actualValue, getErr := getFieldValueByColumnName(scannedResult, columnName)
	if getErr != nil {
		a.NoError(getErr, "Failed to get field value for assertion")
		return nil, false
	}
	return actualValue, true
}

func (q *Query[T]) ExpectFound() *Query[T] {
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep("Expect: Found", func(sCtx provider.StepCtx) {
			a := sCtx.Require()
			a.NoError(err, "Expected to find a result, but query failed")
		})
	})
	return q
}

func (q *Query[T]) ExpectNotFound() *Query[T] {
	q.expectsNotFound = true
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep("Expect: Not Found", func(sCtx provider.StepCtx) {
			a := sCtx.Require()
			a.ErrorIs(err, sql.ErrNoRows, "Expected sql.ErrNoRows")
		})
	})
	return q
}

func (q *Query[T]) ExpectColumnEquals(columnName string, expectedValue any) *Query[T] {
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep(fmt.Sprintf("Expect: Column '%s' = %v", columnName, expectedValue), func(sCtx provider.StepCtx) {
			a := sCtx.Require()

			if !ensureQuerySuccessSilent(a, err) {
				return
			}

			actualValue, ok := getColumnValueSilent(a, scannedResult, columnName)
			if !ok {
				return
			}

			// Частый кейс: булевое ожидаемое vs TINYINT(1) из БД
			if expectedBool, ok := expectedValue.(bool); ok {
				if actualInt64, ok := actualValue.(int64); ok {
					a.Equal(expectedBool, actualInt64 == 1, "Expected column '%s' = %v", columnName, expectedValue)
					return
				}
			}

			a.Equal(expectedValue, actualValue, "Expected column '%s' = %v", columnName, expectedValue)
		})
	})
	return q
}

func (q *Query[T]) ExpectColumnNotEmpty(columnName string) *Query[T] {
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep(fmt.Sprintf("Expect: Column '%s' not empty", columnName), func(sCtx provider.StepCtx) {
			a := sCtx.Require()

			if !ensureQuerySuccessSilent(a, err) {
				return
			}

			actualValue, ok := getColumnValueSilent(a, scannedResult, columnName)
			if !ok {
				return
			}

			isEmpty := false

			if str, ok := actualValue.(string); ok {
				isEmpty = strings.TrimSpace(str) == ""
			} else if val := reflect.ValueOf(actualValue); val.Kind() == reflect.Ptr {
				isEmpty = val.IsNil()
			} else {
				val := reflect.ValueOf(actualValue)
				switch v := actualValue.(type) {
				case sql.NullString:
					isEmpty = !v.Valid
				case sql.NullInt64:
					isEmpty = !v.Valid
				case sql.NullInt32:
					isEmpty = !v.Valid
				case sql.NullInt16:
					isEmpty = !v.Valid
				case sql.NullByte:
					isEmpty = !v.Valid
				case sql.NullFloat64:
					isEmpty = !v.Valid
				case sql.NullBool:
					isEmpty = !v.Valid
				case sql.NullTime:
					isEmpty = !v.Valid
				case *sql.NullString:
					isEmpty = v == nil || !v.Valid
				case *sql.NullInt64:
					isEmpty = v == nil || !v.Valid
				case *sql.NullInt32:
					isEmpty = v == nil || !v.Valid
				case *sql.NullInt16:
					isEmpty = v == nil || !v.Valid
				case *sql.NullByte:
					isEmpty = v == nil || !v.Valid
				case *sql.NullFloat64:
					isEmpty = v == nil || !v.Valid
				case *sql.NullBool:
					isEmpty = v == nil || !v.Valid
				case *sql.NullTime:
					isEmpty = v == nil || !v.Valid
				default:
					isEmpty = val.IsZero()
				}
			}

			a.False(isEmpty, "Expected column '%s' to not be empty", columnName)
		})
	})
	return q
}

func (q *Query[T]) ExpectColumnIsNull(columnName string) *Query[T] {
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep(fmt.Sprintf("Expect: Column '%s' IS NULL", columnName), func(sCtx provider.StepCtx) {
			a := sCtx.Require()

			if !ensureQuerySuccessSilent(a, err) {
				return
			}

			actualValue, ok := getColumnValueSilent(a, scannedResult, columnName)
			if !ok {
				return
			}

			val := reflect.ValueOf(actualValue)
			if val.Kind() == reflect.Ptr {
				a.True(val.IsNil(), "Expected pointer for column '%s' to be nil", columnName)
				return
			}

			switch v := actualValue.(type) {
			case sql.NullString:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case sql.NullInt64:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case sql.NullInt32:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case sql.NullInt16:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case sql.NullByte:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case sql.NullFloat64:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case sql.NullBool:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case sql.NullTime:
				a.False(v.Valid, "Expected column '%s' to be NULL (Valid=false)", columnName)
			case *sql.NullString:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			case *sql.NullInt64:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			case *sql.NullInt32:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			case *sql.NullInt16:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			case *sql.NullByte:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			case *sql.NullFloat64:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			case *sql.NullBool:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			case *sql.NullTime:
				a.True(v == nil || !v.Valid, "Expected column '%s' to be NULL", columnName)
			default:
				a.True(false, "Column '%s' type %T does not support NULL check", columnName, actualValue)
			}
		})
	})
	return q
}

func (q *Query[T]) ExpectColumnIsNotNull(columnName string) *Query[T] {
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep(fmt.Sprintf("Expect: Column '%s' IS NOT NULL", columnName), func(sCtx provider.StepCtx) {
			a := sCtx.Require()

			if !ensureQuerySuccessSilent(a, err) {
				return
			}

			actualValue, ok := getColumnValueSilent(a, scannedResult, columnName)
			if !ok {
				return
			}

			val := reflect.ValueOf(actualValue)
			if val.Kind() == reflect.Ptr {
				a.False(val.IsNil(), "Expected pointer for column '%s' to not be nil", columnName)
				return
			}

			switch v := actualValue.(type) {
			case sql.NullString:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case sql.NullInt64:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case sql.NullInt32:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case sql.NullInt16:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case sql.NullByte:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case sql.NullFloat64:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case sql.NullBool:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case sql.NullTime:
				a.True(v.Valid, "Expected column '%s' to be NOT NULL (Valid=true)", columnName)
			case *sql.NullString:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			case *sql.NullInt64:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			case *sql.NullInt32:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			case *sql.NullInt16:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			case *sql.NullByte:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			case *sql.NullFloat64:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			case *sql.NullBool:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			case *sql.NullTime:
				a.True(v != nil && v.Valid, "Expected column '%s' to be NOT NULL", columnName)
			default:
				// Для не-nullable типов считаем IS NOT NULL всегда true
				a.True(true, "Column '%s' is NOT NULL", columnName)
			}
		})
	})
	return q
}

func (q *Query[T]) ExpectColumnTrue(columnName string) *Query[T] {
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep(fmt.Sprintf("Expect: Column '%s' = true", columnName), func(sCtx provider.StepCtx) {
			a := sCtx.Require()

			if !ensureQuerySuccessSilent(a, err) {
				return
			}

			actualValue, ok := getColumnValueSilent(a, scannedResult, columnName)
			if !ok {
				return
			}

			if v, ok := actualValue.(int64); ok {
				a.Equal(int64(1), v, "Expected column '%s' to be true (1)", columnName)
				return
			}
			if v, ok := actualValue.(bool); ok {
				a.True(v, "Expected column '%s' to be true", columnName)
				return
			}

			a.True(false, "Column '%s' is not a boolean/int64 type", columnName)
		})
	})
	return q
}

func (q *Query[T]) ExpectColumnFalse(columnName string) *Query[T] {
	q.expectations = append(q.expectations, func(parent provider.StepCtx, err error, scannedResult any) {
		parent.WithNewStep(fmt.Sprintf("Expect: Column '%s' = false", columnName), func(sCtx provider.StepCtx) {
			a := sCtx.Require()

			if !ensureQuerySuccessSilent(a, err) {
				return
			}

			actualValue, ok := getColumnValueSilent(a, scannedResult, columnName)
			if !ok {
				return
			}

			if v, ok := actualValue.(int64); ok {
				a.Equal(int64(0), v, "Expected column '%s' to be false (0)", columnName)
				return
			}
			if v, ok := actualValue.(bool); ok {
				a.False(v, "Expected column '%s' to be false", columnName)
				return
			}

			a.True(false, "Column '%s' is not a boolean/int64 type", columnName)
		})
	})
	return q
}
