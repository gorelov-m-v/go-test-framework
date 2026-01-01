package dsl

import (
	"database/sql"
	"reflect"

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
		// This is a programming error, so we panic here
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
