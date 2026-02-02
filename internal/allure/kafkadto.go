package allure

import (
	"encoding/json"
	"time"
)

type KafkaSearchDTO struct {
	Topic   string
	Filters map[string]string
	Timeout time.Duration
	Unique  bool
}

func ToKafkaSearchDTO(topic string, filters map[string]string, timeout time.Duration, unique bool) KafkaSearchDTO {
	return KafkaSearchDTO{
		Topic:   topic,
		Filters: filters,
		Timeout: timeout,
		Unique:  unique,
	}
}

type KafkaResultDTO struct {
	Found      bool
	Message    any
	RawMessage []byte
	MatchCount int
}

type KafkaResultParams struct {
	Found           bool
	MessageBytes    []byte
	AllMatchingMsgs [][]byte
	ExpectedCount   int
}

func ToKafkaResultDTO(p KafkaResultParams) KafkaResultDTO {
	if !p.Found {
		return KafkaResultDTO{
			Found: false,
		}
	}

	if p.ExpectedCount > 0 && len(p.AllMatchingMsgs) > 0 {
		return KafkaResultDTO{
			Found:      true,
			MatchCount: len(p.AllMatchingMsgs),
			Message:    parseMessagesToAny(p.AllMatchingMsgs),
		}
	}

	var msgAny any
	if err := json.Unmarshal(p.MessageBytes, &msgAny); err != nil {
		msgAny = string(p.MessageBytes)
	}

	return KafkaResultDTO{
		Found:      true,
		MatchCount: 1,
		Message:    msgAny,
		RawMessage: p.MessageBytes,
	}
}

func parseMessagesToAny(messages [][]byte) any {
	if len(messages) == 0 {
		return nil
	}

	var parsed []any
	for _, msgBytes := range messages {
		var msgMap map[string]interface{}
		if err := json.Unmarshal(msgBytes, &msgMap); err == nil {
			parsed = append(parsed, msgMap)
		}
	}
	return parsed
}
