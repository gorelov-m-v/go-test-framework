package expect

import (
	"fmt"

	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/typeconv"
)

type ValueCheck func(value any, columnName string) polling.CheckResult

type JSONCheck func(res gjson.Result, path string) polling.CheckResult

func CheckEquals(expected any, equalsFunc func(expected, actual any) (bool, bool, string)) ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		equal, retryable, reason := equalsFunc(expected, value)
		if !equal {
			return polling.CheckResult{
				Ok:        false,
				Retryable: retryable,
				Reason:    fmt.Sprintf("Column '%s': %s", columnName, reason),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func CheckNotEquals(notExpected any, equalsFunc func(expected, actual any) (bool, bool, string)) ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		equal, _, _ := equalsFunc(notExpected, value)
		if equal {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Column '%s' equals %v, but expected NOT to equal", columnName, value),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func CheckNotEmpty() ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		if typeconv.IsEmpty(value) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Expected column '%s' to not be empty, but it is", columnName),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func CheckEmpty() ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		if !typeconv.IsEmpty(value) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Expected column '%s' to be empty, but got: %v", columnName, value),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func CheckIsNull() ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		if !typeconv.IsNull(value) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Expected column '%s' to be NULL, but it has a value", columnName),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func CheckIsNotNull() ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		if typeconv.IsNull(value) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Expected column '%s' to be NOT NULL, but it is NULL", columnName),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func CheckTrue() ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		if typeconv.IsNull(value) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
			}
		}
		b, ok := typeconv.ToBool(value)
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
	}
}

func CheckFalse() ValueCheck {
	return func(value any, columnName string) polling.CheckResult {
		if typeconv.IsNull(value) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Column '%s' is NULL yet", columnName),
			}
		}
		b, ok := typeconv.ToBool(value)
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
	}
}

func JSONCheckEquals(expected any) JSONCheck {
	return func(res gjson.Result, path string) polling.CheckResult {
		ok, msg := jsonutil.Compare(res, expected)
		if !ok {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    msg,
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func JSONCheckNotEmpty() JSONCheck {
	return func(res gjson.Result, path string) polling.CheckResult {
		if jsonutil.IsEmpty(res) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("JSON field '%s' is empty", path),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func JSONCheckEmpty() JSONCheck {
	return func(res gjson.Result, path string) polling.CheckResult {
		if !jsonutil.IsEmpty(res) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Field '%s' is not empty, got: %s", path, res.String()),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func JSONCheckIsNull() JSONCheck {
	return func(res gjson.Result, path string) polling.CheckResult {
		if res.Type != gjson.Null {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Expected null, got %s: %s", jsonutil.TypeToString(res.Type), jsonutil.DebugValue(res)),
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func JSONCheckIsNotNull() JSONCheck {
	return func(res gjson.Result, path string) polling.CheckResult {
		if res.Type == gjson.Null {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Expected non-null value, got null",
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func JSONCheckTrue() JSONCheck {
	return func(res gjson.Result, path string) polling.CheckResult {
		if res.Type != gjson.True && res.Type != gjson.False {
			return polling.CheckResult{
				Ok:        false,
				Retryable: false,
				Reason:    fmt.Sprintf("Expected boolean, got %s: %s", jsonutil.TypeToString(res.Type), jsonutil.DebugValue(res)),
			}
		}
		if res.Type != gjson.True {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Expected true, got false",
			}
		}
		return polling.CheckResult{Ok: true}
	}
}

func JSONCheckFalse() JSONCheck {
	return func(res gjson.Result, path string) polling.CheckResult {
		if res.Type != gjson.True && res.Type != gjson.False {
			return polling.CheckResult{
				Ok:        false,
				Retryable: false,
				Reason:    fmt.Sprintf("Expected boolean, got %s: %s", jsonutil.TypeToString(res.Type), jsonutil.DebugValue(res)),
			}
		}
		if res.Type != gjson.False {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    "Expected false, got true",
			}
		}
		return polling.CheckResult{Ok: true}
	}
}
