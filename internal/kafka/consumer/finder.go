package consumer

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"

	"github.com/tidwall/gjson"

	kafkaErrors "github.com/gorelov-m-v/go-test-framework/internal/kafka/errors"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

type MessageFinder struct{}

func NewMessageFinder() *MessageFinder {
	return &MessageFinder{}
}

func (mf *MessageFinder) SearchAndDeserialize(
	messages []*types.KafkaMessage,
	filters map[string]string,
	targetType reflect.Type,
) (interface{}, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if mf.matchesFilter(msg.Value, filters) {
			result, err := mf.deserialize(msg, targetType)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	}

	return nil, nil
}

func (mf *MessageFinder) FindAndCount(
	messages []*types.KafkaMessage,
	filters map[string]string,
	targetType reflect.Type,
) (*types.FindResult[interface{}], error) {
	if len(messages) == 0 {
		return &types.FindResult[interface{}]{
			FirstMatch: nil,
			AllMatches: []interface{}{},
			Count:      0,
		}, nil
	}

	matches := make([]interface{}, 0)
	var firstMatch interface{}

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if mf.matchesFilter(msg.Value, filters) {
			result, err := mf.deserialize(msg, targetType)
			if err != nil {
				return nil, err
			}

			matches = append(matches, result)
			if firstMatch == nil {
				firstMatch = result
			}
		}
	}

	return &types.FindResult[interface{}]{
		FirstMatch: &firstMatch,
		AllMatches: matches,
		Count:      len(matches),
	}, nil
}

func (mf *MessageFinder) FindAndCountWithinWindow(
	messages []*types.KafkaMessage,
	filters map[string]string,
	targetType reflect.Type,
	windowMs int64,
) (*types.FindResult[interface{}], error) {
	if len(messages) == 0 {
		return &types.FindResult[interface{}]{
			FirstMatch: nil,
			AllMatches: []interface{}{},
			Count:      0,
		}, nil
	}

	matches := make([]interface{}, 0)
	var firstMatch interface{}
	var firstMatchTimestamp int64

	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if mf.matchesFilter(msg.Value, filters) {
			result, err := mf.deserialize(msg, targetType)
			if err != nil {
				return nil, err
			}

			if firstMatch == nil {
				firstMatch = result
				firstMatchTimestamp = msg.Timestamp
				matches = append(matches, result)
			} else {
				timeDiff := int64(math.Abs(float64(msg.Timestamp - firstMatchTimestamp)))
				if timeDiff <= windowMs {
					matches = append(matches, result)
				}
			}
		}
	}

	return &types.FindResult[interface{}]{
		FirstMatch: &firstMatch,
		AllMatches: matches,
		Count:      len(matches),
	}, nil
}

func (mf *MessageFinder) CountMatching(messages []*types.KafkaMessage, filters map[string]string) int {
	if len(messages) == 0 {
		return 0
	}

	count := 0
	for _, msg := range messages {
		if mf.matchesFilter(msg.Value, filters) {
			count++
		}
	}

	return count
}

func (mf *MessageFinder) matchesFilter(jsonValue []byte, filters map[string]string) bool {
	if len(jsonValue) == 0 {
		return len(filters) == 0
	}

	if len(filters) == 0 {
		return true
	}

	if !gjson.ValidBytes(jsonValue) {
		return false
	}

	for path, expectedValue := range filters {
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

func (mf *MessageFinder) deserialize(msg *types.KafkaMessage, targetType reflect.Type) (interface{}, error) {
	if msg == nil || len(msg.Value) == 0 {
		return nil, fmt.Errorf("message is empty")
	}

	resultPtr := reflect.New(targetType)
	result := resultPtr.Interface()

	if err := json.Unmarshal(msg.Value, result); err != nil {
		return nil, &kafkaErrors.KafkaDeserializationError{
			Topic:       msg.Topic,
			Offset:      msg.Offset,
			MessageType: targetType.Name(),
			Err:         err,
		}
	}

	return resultPtr.Elem().Interface(), nil
}
