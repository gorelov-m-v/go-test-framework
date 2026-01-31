package allure

import "time"

type RedisRequestDTO struct {
	Server string
	Key    string
}

type RedisResultDTO struct {
	Key      string
	Exists   bool
	Value    string
	TTL      time.Duration
	Duration time.Duration
	Error    error
}
