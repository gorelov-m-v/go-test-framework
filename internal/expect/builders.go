package expect

import (
	"errors"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

func checkColumnError[T any](err error, errNoRows error, columnName string) (polling.CheckResult, bool) {
	if err != nil {
		if errNoRows != nil && errors.Is(err, errNoRows) {
			return polling.CheckResult{
				Ok:        false,
				Retryable: true,
				Reason:    fmt.Sprintf("Cannot check column '%s': query returned no rows", columnName),
			}, false
		}
		return polling.CheckResult{
			Ok:        false,
			Retryable: false,
			Reason:    fmt.Sprintf("Cannot check column '%s': query failed", columnName),
		}, false
	}
	return polling.CheckResult{}, true
}

func getColumnValue[T any](result T, columnName string, getValue func(T, string) (any, error)) (any, polling.CheckResult, bool) {
	actualValue, getErr := getValue(result, columnName)
	if getErr != nil {
		return nil, polling.CheckResult{
			Ok:        false,
			Retryable: false,
			Reason:    fmt.Sprintf("Failed to get field value: %v", getErr),
		}, false
	}
	return actualValue, polling.CheckResult{}, true
}

type ColumnExpectationConfig[T any] struct {
	ColumnName string
	ExpectName string
	GetValue   func(result T, columnName string) (any, error)
	ErrNoRows  error
	Check      ValueCheck
	Reporter   ReportFunc[T]
}

func BuildColumnExpectation[T any](cfg ColumnExpectationConfig[T]) *Expectation[T] {
	reporter := cfg.Reporter
	if reporter == nil {
		reporter = StandardReport[T](cfg.ExpectName)
	}

	return New(
		cfg.ExpectName,
		func(err error, result T) polling.CheckResult {
			if res, ok := checkColumnError[T](err, cfg.ErrNoRows, cfg.ColumnName); !ok {
				return res
			}

			actualValue, res, ok := getColumnValue(result, cfg.ColumnName, cfg.GetValue)
			if !ok {
				return res
			}

			return cfg.Check(actualValue, cfg.ColumnName)
		},
		reporter,
	)
}

type ColumnBoolExpectationConfig[T any] struct {
	ColumnName   string
	ExpectName   string
	GetValue     func(result T, columnName string) (any, error)
	ErrNoRows    error
	Check        ValueCheck
	ExpectedBool bool
	ToBoolFunc   func(any) (bool, bool)
}

func BuildColumnBoolExpectation[T any](cfg ColumnBoolExpectationConfig[T]) *Expectation[T] {
	return New(
		cfg.ExpectName,
		func(err error, result T) polling.CheckResult {
			if res, ok := checkColumnError[T](err, cfg.ErrNoRows, cfg.ColumnName); !ok {
				return res
			}

			actualValue, res, ok := getColumnValue(result, cfg.ColumnName, cfg.GetValue)
			if !ok {
				return res
			}

			return cfg.Check(actualValue, cfg.ColumnName)
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[%s] %s", cfg.ExpectName, checkRes.Reason)
				return
			}
			actualValue, _, _ := getColumnValue(result, cfg.ColumnName, cfg.GetValue)
			b, _ := cfg.ToBoolFunc(actualValue)
			a.Equal(cfg.ExpectedBool, b, "[%s]", cfg.ExpectName)
		},
	)
}

type ColumnNullExpectationConfig[T any] struct {
	ColumnName   string
	ExpectName   string
	GetValue     func(result T, columnName string) (any, error)
	ErrNoRows    error
	Check        ValueCheck
	ExpectedNull bool
	IsNullFunc   func(any) bool
}

func BuildColumnNullExpectation[T any](cfg ColumnNullExpectationConfig[T]) *Expectation[T] {
	return New(
		cfg.ExpectName,
		func(err error, result T) polling.CheckResult {
			if res, ok := checkColumnError[T](err, cfg.ErrNoRows, cfg.ColumnName); !ok {
				return res
			}

			actualValue, res, ok := getColumnValue(result, cfg.ColumnName, cfg.GetValue)
			if !ok {
				return res
			}

			return cfg.Check(actualValue, cfg.ColumnName)
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[%s] %s", cfg.ExpectName, checkRes.Reason)
				return
			}
			actualValue, _, _ := getColumnValue(result, cfg.ColumnName, cfg.GetValue)
			isNull := cfg.IsNullFunc(actualValue)
			a.Equal(cfg.ExpectedNull, isNull, "[%s]", cfg.ExpectName)
		},
	)
}

type ColumnEmptyExpectationConfig[T any] struct {
	ColumnName    string
	ExpectName    string
	GetValue      func(result T, columnName string) (any, error)
	ErrNoRows     error
	Check         ValueCheck
	ExpectedEmpty bool
	IsEmptyFunc   func(any) bool
}

func BuildColumnEmptyExpectation[T any](cfg ColumnEmptyExpectationConfig[T]) *Expectation[T] {
	return New(
		cfg.ExpectName,
		func(err error, result T) polling.CheckResult {
			if res, ok := checkColumnError[T](err, cfg.ErrNoRows, cfg.ColumnName); !ok {
				return res
			}

			actualValue, res, ok := getColumnValue(result, cfg.ColumnName, cfg.GetValue)
			if !ok {
				return res
			}

			return cfg.Check(actualValue, cfg.ColumnName)
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[%s] %s", cfg.ExpectName, checkRes.Reason)
				return
			}
			actualValue, _, _ := getColumnValue(result, cfg.ColumnName, cfg.GetValue)
			isEmpty := cfg.IsEmptyFunc(actualValue)
			a.Equal(cfg.ExpectedEmpty, isEmpty, "[%s]", cfg.ExpectName)
		},
	)
}

type JSONFieldExpectationConfig[T any] struct {
	Path       string
	ExpectName string
	GetJSON    func(result T) ([]byte, error)
	PreCheck   func(err error, result T) (polling.CheckResult, bool)
	Check      JSONCheck
	Reporter   ReportFunc[T]
}

func checkJSONField[T any](cfg JSONFieldExpectationConfig[T], err error, result T) (gjson.Result, polling.CheckResult, bool) {
	if cfg.PreCheck != nil {
		if res, ok := cfg.PreCheck(err, result); !ok {
			return gjson.Result{}, res, false
		}
	}

	jsonBytes, jsonErr := cfg.GetJSON(result)
	if jsonErr != nil {
		return gjson.Result{}, polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
		}, false
	}

	if !gjson.ValidBytes(jsonBytes) {
		return gjson.Result{}, polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    "Invalid JSON",
		}, false
	}

	jsonRes := gjson.GetBytes(jsonBytes, cfg.Path)
	if !jsonRes.Exists() {
		return gjson.Result{}, polling.CheckResult{
			Ok:        false,
			Retryable: true,
			Reason:    fmt.Sprintf("JSON field '%s' does not exist", cfg.Path),
		}, false
	}

	return jsonRes, polling.CheckResult{}, true
}

func BuildJSONFieldExpectation[T any](cfg JSONFieldExpectationConfig[T]) *Expectation[T] {
	reporter := cfg.Reporter
	if reporter == nil {
		reporter = StandardReport[T](cfg.ExpectName)
	}

	return New(
		cfg.ExpectName,
		func(err error, result T) polling.CheckResult {
			jsonRes, res, ok := checkJSONField(cfg, err, result)
			if !ok {
				return res
			}
			return cfg.Check(jsonRes, cfg.Path)
		},
		reporter,
	)
}

type JSONFieldNullExpectationConfig[T any] struct {
	Path         string
	ExpectName   string
	GetJSON      func(result T) ([]byte, error)
	PreCheck     func(err error, result T) (polling.CheckResult, bool)
	ExpectedNull bool
}

func BuildJSONFieldNullExpectation[T any](cfg JSONFieldNullExpectationConfig[T]) *Expectation[T] {
	var check JSONCheck
	if cfg.ExpectedNull {
		check = JSONCheckIsNull()
	} else {
		check = JSONCheckIsNotNull()
	}

	fieldCfg := JSONFieldExpectationConfig[T]{
		Path:       cfg.Path,
		ExpectName: cfg.ExpectName,
		GetJSON:    cfg.GetJSON,
		PreCheck:   cfg.PreCheck,
		Check:      check,
	}

	return BuildJSONFieldExpectation(fieldCfg)
}

type BytesJSONFieldExpectationConfig struct {
	Path       string
	ExpectName string
	Check      JSONCheck
}

func BuildBytesJSONFieldExpectation(cfg BytesJSONFieldExpectationConfig) *Expectation[[]byte] {
	return New(
		cfg.ExpectName,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}

			result := gjson.GetBytes(msgBytes, cfg.Path)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' not found", cfg.Path),
				}
			}

			return cfg.Check(result, cfg.Path)
		},
		StandardReport[[]byte](cfg.ExpectName),
	)
}

type BytesJSONFieldExistsCheckConfig struct {
	Path          string
	ExpectName    string
	RequireExists bool
	Check         JSONCheck
}

type BytesJSONFieldNullCheckConfig struct {
	Path         string
	ExpectName   string
	ExpectedNull bool
}

func BuildBytesJSONFieldWithExistsCheck(cfg BytesJSONFieldExistsCheckConfig) *Expectation[[]byte] {
	return New(
		cfg.ExpectName,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}

			result := gjson.GetBytes(msgBytes, cfg.Path)

			if !result.Exists() {
				if cfg.RequireExists {
					return polling.CheckResult{
						Ok:        false,
						Retryable: false,
						Reason:    fmt.Sprintf("Field '%s' not found", cfg.Path),
					}
				}
				return polling.CheckResult{Ok: true}
			}

			return cfg.Check(result, cfg.Path)
		},
		StandardReport[[]byte](cfg.ExpectName),
	)
}

func BuildBytesJSONFieldNullCheck(cfg BytesJSONFieldNullCheckConfig) *Expectation[[]byte] {
	return New(
		cfg.ExpectName,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}

			result := gjson.GetBytes(msgBytes, cfg.Path)

			if cfg.ExpectedNull {
				if result.Exists() && result.Type != gjson.Null {
					return polling.CheckResult{
						Ok:        false,
						Retryable: false,
						Reason:    fmt.Sprintf("Expected null, got %s", result.Type.String()),
					}
				}
				return polling.CheckResult{Ok: true}
			}

			if !result.Exists() || result.Type == gjson.Null {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' is null or does not exist", cfg.Path),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		StandardReport[[]byte](cfg.ExpectName),
	)
}

type ObjectCompareFunc func(jsonObj gjson.Result, expected any) (bool, string)

type FullObjectExpectationConfig[T any] struct {
	ExpectName string
	GetJSON    func(result T) ([]byte, error)
	PreCheck   func(err error, result T) (polling.CheckResult, bool)
	Expected   any
	Compare    ObjectCompareFunc
	Retryable  bool
}

func BuildFullObjectExpectation[T any](cfg FullObjectExpectationConfig[T]) *Expectation[T] {
	return New(
		cfg.ExpectName,
		func(err error, result T) polling.CheckResult {
			if cfg.PreCheck != nil {
				if res, ok := cfg.PreCheck(err, result); !ok {
					return res
				}
			}

			jsonBytes, jsonErr := cfg.GetJSON(result)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: cfg.Retryable,
					Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
				}
			}

			if !gjson.ValidBytes(jsonBytes) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: cfg.Retryable,
					Reason:    "Invalid JSON",
				}
			}

			jsonRes := gjson.ParseBytes(jsonBytes)
			ok, msg := cfg.Compare(jsonRes, cfg.Expected)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: cfg.Retryable,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		StandardReport[T](cfg.ExpectName),
	)
}

type BytesObjectExpectationConfig struct {
	ExpectName string
	Expected   any
	Compare    ObjectCompareFunc
}

func BuildBytesObjectExpectation(cfg BytesObjectExpectationConfig) *Expectation[[]byte] {
	return New(
		cfg.ExpectName,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}

			if !gjson.ValidBytes(msgBytes) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Invalid JSON message",
				}
			}

			jsonRes := gjson.ParseBytes(msgBytes)
			ok, msg := cfg.Compare(jsonRes, cfg.Expected)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		StandardReport[[]byte](cfg.ExpectName),
	)
}
