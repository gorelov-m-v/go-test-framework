package dsl

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

// ToDo: Вынести название колонок для маскировки результата в конфиг
func attachQuery(sCtx provider.StepCtx, sqlQuery string, args []any) {
	var sb strings.Builder
	sb.WriteString("SQL Query:\n")
	sb.WriteString(sqlQuery)
	sb.WriteString("\n\nArguments:\n")

	if len(args) == 0 {
		sb.WriteString("  (none)")
	} else {
		for i, arg := range args {
			argStr := fmt.Sprintf("%v", arg)
			if shouldMaskValue(arg) {
				argStr = "***MASKED***"
			}
			sb.WriteString(fmt.Sprintf("  [%d] %s\n", i+1, argStr))
		}
	}

	sCtx.WithNewAttachment("SQL Query", allure.Text, []byte(sb.String()))
}

func attachResult(sCtx provider.StepCtx, result any, err error) {
	if err != nil {
		// Distinguish ErrNoRows from other SQL errors for better UX
		if errors.Is(err, sql.ErrNoRows) {
			noRowsJSON, _ := json.MarshalIndent(map[string]string{
				"status": "no rows found",
			}, "", "  ")
			sCtx.WithNewAttachment("SQL Result", allure.JSON, noRowsJSON)
		} else {
			errJSON, _ := json.MarshalIndent(map[string]string{
				"error": err.Error(),
			}, "", "  ")
			sCtx.WithNewAttachment("SQL Error", allure.JSON, errJSON)
		}
		return
	}

	if result == nil {
		return
	}

	cleanResult := convertNullTypes(result)

	resultJSON, err := json.MarshalIndent(cleanResult, "", "  ")
	if err != nil {
		errJSON, _ := json.MarshalIndent(map[string]string{
			"marshal_error": err.Error(),
		}, "", "  ")
		sCtx.WithNewAttachment("SQL Result (Marshal Error)", allure.JSON, errJSON)
		return
	}

	sCtx.WithNewAttachment("SQL Result", allure.JSON, resultJSON)
}

func convertNullTypes(v any) any {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		if isHandled, result := handleNullType(val); isHandled {
			return result
		}

		result := make(map[string]any)
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}

			fieldValue := val.Field(i)

			dbTag := field.Tag.Get("db")
			if dbTag == "-" {
				continue
			}

			fieldName := field.Name
			if dbTag != "" {
				parts := strings.Split(dbTag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				}
			} else {
				jsonTag := field.Tag.Get("json")
				if jsonTag == "-" {
					continue
				}
				if jsonTag != "" {
					parts := strings.Split(jsonTag, ",")
					if parts[0] != "" {
						fieldName = parts[0]
					}
				}
			}

			result[fieldName] = convertNullTypes(fieldValue.Interface())
		}
		return result

	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return val.Interface()
		}

		result := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = convertNullTypes(val.Index(i).Interface())
		}
		return result

	case reflect.Map:
		result := make(map[string]any)
		iter := val.MapRange()
		for iter.Next() {
			key := fmt.Sprintf("%v", iter.Key().Interface())
			result[key] = convertNullTypes(iter.Value().Interface())
		}
		return result

	default:
		return v
	}
}

func handleNullType(val reflect.Value) (bool, any) {
	typ := val.Type()

	switch typ {
	case reflect.TypeOf(sql.NullString{}):
		ns := val.Interface().(sql.NullString)
		if ns.Valid {
			return true, ns.String
		}
		return true, nil

	case reflect.TypeOf(sql.NullInt64{}):
		ni := val.Interface().(sql.NullInt64)
		if ni.Valid {
			return true, ni.Int64
		}
		return true, nil

	case reflect.TypeOf(sql.NullInt32{}):
		ni := val.Interface().(sql.NullInt32)
		if ni.Valid {
			return true, ni.Int32
		}
		return true, nil

	case reflect.TypeOf(sql.NullFloat64{}):
		nf := val.Interface().(sql.NullFloat64)
		if nf.Valid {
			return true, nf.Float64
		}
		return true, nil

	case reflect.TypeOf(sql.NullBool{}):
		nb := val.Interface().(sql.NullBool)
		if nb.Valid {
			return true, nb.Bool
		}
		return true, nil

	case reflect.TypeOf(sql.NullTime{}):
		nt := val.Interface().(sql.NullTime)
		if nt.Valid {
			return true, nt.Time
		}
		return true, nil
	}

	return false, nil
}

func attachExecResult(sCtx provider.StepCtx, res sql.Result, err error) {
	if err != nil {
		errJSON, _ := json.MarshalIndent(map[string]string{
			"error": err.Error(),
		}, "", "  ")
		sCtx.WithNewAttachment("SQL Exec Error", allure.JSON, errJSON)
		return
	}

	if res == nil {
		return
	}

	rowsAffected, _ := res.RowsAffected()
	lastInsertId, _ := res.LastInsertId()

	execResult := map[string]int64{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}

	resultJSON, err := json.MarshalIndent(execResult, "", "  ")
	if err != nil {
		return
	}

	sCtx.WithNewAttachment("SQL Exec Result", allure.JSON, resultJSON)
}

func shouldMaskValue(arg any) bool {
	if str, ok := arg.(string); ok {
		lower := strings.ToLower(str)
		if strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") {
			return true
		}
	}
	return false
}
