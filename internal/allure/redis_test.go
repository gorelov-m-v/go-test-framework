package allure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWriteRedisTTL(t *testing.T) {
	tests := []struct {
		name     string
		ttl      time.Duration
		expected string
	}{
		{
			name:     "positive TTL",
			ttl:      5 * time.Minute,
			expected: "TTL: 5m0s",
		},
		{
			name:     "no expiration",
			ttl:      -1,
			expected: "TTL: no expiration",
		},
		{
			name:     "key does not exist",
			ttl:      -2,
			expected: "TTL: key does not exist",
		},
		{
			name:     "zero TTL",
			ttl:      0,
			expected: "TTL: 0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewDefaultReporter()
			builder := NewReportBuilder()

			reporter.writeRedisTTL(builder, tt.ttl)

			assert.Contains(t, builder.String(), tt.expected)
		})
	}
}

func TestWriteRedisValue(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		contains    []string
		notContains []string
	}{
		{
			name:  "empty value",
			value: "",
		},
		{
			name:  "simple value",
			value: "hello world",
			contains: []string{
				"Value:",
				"hello world",
			},
		},
		{
			name:  "JSON value",
			value: `{"name": "John", "age": 30}`,
			contains: []string{
				"Value:",
				`{"name": "John", "age": 30}`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewDefaultReporter()
			builder := NewReportBuilder()

			reporter.writeRedisValue(builder, tt.value)

			content := builder.String()

			if tt.value == "" {
				assert.Empty(t, content)
			} else {
				for _, expected := range tt.contains {
					assert.Contains(t, content, expected)
				}
			}
		})
	}
}

func TestWriteRedisValue_Truncation(t *testing.T) {
	reporter := NewDefaultReporter()
	builder := NewReportBuilder()

	longValue := make([]byte, 2000)
	for i := range longValue {
		longValue[i] = 'x'
	}

	reporter.writeRedisValue(builder, string(longValue))

	content := builder.String()
	assert.Contains(t, content, "truncated")
	assert.Contains(t, content, "2000 bytes total")
}

func TestRedisRequestDTO(t *testing.T) {
	dto := RedisRequestDTO{
		Server: "localhost:6379",
		Key:    "user:123",
	}

	assert.Equal(t, "localhost:6379", dto.Server)
	assert.Equal(t, "user:123", dto.Key)
}

func TestToRedisRequestDTO(t *testing.T) {
	dto := ToRedisRequestDTO("redis-cluster:6380", "session:abc")

	assert.Equal(t, "redis-cluster:6380", dto.Server)
	assert.Equal(t, "session:abc", dto.Key)
}

func TestRedisResultDTO(t *testing.T) {
	dto := RedisResultDTO{
		Key:      "user:123",
		Exists:   true,
		Value:    `{"name": "John"}`,
		TTL:      5 * time.Minute,
		Duration: 10 * time.Millisecond,
		Error:    nil,
	}

	assert.Equal(t, "user:123", dto.Key)
	assert.True(t, dto.Exists)
	assert.Equal(t, `{"name": "John"}`, dto.Value)
	assert.Equal(t, 5*time.Minute, dto.TTL)
	assert.Equal(t, 10*time.Millisecond, dto.Duration)
	assert.Nil(t, dto.Error)
}
