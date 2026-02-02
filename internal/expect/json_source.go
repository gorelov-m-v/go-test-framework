package expect

import (
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type JSONExpectationSource[T any] struct {
	GetJSON          func(result T) ([]byte, error)
	PreCheck         func(err error, result T) (polling.CheckResult, bool)
	PreCheckWithBody func(err error, result T) (polling.CheckResult, bool)
}

func (s *JSONExpectationSource[T]) withPathValidation(path string) func(error, T) (polling.CheckResult, bool) {
	return func(err error, result T) (polling.CheckResult, bool) {
		if pathErr := ValidateJSONPath(path); pathErr != nil {
			return polling.CheckResult{
				Ok:        false,
				Retryable: false,
				Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
			}, false
		}
		return s.PreCheckWithBody(err, result)
	}
}

func (s *JSONExpectationSource[T]) FieldEquals(path string, expected any) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' == %v", path, expected)
	return New(
		name,
		func(err error, result T) polling.CheckResult {
			if pathErr := ValidateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := s.PreCheckWithBody(err, result); !ok {
				return res
			}
			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
				}
			}
			jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Path '%s' does not exist in response yet", path),
				}
			}
			ok, msg := jsonutil.Compare(jsonRes, expected)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect JSON field '%s' == %v] %s", path, expected, checkRes.Reason)
				return
			}

			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr == nil && len(jsonBytes) > 0 {
				res, parseErr := jsonutil.GetField(jsonBytes, path)
				if parseErr == nil && res.Exists() {
					actualValue := jsonutil.DebugValue(res)
					a.True(true, "[Expect JSON field '%s' == %v] actual: %s", path, expected, actualValue)
					return
				}
			}
			a.True(true, "[Expect JSON field '%s' == %v]", path, expected)
		},
	)
}

func (s *JSONExpectationSource[T]) FieldNotEmpty(path string) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' not empty", path)
	return BuildJSONFieldExpectation(JSONFieldExpectationConfig[T]{
		Path:       path,
		ExpectName: name,
		GetJSON:    s.GetJSON,
		PreCheck:   s.withPathValidation(path),
		Check:      JSONCheckNotEmpty(),
	})
}

func (s *JSONExpectationSource[T]) FieldIsNull(path string) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' is null", path)
	return BuildJSONFieldNullExpectation(JSONFieldNullExpectationConfig[T]{
		Path:         path,
		ExpectName:   name,
		GetJSON:      s.GetJSON,
		PreCheck:     s.withPathValidation(path),
		ExpectedNull: true,
	})
}

func (s *JSONExpectationSource[T]) FieldIsNullOrMissing(path string) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' is null or missing", path)
	return New(
		name,
		func(err error, result T) polling.CheckResult {
			if pathErr := ValidateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := s.PreCheck(err, result); !ok {
				return res
			}
			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
				}
			}
			jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{Ok: true}
			}
			if !jsonutil.IsNull(jsonRes) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Expected null, got %s", jsonutil.TypeToString(jsonRes.Type)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		StandardReport[T](name),
	)
}

func (s *JSONExpectationSource[T]) FieldIsNotNull(path string) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' is not null", path)
	return New(
		name,
		func(err error, result T) polling.CheckResult {
			if pathErr := ValidateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := s.PreCheckWithBody(err, result); !ok {
				return res
			}
			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
				}
			}
			jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("JSON field '%s' does not exist", path),
				}
			}
			if jsonutil.IsNull(jsonRes) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Expected non-null value, got null",
				}
			}
			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect JSON field '%s' is not null] %s", path, checkRes.Reason)
				return
			}
			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr == nil {
				res, _ := jsonutil.GetField(jsonBytes, path)
				a.True(true, "[Expect JSON field '%s' is not null] actual: %s", path, jsonutil.DebugValue(res))
			} else {
				a.True(true, "[Expect JSON field '%s' is not null]", path)
			}
		},
	)
}

func (s *JSONExpectationSource[T]) FieldTrue(path string) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' is true", path)
	return BuildJSONFieldExpectation(JSONFieldExpectationConfig[T]{
		Path:       path,
		ExpectName: name,
		GetJSON:    s.GetJSON,
		PreCheck:   s.withPathValidation(path),
		Check:      JSONCheckTrue(),
	})
}

func (s *JSONExpectationSource[T]) FieldFalse(path string) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' is false", path)
	return BuildJSONFieldExpectation(JSONFieldExpectationConfig[T]{
		Path:       path,
		ExpectName: name,
		GetJSON:    s.GetJSON,
		PreCheck:   s.withPathValidation(path),
		Check:      JSONCheckFalse(),
	})
}

func (s *JSONExpectationSource[T]) BodyEquals(expected any) *Expectation[T] {
	return BuildFullObjectExpectation(FullObjectExpectationConfig[T]{
		ExpectName: "Expect body matches (exact)",
		GetJSON:    s.GetJSON,
		PreCheck:   s.PreCheckWithBody,
		Expected:   expected,
		Compare:    jsonutil.CompareObjectExact,
		Retryable:  true,
	})
}

func (s *JSONExpectationSource[T]) BodyPartial(expected any) *Expectation[T] {
	return BuildFullObjectExpectation(FullObjectExpectationConfig[T]{
		ExpectName: "Expect body matches (partial)",
		GetJSON:    s.GetJSON,
		PreCheck:   s.PreCheckWithBody,
		Expected:   expected,
		Compare:    jsonutil.CompareObjectPartial,
		Retryable:  true,
	})
}

func (s *JSONExpectationSource[T]) FieldEmpty(path string) *Expectation[T] {
	name := fmt.Sprintf("Expect JSON field '%s' is empty", path)
	return New(
		name,
		func(err error, result T) polling.CheckResult {
			if pathErr := ValidateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := s.PreCheck(err, result); !ok {
				return res
			}
			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
				}
			}
			jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{Ok: true}
			}
			if !jsonutil.IsEmpty(jsonRes) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Field '%s' is not empty, got: %s", path, jsonRes.String()),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		StandardReport[T](name),
	)
}

func (s *JSONExpectationSource[T]) FieldJSON(path string, expected map[string]interface{}) []*Expectation[T] {
	var expectations []*Expectation[T]
	for key, value := range expected {
		fullPath := path + "." + key
		expectations = append(expectations, s.FieldEquals(fullPath, value))
	}
	return expectations
}

func (s *JSONExpectationSource[T]) ArrayContains(path string, expected any) *Expectation[T] {
	return New(
		fmt.Sprintf("Expect array '%s' contains matching object", path),
		func(err error, result T) polling.CheckResult {
			if pathErr := ValidateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := s.PreCheckWithBody(err, result); !ok {
				return res
			}
			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
				}
			}
			jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Path '%s' does not exist in response", path),
				}
			}
			if !jsonRes.IsArray() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected array at '%s', got %s", path, jsonutil.TypeToString(jsonRes.Type)),
				}
			}

			idx, _ := jsonutil.FindInArray(jsonRes, expected)
			if idx == -1 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("No matching object found in array '%s'", path),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect array '%s' contains matching object] %s", path, checkRes.Reason)
				return
			}

			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr == nil && len(jsonBytes) > 0 {
				jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
				if parseErr == nil && jsonRes.Exists() {
					idx, matchedItem := jsonutil.FindInArray(jsonRes, expected)
					if idx >= 0 {
						a.True(true, "[Expect array '%s' contains matching object] Found at index %d: %s", path, idx, jsonutil.DebugValue(matchedItem))
						return
					}
				}
			}
			a.True(true, "[Expect array '%s' contains matching object]", path)
		},
	)
}

func (s *JSONExpectationSource[T]) ArrayContainsExact(path string, expected any) *Expectation[T] {
	return New(
		fmt.Sprintf("Expect array '%s' contains exact matching object", path),
		func(err error, result T) polling.CheckResult {
			if pathErr := ValidateJSONPath(path); pathErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Invalid JSON path: %v", pathErr),
				}
			}
			if res, ok := s.PreCheckWithBody(err, result); !ok {
				return res
			}
			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Cannot get JSON: %v", jsonErr),
				}
			}
			jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
			if parseErr != nil {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    "Invalid JSON response body",
				}
			}
			if !jsonRes.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("Path '%s' does not exist in response", path),
				}
			}
			if !jsonRes.IsArray() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected array at '%s', got %s", path, jsonutil.TypeToString(jsonRes.Type)),
				}
			}

			idx, _ := jsonutil.FindInArrayExact(jsonRes, expected)
			if idx == -1 {
				partialIdx, partialItem := jsonutil.FindInArray(jsonRes, expected)
				if partialIdx >= 0 {
					_, diff := jsonutil.CompareObjectExact(partialItem, expected)
					return polling.CheckResult{
						Ok:        false,
						Retryable: true,
						Reason:    fmt.Sprintf("Found similar object at index %d but exact match failed: %s", partialIdx, diff),
					}
				}
				return polling.CheckResult{
					Ok:        false,
					Retryable: true,
					Reason:    fmt.Sprintf("No matching object found in array '%s'", path),
				}
			}

			return polling.CheckResult{Ok: true}
		},
		func(stepCtx provider.StepCtx, mode polling.AssertionMode, err error, result T, checkRes polling.CheckResult) {
			a := polling.PickAsserter(stepCtx, mode)
			if !checkRes.Ok {
				a.True(false, "[Expect array '%s' contains exact matching object] %s", path, checkRes.Reason)
				return
			}

			jsonBytes, jsonErr := s.GetJSON(result)
			if jsonErr == nil && len(jsonBytes) > 0 {
				jsonRes, parseErr := jsonutil.GetField(jsonBytes, path)
				if parseErr == nil && jsonRes.Exists() {
					idx, matchedItem := jsonutil.FindInArrayExact(jsonRes, expected)
					if idx >= 0 {
						a.True(true, "[Expect array '%s' contains exact matching object] Found at index %d: %s", path, idx, jsonutil.DebugValue(matchedItem))
						return
					}
				}
			}
			a.True(true, "[Expect array '%s' contains exact matching object]", path)
		},
	)
}

