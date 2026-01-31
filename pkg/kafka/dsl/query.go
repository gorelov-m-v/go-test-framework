package dsl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	kafkaErrors "github.com/gorelov-m-v/go-test-framework/internal/kafka/errors"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/validation"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/topic"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

// Query represents a Kafka message consumer with filtering and expectations.
// It searches for messages in the client's buffer that match specified filters
// and validates them against expectations.
//
// Type parameter T should implement topic.TopicName interface for typed consumption.
//
// Example:
//
//	dsl.Consume[topics.UserEvents](sCtx, kafkaClient).
//	    With("userId", userID).
//	    ExpectFieldEquals("eventType", "USER_CREATED").
//	    Send()
type Query[T any] struct {
	sCtx   provider.StepCtx
	client *client.Client
	ctx    context.Context

	topicName string

	filters         map[string]string
	containsFilters map[string]string
	unique          bool
	duplicateWindow time.Duration
	expectedCount   int

	result *Result[T]
	sent   bool

	expectations    []*expect.Expectation[[]byte]
	allMatchingMsgs [][]byte
	messageBytes    []byte
	found           bool
	lastError       error
}

// Result represents the outcome of a Kafka message search.
//
// Fields:
//   - Found: Whether a matching message was found
//   - Message: Deserialized message of type T
//   - RawMessage: Raw message bytes
//   - AllMessages: All matching messages (when using ExpectCount)
//   - MatchCount: Number of matching messages
//   - ParseError: Error if message could not be deserialized to T
type Result[T any] struct {
	Found       bool
	Message     T
	RawMessage  []byte
	AllMessages [][]byte
	MatchCount  int
	ParseError  error
}

// NewQuery creates a new Kafka query builder for the specified topic.
//
// Parameters:
//   - sCtx: Allure step context for test reporting
//   - kafkaClient: Kafka client with message buffer
//   - topicName: Full topic name to search in
//
// Prefer using Consume[T] for typed topic consumption.
func NewQuery[T any](sCtx provider.StepCtx, kafkaClient *client.Client, topicName string) *Query[T] {
	return &Query[T]{
		sCtx:            sCtx,
		client:          kafkaClient,
		ctx:             context.Background(),
		topicName:       topicName,
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
		expectations:    make([]*expect.Expectation[[]byte], 0),
	}
}

// Consume creates a typed Kafka query for messages from a topic.
// The topic name is automatically derived from the TTopic type's TopicName() method.
//
// Type parameter TTopic must implement topic.TopicName interface.
//
// Example:
//
//	dsl.Consume[topics.PlayerEvents](sCtx, kafkaClient).
//	    With("playerId", playerID).
//	    With("eventType", "PLAYER_CREATED").
//	    ExpectFieldEquals("playerName", "John").
//	    Send()
func Consume[TTopic topic.TopicName](sCtx provider.StepCtx, kafkaClient *client.Client) *Query[TTopic] {
	var topicName TTopic
	fullTopicName := kafkaClient.GetTopicPrefix() + topicName.TopicName()
	return NewQuery[TTopic](sCtx, kafkaClient, fullTopicName)
}

// Context sets a custom context for the query operation.
func (q *Query[T]) Context(ctx context.Context) *Query[T] {
	q.ctx = ctx
	return q
}

func (q *Query[T]) validate() {
	v := validation.New(q.sCtx, "Kafka")
	v.RequireNotNil(q.client, "Kafka client")
	v.RequireNotEmptyWithHint(q.topicName, "Topic name", "Use Consume[TopicType]() or NewQuery().")
}

// Send executes the Kafka message search and validates all expectations.
// In async mode (AsyncStep), automatically retries with backoff until a matching message is found.
// Returns the Result containing the found message and metadata.
func (q *Query[T]) Send() *Result[T] {
	q.validate()

	q.sCtx.WithNewStep(q.stepName(), func(stepCtx provider.StepCtx) {
		var summary polling.PollingSummary
		var err error
		q.messageBytes, q.found, err, summary = q.execute(stepCtx)
		q.lastError = err

		attachKafkaReport(stepCtx, q, summary)

		if !q.found {
			q.handleNotFound(stepCtx, summary)
			return
		}

		q.assertResults(stepCtx, err)
		q.result = q.buildResult()
		q.sent = true
	})

	return q.result
}

func (q *Query[T]) stepName() string {
	return fmt.Sprintf("Kafka: Consume from '%s'", q.topicName)
}

func (q *Query[T]) handleNotFound(stepCtx provider.StepCtx, summary polling.PollingSummary) {
	mode := polling.GetStepMode(stepCtx)
	assertionMode := polling.GetAssertionModeFromStepMode(mode)

	msg := fmt.Sprintf("Kafka message in topic '%s' not found within %s. Filters: %v",
		q.topicName, q.client.GetDefaultTimeout(), q.filters)

	if mode == polling.AsyncMode {
		msg = polling.FinalFailureMessage(summary)
	}

	polling.NoError(stepCtx, assertionMode, fmt.Errorf("%s", msg), msg)
	q.result = &Result[T]{Found: false}
}

func (q *Query[T]) assertResults(stepCtx provider.StepCtx, err error) {
	mode := polling.GetStepMode(stepCtx)
	assertionMode := polling.GetAssertionModeFromStepMode(mode)

	if q.expectedCount > 0 {
		q.checkExpectedCount(stepCtx, assertionMode)
	}

	if q.unique {
		q.checkUniqueness(stepCtx, assertionMode)
	}

	q.runExpectations(stepCtx, err)
}

func (q *Query[T]) buildResult() *Result[T] {
	result := &Result[T]{
		Found:       q.found,
		RawMessage:  q.messageBytes,
		AllMessages: q.allMatchingMsgs,
		MatchCount:  len(q.allMatchingMsgs),
	}

	if q.found && len(q.messageBytes) > 0 {
		var msg T
		if err := json.Unmarshal(q.messageBytes, &msg); err != nil {
			result.ParseError = fmt.Errorf("failed to parse message to %T: %w", msg, err)
		} else {
			result.Message = msg
		}
	}

	if result.MatchCount == 0 && q.found {
		result.MatchCount = 1
	}

	return result
}

func (q *Query[T]) doSearch() ([]byte, error) {
	if !q.client.GetBuffer().IsTopicConfigured(q.topicName) {
		return nil, &kafkaErrors.KafkaTopicNotListenedError{
			TopicName:        q.topicName,
			MessageType:      "unknown",
			ConfiguredTopics: q.client.GetBuffer().GetConfiguredTopics(),
		}
	}

	messages := q.client.GetBuffer().GetMessages(q.topicName)

	if q.expectedCount > 0 {
		allMatching, err := q.findAllMatching(messages)
		if err != nil {
			return nil, err
		}
		q.allMatchingMsgs = allMatching
		if len(allMatching) < q.expectedCount {
			return nil, fmt.Errorf("expected %d messages, found %d", q.expectedCount, len(allMatching))
		}
		return allMatching[0], nil
	}

	if q.unique {
		msgBytes, err := q.findAndCountWithinWindow(messages)
		return msgBytes, err
	}

	msgBytes, err := q.searchMessage(messages)
	return msgBytes, err
}

func (q *Query[T]) searchMessage(messages []*types.KafkaMessage) ([]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages in buffer")
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if q.matchesFilter(msg.Value) {
			return msg.Value, nil
		}
	}

	return nil, fmt.Errorf("message not found")
}

func (q *Query[T]) findAllMatching(messages []*types.KafkaMessage) ([][]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages in buffer")
	}

	var result [][]byte
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if q.matchesFilter(msg.Value) {
			result = append(result, msg.Value)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("message not found")
	}

	return result, nil
}

func (q *Query[T]) findAndCountWithinWindow(messages []*types.KafkaMessage) ([]byte, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages in buffer")
	}

	var firstMatchBytes []byte
	var firstMatchTimestamp int64
	count := 0

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if q.matchesFilter(msg.Value) {
			if count == 0 {
				firstMatchBytes = msg.Value
				firstMatchTimestamp = msg.Timestamp
				count++
			} else {
				timeDiff := abs(msg.Timestamp - firstMatchTimestamp)
				if timeDiff <= q.duplicateWindow.Milliseconds() {
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
			MessageType: q.topicName,
			Filters:     q.filters,
			Count:       count,
			WindowMs:    q.duplicateWindow.Milliseconds(),
		}
	}

	return firstMatchBytes, nil
}

func (q *Query[T]) checkUniqueness(stepCtx provider.StepCtx, mode polling.AssertionMode) {
	messages := q.client.GetBuffer().GetMessages(q.topicName)
	_, err := q.findAndCountWithinWindow(messages)

	if err != nil {
		if notUniqueErr, ok := err.(*kafkaErrors.KafkaMessageNotUniqueError); ok {
			polling.NoError(stepCtx, mode, notUniqueErr, notUniqueErr.Error())
		}
	}
}

func (q *Query[T]) checkExpectedCount(stepCtx provider.StepCtx, mode polling.AssertionMode) {
	actualCount := len(q.allMatchingMsgs)
	if actualCount != q.expectedCount {
		msg := fmt.Sprintf("Expected %d Kafka messages, but found %d. Topic: %s, Filters: %v",
			q.expectedCount, actualCount, q.topicName, q.filters)
		polling.NoError(stepCtx, mode, fmt.Errorf("%s", msg), msg)
	}
}

func (q *Query[T]) runExpectations(stepCtx provider.StepCtx, err error) {
	expect.AssertExpectations(stepCtx, q.expectations, err, q.messageBytes, nil)
}
