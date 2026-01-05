package allure

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	dbclient "go-test-framework/pkg/database/client"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func (r *Reporter) AttachSQLQuery(sCtx provider.StepCtx, query string, args []any) {
	builder := NewReportBuilder()
	builder.WriteLine("SQL Query:")
	builder.WriteLine("%s", query)
	builder.WriteSection("Arguments")

	if len(args) == 0 {
		builder.WriteLine("  (none)")
	} else {
		for i, arg := range args {
			argStr := fmt.Sprintf("%v", arg)

			if strArg, ok := arg.(string); ok {
				if r.Config.ShouldMaskValue(strArg) {
					argStr = r.Config.MaskValue
				}
			}

			builder.WriteLine("  [%d] %s", i+1, argStr)
		}
	}

	sCtx.WithNewAttachment("SQL Query", allure.Text, builder.Bytes())
}

func (r *Reporter) AttachSQLResult(sCtx provider.StepCtx, db *dbclient.Client, result any, err error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.attachJSON(sCtx, "SQL Result", map[string]string{
				"status": "no rows found",
			})
		} else {
			r.attachJSON(sCtx, "SQL Error", map[string]string{
				"error": err.Error(),
			})
		}
		return
	}

	if result == nil {
		return
	}

	cleanData := CleanResult(result)
	maskedData := maskSQLResult(db, cleanData)
	r.attachJSON(sCtx, "SQL Result", maskedData)
}

func (r *Reporter) AttachSQLExecResult(sCtx provider.StepCtx, res sql.Result, err error) {
	if err != nil {
		r.attachJSON(sCtx, "SQL Exec Error", map[string]string{
			"error": err.Error(),
		})
		return
	}

	if res == nil {
		return
	}

	rowsAffected, _ := res.RowsAffected()
	lastInsertId, _ := res.LastInsertId()

	r.attachJSON(sCtx, "SQL Exec Result", map[string]int64{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	})
}

func (r *Reporter) attachJSON(sCtx provider.StepCtx, name string, content any) {
	bytes, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		errJSON, _ := json.MarshalIndent(map[string]string{
			"marshal_error": err.Error(),
		}, "", "  ")
		sCtx.WithNewAttachment(name+" (Marshal Error)", allure.JSON, errJSON)
		return
	}
	sCtx.WithNewAttachment(name, allure.JSON, bytes)
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
