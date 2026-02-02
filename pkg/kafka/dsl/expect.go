package dsl

import (
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
)

var bytesPreCheck = client.BuildBytesPreCheck()

var bytesSource = &expect.JSONExpectationSource[[]byte]{
	GetJSON:          func(b []byte) ([]byte, error) { return b, nil },
	PreCheck:         bytesPreCheck,
	PreCheckWithBody: bytesPreCheck,
}

func (q *Query[T]) addExpectation(exp *expect.Expectation[[]byte]) {
	expect.AddExpectation(q.sCtx, q.sent, &q.expectations, exp, "Kafka")
}

func (q *Query[T]) ExpectFieldEquals(field string, expectedValue interface{}) *Query[T] {
	q.addExpectation(bytesSource.FieldEquals(field, expectedValue))
	return q
}

func (q *Query[T]) ExpectFieldJSON(field string, expected map[string]interface{}) *Query[T] {
	for _, exp := range bytesSource.FieldJSON(field, expected) {
		q.addExpectation(exp)
	}
	return q
}

func (q *Query[T]) ExpectFieldNotEmpty(field string) *Query[T] {
	q.addExpectation(bytesSource.FieldNotEmpty(field))
	return q
}

func (q *Query[T]) ExpectFieldEmpty(field string) *Query[T] {
	q.addExpectation(bytesSource.FieldEmpty(field))
	return q
}

func (q *Query[T]) ExpectFieldIsNull(field string) *Query[T] {
	q.addExpectation(bytesSource.FieldIsNull(field))
	return q
}

func (q *Query[T]) ExpectFieldIsNotNull(field string) *Query[T] {
	q.addExpectation(bytesSource.FieldIsNotNull(field))
	return q
}

func (q *Query[T]) ExpectFieldTrue(field string) *Query[T] {
	q.addExpectation(bytesSource.FieldTrue(field))
	return q
}

func (q *Query[T]) ExpectFieldFalse(field string) *Query[T] {
	q.addExpectation(bytesSource.FieldFalse(field))
	return q
}

func (q *Query[T]) ExpectBodyEquals(expected any) *Query[T] {
	q.addExpectation(bytesSource.BodyEquals(expected))
	return q
}

func (q *Query[T]) ExpectBodyPartial(expected any) *Query[T] {
	q.addExpectation(bytesSource.BodyPartial(expected))
	return q
}
