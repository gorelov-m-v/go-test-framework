package allure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestKafkaSearchDTO(t *testing.T) {
	dto := KafkaSearchDTO{
		Topic:   "user-events",
		Filters: map[string]string{"userId": "123", "eventType": "CREATED"},
		Timeout: 30 * time.Second,
		Unique:  true,
	}

	assert.Equal(t, "user-events", dto.Topic)
	assert.Equal(t, "123", dto.Filters["userId"])
	assert.Equal(t, "CREATED", dto.Filters["eventType"])
	assert.Equal(t, 30*time.Second, dto.Timeout)
	assert.True(t, dto.Unique)
}

func TestToKafkaSearchDTO(t *testing.T) {
	filters := map[string]string{"userId": "456", "eventType": "UPDATED"}
	dto := ToKafkaSearchDTO("order-events", filters, 15*time.Second, false)

	assert.Equal(t, "order-events", dto.Topic)
	assert.Equal(t, filters, dto.Filters)
	assert.Equal(t, 15*time.Second, dto.Timeout)
	assert.False(t, dto.Unique)
}

func TestKafkaResultDTO(t *testing.T) {
	dto := KafkaResultDTO{
		Found:      true,
		Message:    map[string]string{"id": "123"},
		RawMessage: []byte(`{"id": "123"}`),
		MatchCount: 1,
	}

	assert.True(t, dto.Found)
	assert.NotNil(t, dto.Message)
	assert.NotEmpty(t, dto.RawMessage)
	assert.Equal(t, 1, dto.MatchCount)
}

func TestReportBuilder_WriteMap(t *testing.T) {
	builder := NewReportBuilder()

	filters := map[string]string{
		"userId":    "123",
		"eventType": "CREATED",
	}

	builder.WriteMap(filters)

	content := builder.String()
	assert.Contains(t, content, "userId: 123")
	assert.Contains(t, content, "eventType: CREATED")
}

func TestReportBuilder_WriteSection(t *testing.T) {
	builder := NewReportBuilder()

	builder.WriteSection("Filters")
	builder.WriteKeyValue("key", "value")

	content := builder.String()
	assert.Contains(t, content, "Filters:")
	assert.Contains(t, content, "key: value")
}
