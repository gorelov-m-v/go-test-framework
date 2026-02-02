package allure

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToSQLRequestDTO(t *testing.T) {
	dto := ToSQLRequestDTO("SELECT * FROM users WHERE id = $1", []any{123})

	assert.Equal(t, "SELECT * FROM users WHERE id = $1", dto.Query)
	assert.Equal(t, []any{123}, dto.Args)
}

func TestToSQLResultDTO(t *testing.T) {
	t.Run("success with data", func(t *testing.T) {
		data := map[string]any{"id": 1, "name": "test"}
		dto := ToSQLResultDTO(SQLResultParams{
			Result:   data,
			RowCount: 1,
			Duration: 50 * time.Millisecond,
			Error:    nil,
		})

		assert.True(t, dto.Found)
		assert.Equal(t, 1, dto.RowCount)
		assert.Equal(t, data, dto.Data)
		assert.Equal(t, 50*time.Millisecond, dto.Duration)
		assert.Nil(t, dto.Error)
	})

	t.Run("not found (sql.ErrNoRows)", func(t *testing.T) {
		dto := ToSQLResultDTO(SQLResultParams{
			Result:   nil,
			RowCount: 0,
			Duration: 10 * time.Millisecond,
			Error:    sql.ErrNoRows,
		})

		assert.False(t, dto.Found)
		assert.Equal(t, 0, dto.RowCount)
		assert.Nil(t, dto.Data)
		assert.Equal(t, sql.ErrNoRows, dto.Error)
	})

	t.Run("other error", func(t *testing.T) {
		customErr := errors.New("connection refused")
		dto := ToSQLResultDTO(SQLResultParams{
			Result:   nil,
			RowCount: 0,
			Duration: 5 * time.Millisecond,
			Error:    customErr,
		})

		assert.False(t, dto.Found)
		assert.Equal(t, 0, dto.RowCount)
		assert.Nil(t, dto.Data)
		assert.Equal(t, customErr, dto.Error)
	})
}
