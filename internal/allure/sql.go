package allure

import (
	"database/sql"
	"errors"
	"fmt"

	dbclient "github.com/gorelov-m-v/go-test-framework/pkg/database/client"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

type SQLReportDTO struct {
	Request SQLRequestDTO
	Result  SQLResultDTO
	Polling *PollingSummaryDTO
}

func (r *Reporter) AttachSQLReport(sCtx provider.StepCtx, db *dbclient.Client, report SQLReportDTO) {
	builder := NewReportBuilder()

	title := "SQL Query"
	if report.Result.Found {
		title = fmt.Sprintf("SQL Query → %d row(s)", report.Result.RowCount)
	} else if report.Result.Error != nil {
		if errors.Is(report.Result.Error, sql.ErrNoRows) {
			title = "SQL Query → No Rows"
		} else {
			title = "SQL Query → Error"
		}
	}
	builder.WriteHeader(title)

	r.writeSQLRequestSection(builder, report.Request)
	r.writeSQLResultSection(builder, db, report.Result)

	if report.Polling != nil && report.Polling.Attempts > 0 {
		r.writePollingSection(builder, report.Polling)
	}

	sCtx.WithNewAttachment("SQL Query", allure.Text, builder.Bytes())
}

func (r *Reporter) writeSQLRequestSection(builder *ReportBuilder, req SQLRequestDTO) {
	builder.WriteSectionHeader("QUERY")
	builder.WriteLine("%s", req.Query)

	if len(req.Args) > 0 {
		builder.WriteSection("Arguments")
		for i, arg := range req.Args {
			argStr := fmt.Sprintf("%v", arg)
			if strArg, ok := arg.(string); ok {
				if r.Config.ShouldMaskValue(strArg) {
					argStr = r.Config.MaskValue
				}
			}
			builder.WriteLine("  [%d] %s", i+1, argStr)
		}
	}
}

func (r *Reporter) writeSQLResultSection(builder *ReportBuilder, db *dbclient.Client, result SQLResultDTO) {
	status := "Found"
	if !result.Found {
		status = "Not Found"
	}
	if result.Error != nil && !errors.Is(result.Error, sql.ErrNoRows) {
		status = "Error"
	}
	builder.WriteSectionHeader(fmt.Sprintf("RESULT [%s]", status))

	if result.Duration > 0 {
		builder.WriteLine("Duration: %v", result.Duration)
	}

	if result.Error != nil {
		if errors.Is(result.Error, sql.ErrNoRows) {
			builder.WriteLine("Status: No rows found")
		} else {
			builder.WriteSection("Error")
			builder.WriteLine("%s", result.Error.Error())
		}
		return
	}

	if result.RowCount > 0 {
		builder.WriteLine("Rows: %d", result.RowCount)
	}

	if result.Data != nil {
		cleanData := CleanResult(result.Data)
		maskedData := maskSQLResult(db, cleanData)
		builder.WriteSection("Data")
		builder.WriteJSONOrError(maskedData)
	}
}

func maskSQLResult(db *dbclient.Client, data any) any {
	if db == nil || data == nil {
		return data
	}

	switch v := data.(type) {
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, value := range v {
			if db.ShouldMaskColumn(key) && value != nil {
				result[key] = MaskValue
			} else {
				result[key] = maskSQLResult(db, value)
			}
		}
		return result

	case []any:
		result := make([]any, len(v))
		for i, item := range v {
			result[i] = maskSQLResult(db, item)
		}
		return result

	default:
		return data
	}
}
