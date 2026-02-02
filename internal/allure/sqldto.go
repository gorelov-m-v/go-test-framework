package allure

import (
	"database/sql"
	"errors"
	"time"
)

type SQLRequestDTO struct {
	Query string
	Args  []any
}

type SQLResultDTO struct {
	Found    bool
	RowCount int
	Data     any
	Duration time.Duration
	Error    error
}

type SQLResultParams struct {
	Result   any
	RowCount int
	Duration time.Duration
	Error    error
}

type SQLAttachParams struct {
	Query    string
	Args     []any
	Result   any
	RowCount int
	Duration time.Duration
	Error    error
}

func ToSQLRequestDTO(query string, args []any) SQLRequestDTO {
	return SQLRequestDTO{
		Query: query,
		Args:  args,
	}
}

func ToSQLResultDTO(p SQLResultParams) SQLResultDTO {
	dto := SQLResultDTO{
		Duration: p.Duration,
		Error:    p.Error,
	}

	if p.Error == nil {
		dto.Found = true
		dto.RowCount = p.RowCount
		dto.Data = p.Result
	} else if errors.Is(p.Error, sql.ErrNoRows) {
		dto.Found = false
	}

	return dto
}
