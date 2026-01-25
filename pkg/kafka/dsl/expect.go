package dsl

import (
	"fmt"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
)

func (q *Query[T]) addExpectation(exp *expect.Expectation[[]byte]) {
	if q.sent {
		q.sCtx.Break("Kafka DSL Error: Expectations must be added before Send().")
		q.sCtx.BrokenNow()
		return
	}
	q.expectations = append(q.expectations, exp)
}

// ExpectFieldEquals checks that a JSON field at the given path equals the expected value.
// Supports numeric type coercion (int, int64, float64 are compared by value).
func (q *Query[T]) ExpectFieldEquals(field string, expectedValue interface{}) *Query[T] {
	q.addExpectation(makeFieldValueExpectation(field, expectedValue))
	return q
}

// Deprecated: Use ExpectFieldEquals instead. Will be removed in v2.0.
func (q *Query[T]) ExpectField(field string, expectedValue interface{}) *Query[T] {
	return q.ExpectFieldEquals(field, expectedValue)
}

// ExpectFieldJSON checks that a JSON object field contains all expected key-value pairs.
func (q *Query[T]) ExpectFieldJSON(field string, expected map[string]interface{}) *Query[T] {
	for key, value := range expected {
		path := field + "." + key
		q.addExpectation(makeFieldValueExpectation(path, value))
	}
	return q
}

// Deprecated: Use ExpectFieldJSON instead. Will be removed in v2.0.
func (q *Query[T]) ExpectJSONField(field string, expected map[string]interface{}) *Query[T] {
	return q.ExpectFieldJSON(field, expected)
}

// ExpectFieldNotEmpty checks that a JSON field at the given path is not empty.
func (q *Query[T]) ExpectFieldNotEmpty(field string) *Query[T] {
	q.addExpectation(makeFieldNotEmptyExpectation(field))
	return q
}

// ExpectFieldEmpty checks that a JSON field at the given path is empty.
func (q *Query[T]) ExpectFieldEmpty(field string) *Query[T] {
	q.addExpectation(makeFieldEmptyExpectation(field))
	return q
}

// ExpectFieldIsNull checks that a JSON field at the given path is null.
func (q *Query[T]) ExpectFieldIsNull(field string) *Query[T] {
	q.addExpectation(makeFieldIsNullExpectation(field))
	return q
}

// ExpectFieldIsNotNull checks that a JSON field at the given path is not null.
func (q *Query[T]) ExpectFieldIsNotNull(field string) *Query[T] {
	q.addExpectation(makeFieldIsNotNullExpectation(field))
	return q
}

// ExpectFieldTrue checks that a JSON boolean field at the given path is true.
func (q *Query[T]) ExpectFieldTrue(field string) *Query[T] {
	q.addExpectation(makeFieldTrueExpectation(field))
	return q
}

// ExpectFieldFalse checks that a JSON boolean field at the given path is false.
func (q *Query[T]) ExpectFieldFalse(field string) *Query[T] {
	q.addExpectation(makeFieldFalseExpectation(field))
	return q
}

// ExpectBodyEquals checks that the message body exactly matches the expected struct or map (all fields must match).
func (q *Query[T]) ExpectBodyEquals(expected any) *Query[T] {
	q.addExpectation(makeMessageExpectation(expected))
	return q
}

// Deprecated: Use ExpectBodyEquals instead. Will be removed in v2.0.
func (q *Query[T]) ExpectMessage(expected any) *Query[T] {
	return q.ExpectBodyEquals(expected)
}

// ExpectBodyPartial checks that the message body contains fields from the expected struct or map (non-zero fields only).
func (q *Query[T]) ExpectBodyPartial(expected any) *Query[T] {
	q.addExpectation(makeMessagePartialExpectation(expected))
	return q
}

// Deprecated: Use ExpectBodyPartial instead. Will be removed in v2.0.
func (q *Query[T]) ExpectMessagePartial(expected any) *Query[T] {
	return q.ExpectBodyPartial(expected)
}

func makeFieldValueExpectation(field string, expectedValue interface{}) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' = %v", field, expectedValue)
	return expect.BuildBytesJSONFieldExpectation(expect.BytesJSONFieldExpectationConfig{
		Path:       field,
		ExpectName: name,
		Check:      expect.JSONCheckEquals(expectedValue),
	})
}

func makeFieldNotEmptyExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' not empty", field)
	return expect.BuildBytesJSONFieldExpectation(expect.BytesJSONFieldExpectationConfig{
		Path:       field,
		ExpectName: name,
		Check:      expect.JSONCheckNotEmpty(),
	})
}

func makeFieldEmptyExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is empty", field)
	return expect.BuildBytesJSONFieldWithExistsCheck(expect.BytesJSONFieldExistsCheckConfig{
		Path:          field,
		ExpectName:    name,
		RequireExists: false,
		Check:         expect.JSONCheckEmpty(),
	})
}

func makeFieldIsNullExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is null", field)
	return expect.BuildBytesJSONFieldNullCheck(expect.BytesJSONFieldNullCheckConfig{
		Path:         field,
		ExpectName:   name,
		ExpectedNull: true,
	})
}

func makeFieldIsNotNullExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is not null", field)
	return expect.BuildBytesJSONFieldNullCheck(expect.BytesJSONFieldNullCheckConfig{
		Path:         field,
		ExpectName:   name,
		ExpectedNull: false,
	})
}

func makeFieldTrueExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is true", field)
	return expect.BuildBytesJSONFieldExpectation(expect.BytesJSONFieldExpectationConfig{
		Path:       field,
		ExpectName: name,
		Check:      expect.JSONCheckTrue(),
	})
}

func makeFieldFalseExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is false", field)
	return expect.BuildBytesJSONFieldExpectation(expect.BytesJSONFieldExpectationConfig{
		Path:       field,
		ExpectName: name,
		Check:      expect.JSONCheckFalse(),
	})
}

func makeMessageExpectation(expected any) *expect.Expectation[[]byte] {
	return expect.BuildBytesObjectExpectation(expect.BytesObjectExpectationConfig{
		ExpectName: "Expect: Message matches (exact)",
		Expected:   expected,
		Compare:    jsonutil.CompareObjectExact,
	})
}

func makeMessagePartialExpectation(expected any) *expect.Expectation[[]byte] {
	return expect.BuildBytesObjectExpectation(expect.BytesObjectExpectationConfig{
		ExpectName: "Expect: Message matches (partial)",
		Expected:   expected,
		Compare:    jsonutil.CompareObjectPartial,
	})
}
