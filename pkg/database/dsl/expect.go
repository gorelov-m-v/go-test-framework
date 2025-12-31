package dsl

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

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
			a.NoError(err, "Expected query to execute successfully and return at least one result")
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

			if expectedBool, ok := expectedValue.(bool); ok {
				if actualBool, ok := asBool(actualValue); ok {
					a.Equal(expectedBool, actualBool, "Expected column '%s' = %v", columnName, expectedValue)
					return
				}
				a.True(false, "Column '%s' is not a boolean/numeric(0/1) type", columnName)
				return
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

			if isNilAny(actualValue) {
				a.False(true, "Expected column '%s' to not be empty", columnName)
				return
			}

			isEmpty := false

			if str, ok := actualValue.(string); ok {
				isEmpty = strings.TrimSpace(str) == ""
			} else if val := reflect.ValueOf(actualValue); val.Kind() == reflect.Ptr {
				isEmpty = val.IsNil()
			} else {
				switch v := actualValue.(type) {
				case sql.NullString:
					isEmpty = !v.Valid || (v.Valid && strings.TrimSpace(v.String) == "")
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
					isEmpty = v == nil || !v.Valid || (v.Valid && strings.TrimSpace(v.String) == "")
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
					val := reflect.ValueOf(actualValue)
					switch val.Kind() {
					case reflect.Slice, reflect.Map, reflect.Array:
						isEmpty = val.Len() == 0
					case reflect.String:
						isEmpty = strings.TrimSpace(val.String()) == ""
					default:
						isEmpty = false
					}
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

			if isNilAny(actualValue) {
				a.True(true, "Column '%s' is NULL (nil value)", columnName)
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

			if isNilAny(actualValue) {
				a.True(false, "Expected column '%s' to be NOT NULL", columnName)
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

			if b, ok := asBool(actualValue); ok {
				a.True(b, "Expected column '%s' to be true", columnName)
				return
			}
			a.True(false, "Column '%s' is not a boolean/numeric(0/1) type", columnName)
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

			if b, ok := asBool(actualValue); ok {
				a.False(b, "Expected column '%s' to be false", columnName)
				return
			}
			a.True(false, "Column '%s' is not a boolean/numeric(0/1) type", columnName)
		})
	})
	return q
}
