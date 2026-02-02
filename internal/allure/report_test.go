package allure

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

func TestToPollingSummaryDTO(t *testing.T) {
	t.Run("zero attempts returns nil", func(t *testing.T) {
		ps := polling.PollingSummary{Attempts: 0}
		dto := ToPollingSummaryDTO(ps)

		assert.Nil(t, dto)
	})

	t.Run("with attempts", func(t *testing.T) {
		ps := polling.PollingSummary{
			Attempts:     3,
			ElapsedTime:  "1.5s",
			Success:      true,
			LastError:    "",
			FailedChecks: nil,
		}
		dto := ToPollingSummaryDTO(ps)

		assert.NotNil(t, dto)
		assert.Equal(t, 3, dto.Attempts)
		assert.Equal(t, "1.5s", dto.ElapsedTime)
		assert.True(t, dto.Success)
		assert.Empty(t, dto.LastError)
	})

	t.Run("with failed checks", func(t *testing.T) {
		ps := polling.PollingSummary{
			Attempts:     5,
			ElapsedTime:  "10s",
			Success:      false,
			LastError:    "timeout",
			FailedChecks: []string{"check1 failed", "check2 failed"},
		}
		dto := ToPollingSummaryDTO(ps)

		assert.NotNil(t, dto)
		assert.Equal(t, 5, dto.Attempts)
		assert.False(t, dto.Success)
		assert.Equal(t, "timeout", dto.LastError)
		assert.Len(t, dto.FailedChecks, 2)
	})
}
