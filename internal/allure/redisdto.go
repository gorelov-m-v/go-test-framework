package allure

import (
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

type RedisRequestDTO struct {
	Server string
	Key    string
}

func ToRedisRequestDTO(server, key string) RedisRequestDTO {
	return RedisRequestDTO{
		Server: server,
		Key:    key,
	}
}

type RedisResultDTO struct {
	Key      string
	Exists   bool
	Value    string
	TTL      time.Duration
	Duration time.Duration
	Error    error
}

func ToRedisResultDTO(result *client.Result) RedisResultDTO {
	if result == nil {
		return RedisResultDTO{}
	}
	return RedisResultDTO{
		Key:      result.Key,
		Exists:   result.Exists,
		Value:    result.Value,
		TTL:      result.TTL,
		Duration: result.Duration,
		Error:    result.Error,
	}
}
