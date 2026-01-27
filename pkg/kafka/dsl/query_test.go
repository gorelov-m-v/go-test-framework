package dsl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

func TestSearchMessage_Found(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"id": 1, "name": "first"}`)},
		{Value: []byte(`{"id": 2, "name": "second"}`)},
		{Value: []byte(`{"id": 3, "name": "third"}`)},
	}

	q := &Query[any]{
		filters:         map[string]string{"id": "2"},
		containsFilters: make(map[string]string),
	}

	result, err := q.searchMessage(messages)

	require.NoError(t, err)
	assert.Equal(t, []byte(`{"id": 2, "name": "second"}`), result)
}

func TestSearchMessage_NotFound(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"id": 1, "name": "first"}`)},
		{Value: []byte(`{"id": 2, "name": "second"}`)},
	}

	q := &Query[any]{
		filters:         map[string]string{"id": "999"},
		containsFilters: make(map[string]string),
	}

	result, err := q.searchMessage(messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
	assert.Nil(t, result)
}

func TestSearchMessage_EmptyBuffer(t *testing.T) {
	messages := []*types.KafkaMessage{}

	q := &Query[any]{
		filters:         map[string]string{"id": "1"},
		containsFilters: make(map[string]string),
	}

	result, err := q.searchMessage(messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no messages in buffer")
	assert.Nil(t, result)
}

func TestSearchMessage_ReturnsLatestMatch(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "event", "seq": 1}`)},
		{Value: []byte(`{"type": "event", "seq": 2}`)},
		{Value: []byte(`{"type": "event", "seq": 3}`)},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "event"},
		containsFilters: make(map[string]string),
	}

	result, err := q.searchMessage(messages)

	require.NoError(t, err)
	assert.Contains(t, string(result), `"seq": 3`)
}

func TestSearchMessage_NoFilters(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"id": 1}`)},
		{Value: []byte(`{"id": 2}`)},
	}

	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	result, err := q.searchMessage(messages)

	require.NoError(t, err)
	assert.Equal(t, []byte(`{"id": 2}`), result)
}

func TestFindAllMatching_Multiple(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "A", "id": 1}`)},
		{Value: []byte(`{"type": "B", "id": 2}`)},
		{Value: []byte(`{"type": "A", "id": 3}`)},
		{Value: []byte(`{"type": "A", "id": 4}`)},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "A"},
		containsFilters: make(map[string]string),
	}

	result, err := q.findAllMatching(messages)

	require.NoError(t, err)
	assert.Len(t, result, 3)
}

func TestFindAllMatching_None(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "A"}`)},
		{Value: []byte(`{"type": "B"}`)},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "C"},
		containsFilters: make(map[string]string),
	}

	result, err := q.findAllMatching(messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
	assert.Nil(t, result)
}

func TestFindAllMatching_EmptyBuffer(t *testing.T) {
	messages := []*types.KafkaMessage{}

	q := &Query[any]{
		filters:         map[string]string{"type": "A"},
		containsFilters: make(map[string]string),
	}

	result, err := q.findAllMatching(messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no messages in buffer")
	assert.Nil(t, result)
}

func TestFindAllMatching_ReturnsInReverseOrder(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "A", "seq": 1}`)},
		{Value: []byte(`{"type": "A", "seq": 2}`)},
		{Value: []byte(`{"type": "A", "seq": 3}`)},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "A"},
		containsFilters: make(map[string]string),
	}

	result, err := q.findAllMatching(messages)

	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Contains(t, string(result[0]), `"seq": 3`)
	assert.Contains(t, string(result[2]), `"seq": 1`)
}

func TestFindAndCountWithinWindow_Unique(t *testing.T) {
	baseTime := time.Now().UnixMilli()
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "A"}`), Timestamp: baseTime},
		{Value: []byte(`{"type": "B"}`), Timestamp: baseTime + 100},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "A"},
		containsFilters: make(map[string]string),
		duplicateWindow: 1 * time.Second,
	}

	result, err := q.findAndCountWithinWindow(messages)

	require.NoError(t, err)
	assert.Equal(t, []byte(`{"type": "A"}`), result)
}

func TestFindAndCountWithinWindow_Duplicates(t *testing.T) {
	baseTime := time.Now().UnixMilli()
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "A", "seq": 1}`), Timestamp: baseTime},
		{Value: []byte(`{"type": "A", "seq": 2}`), Timestamp: baseTime + 100},
		{Value: []byte(`{"type": "A", "seq": 3}`), Timestamp: baseTime + 200},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "A"},
		containsFilters: make(map[string]string),
		duplicateWindow: 1 * time.Second,
	}

	result, err := q.findAndCountWithinWindow(messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected once but found")
	assert.Nil(t, result)
}

func TestFindAndCountWithinWindow_DuplicatesOutsideWindow(t *testing.T) {
	baseTime := time.Now().UnixMilli()
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "A", "seq": 1}`), Timestamp: baseTime},
		{Value: []byte(`{"type": "A", "seq": 2}`), Timestamp: baseTime + 5000},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "A"},
		containsFilters: make(map[string]string),
		duplicateWindow: 1 * time.Second,
	}

	result, err := q.findAndCountWithinWindow(messages)

	require.NoError(t, err)
	assert.Contains(t, string(result), `"seq": 2`)
}

func TestFindAndCountWithinWindow_NotFound(t *testing.T) {
	messages := []*types.KafkaMessage{
		{Value: []byte(`{"type": "A"}`)},
	}

	q := &Query[any]{
		filters:         map[string]string{"type": "B"},
		containsFilters: make(map[string]string),
		duplicateWindow: 1 * time.Second,
	}

	result, err := q.findAndCountWithinWindow(messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
	assert.Nil(t, result)
}

func TestFindAndCountWithinWindow_EmptyBuffer(t *testing.T) {
	messages := []*types.KafkaMessage{}

	q := &Query[any]{
		filters:         map[string]string{"type": "A"},
		containsFilters: make(map[string]string),
		duplicateWindow: 1 * time.Second,
	}

	result, err := q.findAndCountWithinWindow(messages)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no messages in buffer")
	assert.Nil(t, result)
}

func TestBuildResult_Found(t *testing.T) {
	type TestMessage struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	q := &Query[TestMessage]{
		found:        true,
		messageBytes: []byte(`{"id": 123, "name": "test"}`),
	}

	result := q.buildResult()

	assert.True(t, result.Found)
	assert.Equal(t, 123, result.Message.ID)
	assert.Equal(t, "test", result.Message.Name)
	assert.Nil(t, result.ParseError)
	assert.Equal(t, 1, result.MatchCount)
}

func TestBuildResult_NotFound(t *testing.T) {
	q := &Query[any]{
		found:        false,
		messageBytes: nil,
	}

	result := q.buildResult()

	assert.False(t, result.Found)
	assert.Nil(t, result.RawMessage)
	assert.Equal(t, 0, result.MatchCount)
}

func TestBuildResult_ParseError(t *testing.T) {
	type TestMessage struct {
		ID int `json:"id"`
	}

	q := &Query[TestMessage]{
		found:        true,
		messageBytes: []byte(`{"id": "not_a_number"}`),
	}

	result := q.buildResult()

	assert.True(t, result.Found)
	assert.NotNil(t, result.ParseError)
	assert.Contains(t, result.ParseError.Error(), "failed to parse message")
}

func TestBuildResult_WithAllMessages(t *testing.T) {
	q := &Query[any]{
		found:           true,
		messageBytes:    []byte(`{"id": 1}`),
		allMatchingMsgs: [][]byte{[]byte(`{"id": 1}`), []byte(`{"id": 2}`), []byte(`{"id": 3}`)},
	}

	result := q.buildResult()

	assert.True(t, result.Found)
	assert.Len(t, result.AllMessages, 3)
	assert.Equal(t, 3, result.MatchCount)
}

func TestBuildResult_EmptyMessageBytes(t *testing.T) {
	type TestMessage struct {
		ID int `json:"id"`
	}

	q := &Query[TestMessage]{
		found:        true,
		messageBytes: []byte{},
	}

	result := q.buildResult()

	assert.True(t, result.Found)
	assert.Nil(t, result.ParseError)
}

func TestQuery_With(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	q.With("key1", "value1").With("key2", 123)

	assert.Equal(t, "value1", q.filters["key1"])
	assert.Equal(t, "123", q.filters["key2"])
}

func TestQuery_With_NilValue(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	q.With("key", nil)

	assert.Empty(t, q.filters)
}

func TestQuery_WithContains(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	q.WithContains("tags", "important")

	assert.Equal(t, "important", q.containsFilters["tags"])
}

func TestQuery_WithContains_NilValue(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	q.WithContains("tags", nil)

	assert.Empty(t, q.containsFilters)
}

func TestQuery_ExpectCount(t *testing.T) {
	q := &Query[any]{}

	q.ExpectCount(5)

	assert.Equal(t, 5, q.expectedCount)
}

func TestQuery_Chaining(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	result := q.With("id", "123").WithContains("tags", "test").ExpectCount(2)

	assert.Same(t, q, result)
	assert.Equal(t, "123", q.filters["id"])
	assert.Equal(t, "test", q.containsFilters["tags"])
	assert.Equal(t, 2, q.expectedCount)
}

func TestQuery_UniqueWithWindow(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	result := q.UniqueWithWindow(5 * time.Second)

	assert.Same(t, q, result)
	assert.True(t, q.unique)
	assert.Equal(t, 5*time.Second, q.duplicateWindow)
}

func TestQuery_UniqueWithWindow_Chaining(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	q.With("id", "123").UniqueWithWindow(10 * time.Second).ExpectCount(1)

	assert.Equal(t, "123", q.filters["id"])
	assert.True(t, q.unique)
	assert.Equal(t, 10*time.Second, q.duplicateWindow)
	assert.Equal(t, 1, q.expectedCount)
}

func TestQuery_Context(t *testing.T) {
	q := &Query[any]{
		ctx: context.Background(),
	}

	customCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := q.Context(customCtx)

	assert.Same(t, q, result)
	assert.Same(t, customCtx, q.ctx)
}

func TestQuery_StepName(t *testing.T) {
	q := &Query[any]{
		topicName: "test-topic",
	}

	name := q.stepName()

	assert.Equal(t, "Kafka: Consume from 'test-topic'", name)
}

func TestQuery_StepName_WithPrefix(t *testing.T) {
	q := &Query[any]{
		topicName: "prefix.events.user-created",
	}

	name := q.stepName()

	assert.Equal(t, "Kafka: Consume from 'prefix.events.user-created'", name)
}

func TestQuery_BuildChecker_WithError(t *testing.T) {
	q := &Query[any]{
		expectations: nil,
	}

	checker := q.buildChecker()
	results := checker(nil, assert.AnError)

	require.Len(t, results, 1)
	assert.False(t, results[0].Ok)
	assert.True(t, results[0].Retryable)
	assert.Contains(t, results[0].Reason, "assert.AnError")
}

func TestQuery_BuildChecker_NoExpectations(t *testing.T) {
	q := &Query[any]{
		expectations: nil,
	}

	checker := q.buildChecker()
	results := checker([]byte(`{"id": 1}`), nil)

	require.Len(t, results, 1)
	assert.True(t, results[0].Ok)
}

func TestQuery_BuildChecker_WithExpectations(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
	}

	q.expectations = append(q.expectations, makeFieldValueExpectation("id", 123))

	checker := q.buildChecker()
	results := checker([]byte(`{"id": 123}`), nil)

	require.Len(t, results, 1)
	assert.True(t, results[0].Ok)
}

func TestQuery_BuildChecker_WithFailingExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
	}

	q.expectations = append(q.expectations, makeFieldValueExpectation("id", 999))

	checker := q.buildChecker()
	results := checker([]byte(`{"id": 123}`), nil)

	require.Len(t, results, 1)
	assert.False(t, results[0].Ok)
	assert.True(t, results[0].Retryable)
}

func TestQuery_BuildChecker_MultipleExpectations(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
	}

	q.expectations = append(q.expectations, makeFieldValueExpectation("id", 123))
	q.expectations = append(q.expectations, makeFieldValueExpectation("name", "test"))

	checker := q.buildChecker()
	results := checker([]byte(`{"id": 123, "name": "test"}`), nil)

	require.Len(t, results, 2)
	assert.True(t, results[0].Ok)
	assert.True(t, results[1].Ok)
}

func TestQuery_ExpectFieldEquals_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	result := q.ExpectFieldEquals("id", 123)

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectFieldNotEmpty_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	result := q.ExpectFieldNotEmpty("id")

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectFieldEmpty_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	result := q.ExpectFieldEmpty("id")

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectFieldIsNull_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	result := q.ExpectFieldIsNull("id")

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectFieldIsNotNull_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	result := q.ExpectFieldIsNotNull("id")

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectFieldTrue_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	result := q.ExpectFieldTrue("active")

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectFieldFalse_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	result := q.ExpectFieldFalse("deleted")

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectBodyEquals_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	expected := map[string]interface{}{"id": 123}
	result := q.ExpectBodyEquals(expected)

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectBodyPartial_AddsExpectation(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	expected := map[string]interface{}{"id": 123}
	result := q.ExpectBodyPartial(expected)

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 1)
}

func TestQuery_ExpectFieldJSON_AddsMultipleExpectations(t *testing.T) {
	q := &Query[any]{
		expectations: make([]*expect.Expectation[[]byte], 0),
		sent:         false,
	}

	expected := map[string]interface{}{
		"name": "test",
		"age":  25,
	}
	result := q.ExpectFieldJSON("user", expected)

	assert.Same(t, q, result)
	assert.Len(t, q.expectations, 2)
}

func TestQuery_ExpectationChaining(t *testing.T) {
	q := &Query[any]{
		expectations:    make([]*expect.Expectation[[]byte], 0),
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
		sent:            false,
	}

	q.With("type", "user").
		ExpectFieldEquals("id", 123).
		ExpectFieldNotEmpty("name").
		ExpectFieldIsNotNull("email").
		ExpectFieldTrue("active")

	assert.Equal(t, "user", q.filters["type"])
	assert.Len(t, q.expectations, 4)
}
