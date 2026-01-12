package dsl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/tidwall/gjson"

	kafkaErrors "github.com/gorelov-m-v/go-test-framework/internal/kafka/errors"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/extension"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

type Expectation struct {
	sCtx        provider.StepCtx
	kafkaClient *client.Client
	topicName   string

	// Параметры поиска
	filters         map[string]string
	unique          bool
	duplicateWindow time.Duration

	// Ожидания (expectations)
	expectations []fieldExpectation

	// Результат
	messageBytes []byte
	found        bool
}

type fieldExpectation struct {
	field     string
	value     interface{}
	checkType string // "equals", "notEmpty", "isNull", "isNotNull", "true", "false"
}

func NewExpectation(sCtx provider.StepCtx, kafkaClient *client.Client, topicName string) *Expectation {
	return &Expectation{
		sCtx:         sCtx,
		kafkaClient:  kafkaClient,
		topicName:    topicName,
		filters:      make(map[string]string),
		expectations: make([]fieldExpectation, 0),
	}
}

// With добавляет фильтр для поиска сообщения
func (e *Expectation) With(key string, value interface{}) *Expectation {
	if value != nil {
		e.filters[key] = fmt.Sprintf("%v", value)
	}
	return e
}

// Unique проверяет уникальность сообщения в окне
func (e *Expectation) Unique() *Expectation {
	e.unique = true
	e.duplicateWindow = e.kafkaClient.GetUniqueWindow()
	return e
}

// UniqueWithWindow проверяет уникальность с кастомным окном
func (e *Expectation) UniqueWithWindow(window time.Duration) *Expectation {
	e.unique = true
	e.duplicateWindow = window
	return e
}

// ExpectField проверяет значение поля
func (e *Expectation) ExpectField(field string, expectedValue interface{}) *Expectation {
	e.expectations = append(e.expectations, fieldExpectation{
		field:     field,
		value:     expectedValue,
		checkType: "equals",
	})
	return e
}

// ExpectFieldNotEmpty проверяет что поле не пустое
func (e *Expectation) ExpectFieldNotEmpty(field string) *Expectation {
	e.expectations = append(e.expectations, fieldExpectation{
		field:     field,
		checkType: "notEmpty",
	})
	return e
}

// ExpectFieldIsNull проверяет что поле null
func (e *Expectation) ExpectFieldIsNull(field string) *Expectation {
	e.expectations = append(e.expectations, fieldExpectation{
		field:     field,
		checkType: "isNull",
	})
	return e
}

// ExpectFieldIsNotNull проверяет что поле не null
func (e *Expectation) ExpectFieldIsNotNull(field string) *Expectation {
	e.expectations = append(e.expectations, fieldExpectation{
		field:     field,
		checkType: "isNotNull",
	})
	return e
}

// ExpectFieldTrue проверяет что поле = true
func (e *Expectation) ExpectFieldTrue(field string) *Expectation {
	e.expectations = append(e.expectations, fieldExpectation{
		field:     field,
		checkType: "true",
	})
	return e
}

// ExpectFieldFalse проверяет что поле = false
func (e *Expectation) ExpectFieldFalse(field string) *Expectation {
	e.expectations = append(e.expectations, fieldExpectation{
		field:     field,
		checkType: "false",
	})
	return e
}

// Send выполняет поиск и проверки (ничего не возвращает)
func (e *Expectation) Send() {
	effectiveTimeout := e.kafkaClient.GetDefaultTimeout()

	stepName := fmt.Sprintf("Kafka: Expect from topic '%s'", e.topicName)

	e.sCtx.WithNewStep(stepName, func(stepCtx provider.StepCtx) {
		mode := extension.GetStepMode(stepCtx)

		var summary extension.PollingSummary

		if mode == extension.AsyncMode {
			e.messageBytes, e.found, summary = e.fetchWithRetry(stepCtx)
		} else {
			e.messageBytes, e.found, summary = e.fetchOnce()
		}

		if mode == extension.AsyncMode {
			extension.AttachPollingSummary(stepCtx, summary)
		}

		attachSearchInfoByTopic(stepCtx, e.topicName, e.filters, effectiveTimeout, e.unique)

		if e.found {
			// Десериализуем для attach
			var msgMap map[string]interface{}
			json.Unmarshal(e.messageBytes, &msgMap)
			attachFoundMessage(stepCtx, msgMap)
		} else {
			attachNotFoundMessageByTopic(stepCtx, e.topicName, e.filters)
		}

		assertionMode := extension.GetAssertionModeFromStepMode(mode)

		if !e.found {
			msg := fmt.Sprintf("Kafka message in topic '%s' not found within %s. Filters: %v",
				e.topicName, effectiveTimeout, e.filters)

			if mode == extension.AsyncMode {
				msg = extension.FinalFailureMessage(summary)
			}

			extension.NoError(stepCtx, assertionMode, fmt.Errorf("%s", msg), msg)
			return
		}

		if e.unique && e.found {
			e.checkUniqueness(stepCtx, assertionMode)
		}

		// Выполняем все expectations
		e.runExpectations(stepCtx, assertionMode)
	})
}

func (e *Expectation) fetchOnce() ([]byte, bool, extension.PollingSummary) {
	executor := func(ctx context.Context) ([]byte, error) {
		return e.doSearch()
	}

	result, err, summary := retry.ExecuteSingle(context.Background(), executor)

	if err != nil {
		summary.Success = false
		summary.LastError = err.Error()
		return nil, false, summary
	}

	summary.Success = true
	return result, true, summary
}

func (e *Expectation) fetchWithRetry(stepCtx provider.StepCtx) ([]byte, bool, extension.PollingSummary) {
	asyncCfg := e.kafkaClient.GetAsyncConfig()

	executor := func(ctx context.Context) ([]byte, error) {
		return e.doSearch()
	}

	checker := func(result []byte, err error) []retry.CheckResult {
		if err != nil {
			return []retry.CheckResult{{
				Ok:        false,
				Retryable: true,
				Reason:    err.Error(),
			}}
		}

		return []retry.CheckResult{{
			Ok:        true,
			Retryable: false,
		}}
	}

	result, err, summary := retry.ExecuteWithRetry(
		context.Background(),
		stepCtx,
		asyncCfg,
		executor,
		checker,
	)

	if err != nil {
		return nil, false, summary
	}

	return result, true, summary
}

func (e *Expectation) doSearch() ([]byte, error) {
	// Проверяем что топик слушается
	if !e.kafkaClient.GetBuffer().IsTopicConfigured(e.topicName) {
		return nil, &kafkaErrors.KafkaTopicNotListenedError{
			TopicName:        e.topicName,
			MessageType:      "unknown",
			ConfiguredTopics: e.kafkaClient.GetBuffer().GetConfiguredTopics(),
		}
	}

	messages := e.kafkaClient.GetBuffer().GetMessages(e.topicName)

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

	// Ищем с конца (самые свежие)
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if e.matchesFilter(msg.Value) {
			return msg.Value, nil
		}
	}

	return nil, fmt.Errorf("message not found")
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

func (e *Expectation) checkUniqueness(stepCtx provider.StepCtx, mode extension.AssertionMode) {
	messages := e.kafkaClient.GetBuffer().GetMessages(e.topicName)
	_, err := e.findAndCountWithinWindow(messages)

	if err != nil {
		if notUniqueErr, ok := err.(*kafkaErrors.KafkaMessageNotUniqueError); ok {
			extension.NoError(stepCtx, mode, notUniqueErr, notUniqueErr.Error())
		}
	}
}

func (e *Expectation) runExpectations(stepCtx provider.StepCtx, mode extension.AssertionMode) {
	if len(e.expectations) == 0 {
		return
	}

	for _, exp := range e.expectations {
		e.checkExpectation(stepCtx, mode, exp)
	}
}

func (e *Expectation) checkExpectation(stepCtx provider.StepCtx, mode extension.AssertionMode, exp fieldExpectation) {
	result := gjson.GetBytes(e.messageBytes, exp.field)

	switch exp.checkType {
	case "equals":
		if !result.Exists() {
			extension.NoError(stepCtx, mode, fmt.Errorf("field '%s' not found", exp.field), "Field '%s' not found", exp.field)
			return
		}
		actualStr := result.String()
		expectedStr := fmt.Sprintf("%v", exp.value)
		if actualStr != expectedStr {
			extension.NoError(stepCtx, mode, fmt.Errorf("field '%s' expected '%v', got '%s'", exp.field, exp.value, actualStr),
				"Field '%s' expected '%v', got '%s'", exp.field, exp.value, actualStr)
		}

	case "notEmpty":
		if !result.Exists() || result.String() == "" {
			extension.NoError(stepCtx, mode, fmt.Errorf("field '%s' is empty", exp.field), "Field '%s' is empty", exp.field)
		}

	case "isNull":
		if result.Exists() && result.Type != gjson.Null {
			extension.NoError(stepCtx, mode, fmt.Errorf("field '%s' expected null, got '%s'", exp.field, result.String()),
				"Field '%s' expected null, got '%s'", exp.field, result.String())
		}

	case "isNotNull":
		if !result.Exists() || result.Type == gjson.Null {
			extension.NoError(stepCtx, mode, fmt.Errorf("field '%s' is null", exp.field), "Field '%s' is null", exp.field)
		}

	case "true":
		if !result.Exists() || !result.Bool() {
			extension.NoError(stepCtx, mode, fmt.Errorf("field '%s' expected true, got '%s'", exp.field, result.String()),
				"Field '%s' expected true, got '%s'", exp.field, result.String())
		}

	case "false":
		if !result.Exists() || result.Bool() {
			extension.NoError(stepCtx, mode, fmt.Errorf("field '%s' expected false, got '%s'", exp.field, result.String()),
				"Field '%s' expected false, got '%s'", exp.field, result.String())
		}
	}
}

func (e *Expectation) matchesFilter(jsonValue []byte) bool {
	if len(jsonValue) == 0 {
		return len(e.filters) == 0
	}

	if len(e.filters) == 0 {
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

		actualValue := result.String()
		if actualValue != expectedValue {
			return false
		}
	}

	return true
}

func (e *Expectation) buildFilterDescription() string {
	if len(e.filters) == 0 {
		return ""
	}

	parts := make([]string, 0, len(e.filters))
	for key, value := range e.filters {
		parts = append(parts, fmt.Sprintf("%s = %s", key, value))
	}

	return strings.Join(parts, ", ")
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
