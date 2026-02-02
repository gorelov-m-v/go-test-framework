package allure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReportBuilder_Bytes(t *testing.T) {
	builder := NewReportBuilder()
	builder.WriteLine("test")

	assert.Equal(t, []byte("test\n"), builder.Bytes())
}

func TestReportBuilder_WriteHeader(t *testing.T) {
	builder := NewReportBuilder()
	builder.WriteHeader("Test Title")

	content := builder.String()
	assert.Contains(t, content, "═══════════════════════════════════════════════════════════════")
	assert.Contains(t, content, "Test Title")
}

func TestReportBuilder_WriteSectionHeader(t *testing.T) {
	builder := NewReportBuilder()
	builder.WriteSectionHeader("Section Name")

	content := builder.String()
	assert.Contains(t, content, "───────────────────────────────────────────────────────────────")
	assert.Contains(t, content, "Section Name")
}

func TestReportBuilder_WriteJSON(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		builder := NewReportBuilder()
		err := builder.WriteJSON(map[string]string{"key": "value"})

		assert.NoError(t, err)
		assert.Contains(t, builder.String(), `"key": "value"`)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		builder := NewReportBuilder()
		err := builder.WriteJSON(make(chan int))

		assert.Error(t, err)
	})
}

func TestReportBuilder_WriteJSONOrError(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		builder := NewReportBuilder()
		builder.WriteJSONOrError(map[string]int{"count": 42})

		assert.Contains(t, builder.String(), `"count": 42`)
	})

	t.Run("invalid JSON shows error", func(t *testing.T) {
		builder := NewReportBuilder()
		builder.WriteJSONOrError(make(chan int))

		assert.Contains(t, builder.String(), "failed to marshal")
	})
}

func TestReportBuilder_WriteTruncated(t *testing.T) {
	t.Run("small data not truncated", func(t *testing.T) {
		builder := NewReportBuilder()
		builder.WriteTruncated([]byte("small"), 100)

		content := builder.String()
		assert.Equal(t, "small\n", content)
		assert.NotContains(t, content, "truncated")
	})

	t.Run("large data truncated", func(t *testing.T) {
		builder := NewReportBuilder()
		data := make([]byte, 100)
		for i := range data {
			data[i] = 'x'
		}
		builder.WriteTruncated(data, 10)

		content := builder.String()
		assert.Contains(t, content, "truncated")
		assert.Contains(t, content, "100 bytes total")
	})
}
