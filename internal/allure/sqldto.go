package allure

import "time"

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

type SQLReportDTO struct {
	Request SQLRequestDTO
	Result  SQLResultDTO
	Polling *PollingSummaryDTO
}
