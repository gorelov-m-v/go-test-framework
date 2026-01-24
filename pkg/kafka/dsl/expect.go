package dsl

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/jsonutil"
	kafkaErrors "github.com/gorelov-m-v/go-test-framework/internal/kafka/errors"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

type Expectation struct {
	sCtx        provider.StepCtx
	kafkaClient *client.Client
	topicName   string

	filters         map[string]string
	containsFilters map[string]string
	unique          bool
	duplicateWindow time.Duration

	expectedCount   int
	allMatchingMsgs [][]byte

	expectations []*expect.Expectation[[]byte]

	messageBytes []byte
	found        bool
}

func NewExpectation(sCtx provider.StepCtx, kafkaClient *client.Client, topicName string) *Expectation {
	return &Expectation{
		sCtx:            sCtx,
		kafkaClient:     kafkaClient,
		topicName:       topicName,
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
		expectations:    make([]*expect.Expectation[[]byte], 0),
	}
}

func (e *Expectation) With(key string, value interface{}) *Expectation {
	if value != nil {
		e.filters[key] = formatFilterValue(value)
	}
	return e
}

func (e *Expectation) WithContains(key string, value interface{}) *Expectation {
	if value != nil {
		e.containsFilters[key] = fmt.Sprintf("%v", value)
	}
	return e
}

func (e *Expectation) Unique() *Expectation {
	e.unique = true
	e.duplicateWindow = e.kafkaClient.GetUniqueWindow()
	return e
}

func (e *Expectation) UniqueWithWindow(window time.Duration) *Expectation {
	e.unique = true
	e.duplicateWindow = window
	return e
}

func (e *Expectation) ExpectCount(count int) *Expectation {
	e.expectedCount = count
	return e
}

func (e *Expectation) ExpectField(field string, expectedValue interface{}) *Expectation {
	e.expectations = append(e.expectations, makeFieldValueExpectation(field, expectedValue))
	return e
}

func (e *Expectation) ExpectJSONField(field string, expected map[string]interface{}) *Expectation {
	for key, value := range expected {
		path := field + "." + key
		e.expectations = append(e.expectations, makeFieldValueExpectation(path, value))
	}
	return e
}

func (e *Expectation) ExpectJsonField(field string, expected map[string]interface{}) *Expectation {
	return e.ExpectJSONField(field, expected)
}

func (e *Expectation) ExpectFieldNotEmpty(field string) *Expectation {
	e.expectations = append(e.expectations, makeFieldNotEmptyExpectation(field))
	return e
}

func (e *Expectation) ExpectFieldEmpty(field string) *Expectation {
	e.expectations = append(e.expectations, makeFieldEmptyExpectation(field))
	return e
}

func (e *Expectation) ExpectFieldIsNull(field string) *Expectation {
	e.expectations = append(e.expectations, makeFieldIsNullExpectation(field))
	return e
}

func (e *Expectation) ExpectFieldIsNotNull(field string) *Expectation {
	e.expectations = append(e.expectations, makeFieldIsNotNullExpectation(field))
	return e
}

func (e *Expectation) ExpectFieldTrue(field string) *Expectation {
	e.expectations = append(e.expectations, makeFieldTrueExpectation(field))
	return e
}

func (e *Expectation) ExpectFieldFalse(field string) *Expectation {
	e.expectations = append(e.expectations, makeFieldFalseExpectation(field))
	return e
}

func (e *Expectation) ExpectMessage(expected any) *Expectation {
	e.expectations = append(e.expectations, makeMessageExpectation(expected))
	return e
}

func (e *Expectation) ExpectMessagePartial(expected any) *Expectation {
	e.expectations = append(e.expectations, makeMessagePartialExpectation(expected))
	return e
}

func (e *Expectation) Send() {
	effectiveTimeout := e.kafkaClient.GetDefaultTimeout()

	stepName := fmt.Sprintf("Kafka: Expect from topic '%s'", e.topicName)

	e.sCtx.WithNewStep(stepName, func(stepCtx provider.StepCtx) {
		mode := polling.GetStepMode(stepCtx)

		var summary polling.PollingSummary

		if mode == polling.AsyncMode {
			e.messageBytes, e.found, summary = e.executeWithRetry(stepCtx)
		} else {
			e.messageBytes, e.found, summary = e.executeSingle()
		}

		if mode == polling.AsyncMode {
			polling.AttachPollingSummary(stepCtx, summary)
		}

		attachSearchInfoByTopic(stepCtx, e.topicName, e.filters, effectiveTimeout, e.unique)

		if e.found {
			if e.expectedCount > 0 && len(e.allMatchingMsgs) > 0 {
				attachAllFoundMessages(stepCtx, e.allMatchingMsgs)
			} else {
				var msgMap map[string]interface{}
				json.Unmarshal(e.messageBytes, &msgMap)
				attachFoundMessage(stepCtx, msgMap)
			}
		} else {
			attachNotFoundMessageByTopic(stepCtx, e.topicName, e.filters)
		}

		assertionMode := polling.GetAssertionModeFromStepMode(mode)

		if !e.found {
			msg := fmt.Sprintf("Kafka message in topic '%s' not found within %s. Filters: %v",
				e.topicName, effectiveTimeout, e.filters)

			if mode == polling.AsyncMode {
				msg = polling.FinalFailureMessage(summary)
			}

			polling.NoError(stepCtx, assertionMode, fmt.Errorf("%s", msg), msg)
			return
		}

		if e.expectedCount > 0 && e.found {
			e.checkExpectedCount(stepCtx, assertionMode)
		}

		if e.unique && e.found {
			e.checkUniqueness(stepCtx, assertionMode)
		}

		e.runExpectations(stepCtx, assertionMode)
	})
}

func (e *Expectation) doSearch() ([]byte, error) {
	if !e.kafkaClient.GetBuffer().IsTopicConfigured(e.topicName) {
		return nil, &kafkaErrors.KafkaTopicNotListenedError{
			TopicName:        e.topicName,
			MessageType:      "unknown",
			ConfiguredTopics: e.kafkaClient.GetBuffer().GetConfiguredTopics(),
		}
	}

	messages := e.kafkaClient.GetBuffer().GetMessages(e.topicName)

	if e.expectedCount > 0 {
		allMatching, err := e.findAllMatching(messages)
		if err != nil {
			return nil, err
		}
		e.allMatchingMsgs = allMatching
		if len(allMatching) < e.expectedCount {
			return nil, fmt.Errorf("expected %d messages, found %d", e.expectedCount, len(allMatching))
		}
		return allMatching[0], nil
	}

	if e.unique {
		msgBytes, err := e.findAndCountWithinWindow(messages)
		return msgBytes, err
	}

	msgBytes, err := e.searchMessage(messages)
	return msgBytes, err
}

func (e *Expectation) searchMessage(messages []*types.KafkaMessage) ([]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages in buffer")
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if e.matchesFilter(msg.Value) {
			return msg.Value, nil
		}
	}

	return nil, fmt.Errorf("message not found")
}

func (e *Expectation) findAllMatching(messages []*types.KafkaMessage) ([][]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages in buffer")
	}

	var result [][]byte
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if e.matchesFilter(msg.Value) {
			result = append(result, msg.Value)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("message not found")
	}

	return result, nil
}

func (e *Expectation) findAndCountWithinWindow(messages []*types.KafkaMessage) ([]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages in buffer")
	}

	var firstMatchBytes []byte
	var firstMatchTimestamp int64
	count := 0

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if e.matchesFilter(msg.Value) {
			if count == 0 {
				firstMatchBytes = msg.Value
				firstMatchTimestamp = msg.Timestamp
				count++
			} else {
				timeDiff := abs(msg.Timestamp - firstMatchTimestamp)
				if timeDiff <= e.duplicateWindow.Milliseconds() {
					count++
				}
			}
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("message not found")
	}

	if count > 1 {
		return nil, &kafkaErrors.KafkaMessageNotUniqueError{
			MessageType: e.topicName,
			Filters:     e.filters,
			Count:       count,
			WindowMs:    e.duplicateWindow.Milliseconds(),
		}
	}

	return firstMatchBytes, nil
}

func (e *Expectation) checkUniqueness(stepCtx provider.StepCtx, mode polling.AssertionMode) {
	messages := e.kafkaClient.GetBuffer().GetMessages(e.topicName)
	_, err := e.findAndCountWithinWindow(messages)

	if err != nil {
		if notUniqueErr, ok := err.(*kafkaErrors.KafkaMessageNotUniqueError); ok {
			polling.NoError(stepCtx, mode, notUniqueErr, notUniqueErr.Error())
		}
	}
}

func (e *Expectation) checkExpectedCount(stepCtx provider.StepCtx, mode polling.AssertionMode) {
	actualCount := len(e.allMatchingMsgs)
	if actualCount != e.expectedCount {
		msg := fmt.Sprintf("Expected %d Kafka messages, but found %d. Topic: %s, Filters: %v",
			e.expectedCount, actualCount, e.topicName, e.filters)
		polling.NoError(stepCtx, mode, fmt.Errorf("%s", msg), msg)
	}
}

func (e *Expectation) runExpectations(stepCtx provider.StepCtx, mode polling.AssertionMode) {
	if len(e.expectations) == 0 {
		return
	}
	expect.ReportAll(stepCtx, mode, e.expectations, nil, e.messageBytes)
}

func (e *Expectation) matchesFilter(jsonValue []byte) bool {
	if len(jsonValue) == 0 {
		return len(e.filters) == 0 && len(e.containsFilters) == 0
	}

	if len(e.filters) == 0 && len(e.containsFilters) == 0 {
		return true
	}

	if !gjson.ValidBytes(jsonValue) {
		return false
	}

	for path, expectedValue := range e.filters {
		result := gjson.GetBytes(jsonValue, path)

		if !result.Exists() {
			return false
		}

		// For arrays, compare as sets (order-independent)
		if result.IsArray() {
			if !compareArraysAsSet(result, expectedValue) {
				return false
			}
		} else {
			actualValue := result.String()
			if actualValue != expectedValue {
				return false
			}
		}
	}

	for path, expectedValue := range e.containsFilters {
		result := gjson.GetBytes(jsonValue, path)

		if !result.Exists() {
			return false
		}

		if !result.IsArray() {
			return false
		}

		found := false
		for _, item := range result.Array() {
			if item.String() == expectedValue {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (e *Expectation) buildFilterDescription() string {
	if len(e.filters) == 0 && len(e.containsFilters) == 0 {
		return ""
	}

	parts := make([]string, 0, len(e.filters)+len(e.containsFilters))
	for key, value := range e.filters {
		parts = append(parts, fmt.Sprintf("%s = %s", key, value))
	}
	for key, value := range e.containsFilters {
		parts = append(parts, fmt.Sprintf("%s contains %s", key, value))
	}

	return strings.Join(parts, ", ")
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func formatFilterValue(value interface{}) string {
	if value == nil {
		return ""
	}
	// Check if value is a slice using reflection
	switch v := value.(type) {
	case []string:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// compareArraysAsSet compares a gjson array result with expected JSON array string
// Returns true if both arrays contain the same elements with same counts (order-independent)
func compareArraysAsSet(result gjson.Result, expectedJSON string) bool {
	var expectedArr []string
	if err := json.Unmarshal([]byte(expectedJSON), &expectedArr); err != nil {
		return false
	}

	actualArr := result.Array()

	if len(actualArr) != len(expectedArr) {
		return false
	}

	expectedCounts := make(map[string]int)
	for _, v := range expectedArr {
		expectedCounts[v]++
	}

	actualCounts := make(map[string]int)
	for _, item := range actualArr {
		actualCounts[item.String()]++
	}

	for key, expectedCount := range expectedCounts {
		if actualCounts[key] != expectedCount {
			return false
		}
	}

	return true
}

func makeFieldValueExpectation(field string, expectedValue interface{}) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' = %v", field, expectedValue)
	return expect.New(
		name,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}
			result := gjson.GetBytes(msgBytes, field)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' not found", field),
				}
			}
			ok, msg := jsonutil.Compare(result, expectedValue)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeFieldNotEmptyExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' not empty", field)
	return expect.New(
		name,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}
			result := gjson.GetBytes(msgBytes, field)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' not found", field),
				}
			}
			if jsonutil.IsEmpty(result) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' is empty", field),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeFieldEmptyExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is empty", field)
	return expect.New(
		name,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}
			result := gjson.GetBytes(msgBytes, field)
			if !result.Exists() {
				return polling.CheckResult{Ok: true}
			}
			if !jsonutil.IsEmpty(result) {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' is not empty, got: %s", field, result.String()),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeFieldIsNullExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is null", field)
	return expect.New(
		name,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}
			result := gjson.GetBytes(msgBytes, field)
			if result.Exists() && result.Type != gjson.Null {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected null, got %s: %s", jsonutil.TypeToString(result.Type), jsonutil.DebugValue(result)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeFieldIsNotNullExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is not null", field)
	return expect.New(
		name,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}
			result := gjson.GetBytes(msgBytes, field)
			if !result.Exists() || result.Type == gjson.Null {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' is null", field),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeFieldTrueExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is true", field)
	return expect.New(
		name,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}
			result := gjson.GetBytes(msgBytes, field)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' not found", field),
				}
			}
			if result.Type != gjson.True {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected true, got %s: %s", jsonutil.TypeToString(result.Type), jsonutil.DebugValue(result)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeFieldFalseExpectation(field string) *expect.Expectation[[]byte] {
	name := fmt.Sprintf("Expect: Field '%s' is false", field)
	return expect.New(
		name,
		func(err error, msgBytes []byte) polling.CheckResult {
			if len(msgBytes) == 0 {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    "Message bytes are empty",
				}
			}
			result := gjson.GetBytes(msgBytes, field)
			if !result.Exists() {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Field '%s' not found", field),
				}
			}
			if result.Type != gjson.False {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    fmt.Sprintf("Expected false, got %s: %s", jsonutil.TypeToString(result.Type), jsonutil.DebugValue(result)),
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeMessageExpectation(expected any) *expect.Expectation[[]byte] {
	name := "Expect: Message matches (exact)"
	return expect.New(
		name,
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
			ok, msg := jsonutil.CompareObjectExact(jsonRes, expected)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}

func makeMessagePartialExpectation(expected any) *expect.Expectation[[]byte] {
	name := "Expect: Message matches (partial)"
	return expect.New(
		name,
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
			ok, msg := jsonutil.CompareObjectPartial(jsonRes, expected)
			if !ok {
				return polling.CheckResult{
					Ok:        false,
					Retryable: false,
					Reason:    msg,
				}
			}
			return polling.CheckResult{Ok: true}
		},
		expect.StandardReport[[]byte](name),
	)
}
