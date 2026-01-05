package allure

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

func CleanResult(v any) any {
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

			fieldName := getFieldName(field)
			if fieldName == "" {
				continue
			}

			result[fieldName] = CleanResult(fieldValue.Interface())
		}
		return result

	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 {
			return val.Interface()
		}

		result := make([]any, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = CleanResult(val.Index(i).Interface())
		}
		return result

	case reflect.Map:
		result := make(map[string]any)
		iter := val.MapRange()
		for iter.Next() {
			key := fmt.Sprintf("%v", iter.Key().Interface())
			result[key] = CleanResult(iter.Value().Interface())
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

func getFieldName(field reflect.StructField) string {
	dbTag := field.Tag.Get("db")
	if dbTag == "-" {
		return "" // Игнорируем поле
	}
	if dbTag != "" {
		parts := strings.Split(dbTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return "" // Игнорируем поле
	}
	if jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}

	return field.Name
}

func MarshalJSONIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

func FormatJSON(data []byte) ([]byte, error) {
	var jsonData any
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return nil, err
	}
	return MarshalJSONIndent(jsonData)
}
