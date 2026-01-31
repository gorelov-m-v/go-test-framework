package allure

import "time"

type KafkaSearchDTO struct {
	Topic   string
	Filters map[string]string
	Timeout time.Duration
	Unique  bool
}

type KafkaResultDTO struct {
	Found      bool
	Message    any
	RawMessage []byte
	MatchCount int
}
