package jsonutil

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode"

	"github.com/tidwall/gjson"
)

type CompareMode int

const (
	ModePartial CompareMode = iota
	ModeExact
)

func IsEmpty(res gjson.Result) bool {
	if !res.Exists() {
		return true
	}
	if res.Type == gjson.Null {
		return true
	}

	switch res.Type {
	case gjson.String:
		return strings.TrimSpace(res.String()) == ""
	case gjson.JSON:
		if res.IsArray() {
			return len(res.Array()) == 0
		}
		if res.IsObject() {
			return len(res.Map()) == 0
		}
		return false
	default:
		return false
	}
}

func TypeToString(t gjson.Type) string {
	switch t {
	case gjson.Null:
		return "null"
	case gjson.False, gjson.True:
		return "boolean"
	case gjson.Number:
		return "number"
	case gjson.String:
		return "string"
	case gjson.JSON:
		return "object/array"
	default:
		return "unknown"
	}
}

func DebugValue(res gjson.Result) string {
	if res.Raw != "" {
		return res.Raw
	}
	return fmt.Sprintf("%v", res.Value())
}

func ToInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	default:
		return 0, false
	}
}

func ToUint64(v any) (uint64, bool) {
	switch val := v.(type) {
	case uint:
		return uint64(val), true
	case uint8:
		return uint64(val), true
	case uint16:
		return uint64(val), true
	case uint32:
		return uint64(val), true
	case uint64:
		return val, true
	default:
		return 0, false
	}
}

func ToFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

func Compare(res gjson.Result, expected any) (bool, string) {
	val := reflect.ValueOf(expected)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			if !res.Exists() {
				return false, "field does not exist (expected null)"
			}
			if res.Type != gjson.Null {
				return false, fmt.Sprintf("expected null, got %s: %s", TypeToString(res.Type), DebugValue(res))
			}
			return true, ""
		}
		return Compare(res, val.Elem().Interface())
	}

	if expected == nil {
		if !res.Exists() {
			return false, "field does not exist (expected null)"
		}
		if res.Type != gjson.Null {
			return false, fmt.Sprintf("expected null, got %s: %s", TypeToString(res.Type), DebugValue(res))
		}
		return true, ""
	}

	switch exp := expected.(type) {
	case string:
		if res.Type != gjson.String {
			return false, fmt.Sprintf("expected string %q, got %s: %s", exp, TypeToString(res.Type), DebugValue(res))
		}
		actual := res.String()
		if actual != exp {
			return false, fmt.Sprintf("expected %q, got %q", exp, actual)
		}
		return true, ""

	case bool:
		if res.Type != gjson.True && res.Type != gjson.False {
			return false, fmt.Sprintf("expected boolean %v, got %s: %s", exp, TypeToString(res.Type), DebugValue(res))
		}
		actual := res.Bool()
		if actual != exp {
			return false, fmt.Sprintf("expected %v, got %v", exp, actual)
		}
		return true, ""

	case []string:
		if !res.IsArray() {
			return false, fmt.Sprintf("expected array, got %s: %s", TypeToString(res.Type), DebugValue(res))
		}
		actualArr := res.Array()
		if len(actualArr) != len(exp) {
			return false, fmt.Sprintf("expected array of length %d, got %d", len(exp), len(actualArr))
		}
		actualSet := make(map[string]bool, len(actualArr))
		for _, item := range actualArr {
			actualSet[item.String()] = true
		}
		for _, expItem := range exp {
			if !actualSet[expItem] {
				return false, fmt.Sprintf("missing element %q in array", expItem)
			}
		}
		return true, ""

	default:
		expVal := reflect.ValueOf(expected)
		if expVal.Kind() == reflect.Map {
			if !res.IsObject() {
				return false, fmt.Sprintf("expected object, got %s: %s", TypeToString(res.Type), DebugValue(res))
			}
			return compareMap(res, expVal)
		}
		if expectedInt, ok := ToInt64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedInt, TypeToString(res.Type), DebugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedInt, actualFloat)
			}
			actualInt := res.Int()
			if actualInt != expectedInt {
				return false, fmt.Sprintf("expected %d, got %d", expectedInt, actualInt)
			}
			return true, ""
		}

		if expectedUint, ok := ToUint64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %d, got %s: %s", expectedUint, TypeToString(res.Type), DebugValue(res))
			}
			actualFloat := res.Float()
			if math.Trunc(actualFloat) != actualFloat {
				return false, fmt.Sprintf("expected integer %d, got float %v", expectedUint, actualFloat)
			}
			actualUint := res.Uint()
			if actualUint != expectedUint {
				return false, fmt.Sprintf("expected %d, got %d", expectedUint, actualUint)
			}
			return true, ""
		}

		if expectedFloat, ok := ToFloat64(expected); ok {
			if res.Type != gjson.Number {
				return false, fmt.Sprintf("expected number %v, got %s: %s", expectedFloat, TypeToString(res.Type), DebugValue(res))
			}
			actualFloat := res.Float()
			if actualFloat != expectedFloat {
				return false, fmt.Sprintf("expected %v, got %v", expectedFloat, actualFloat)
			}
			return true, ""
		}

		return false, fmt.Sprintf("unsupported expected type %T; supported: string/bool/int*/uint*/float*/nil", expected)
	}
}

func toJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	name := field.Name
	if len(name) == 0 {
		return name
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.IsNil() || v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		return false
	default:
		return false
	}
}

func CompareObjectPartial(jsonObj gjson.Result, expected any) (bool, string) {
	return compareObject(jsonObj, expected, ModePartial)
}

func CompareObjectExact(jsonObj gjson.Result, expected any) (bool, string) {
	return compareObject(jsonObj, expected, ModeExact)
}

func compareObject(jsonObj gjson.Result, expected any, mode CompareMode) (bool, string) {
	if expected == nil {
		return true, ""
	}

	val := reflect.ValueOf(expected)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return true, ""
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Map {
		return compareMap(jsonObj, val)
	}

	if val.Kind() != reflect.Struct {
		return false, fmt.Sprintf("expected struct, got %T", expected)
	}

	if !jsonObj.IsObject() {
		return false, fmt.Sprintf("expected JSON object, got %s", TypeToString(jsonObj.Type))
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if !field.IsExported() {
			continue
		}

		jsonFieldName := toJSONFieldName(field)
		jsonField := jsonObj.Get(jsonFieldName)

		if mode == ModePartial {
			if isZeroValue(fieldVal) {
				continue
			}
			if !jsonField.Exists() {
				return false, fmt.Sprintf("field '%s' not found in JSON", jsonFieldName)
			}
		} else {
			if !jsonField.Exists() {
				continue
			}
		}

		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				if mode == ModePartial {
					continue
				}
				if jsonField.Exists() && jsonField.Type != gjson.Null {
					return false, fmt.Sprintf("field '%s': expected null, got %s: %s", jsonFieldName, TypeToString(jsonField.Type), DebugValue(jsonField))
				}
				continue
			}
			derefVal := fieldVal.Elem()
			if derefVal.Kind() == reflect.Struct {
				ok, msg := compareObject(jsonField, derefVal.Interface(), mode)
				if !ok {
					return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
				}
				continue
			}
			fieldVal = derefVal
		}

		if fieldVal.Kind() == reflect.Struct {
			ok, msg := compareObject(jsonField, fieldVal.Interface(), mode)
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		if fieldVal.Kind() == reflect.Map {
			ok, msg := compareMap(jsonField, fieldVal)
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		if fieldVal.Kind() == reflect.Slice {
			ok, msg := compareSliceWithMode(jsonField, fieldVal, mode)
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		ok, msg := Compare(jsonField, fieldVal.Interface())
		if !ok {
			return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
		}
	}

	return true, ""
}

func compareMap(jsonObj gjson.Result, mapVal reflect.Value) (bool, string) {
	if !jsonObj.IsObject() {
		return false, fmt.Sprintf("expected JSON object for map, got %s", TypeToString(jsonObj.Type))
	}

	for _, key := range mapVal.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		jsonField := jsonObj.Get(keyStr)

		if !jsonField.Exists() {
			return false, fmt.Sprintf("key '%s' not found", keyStr)
		}

		expectedVal := mapVal.MapIndex(key).Interface()
		ok, msg := Compare(jsonField, expectedVal)
		if !ok {
			return false, fmt.Sprintf("key '%s': %s", keyStr, msg)
		}
	}

	return true, ""
}

func compareSlice(jsonArr gjson.Result, sliceVal reflect.Value) (bool, string) {
	return compareSliceWithMode(jsonArr, sliceVal, ModePartial)
}

func compareSliceWithMode(jsonArr gjson.Result, sliceVal reflect.Value, mode CompareMode) (bool, string) {
	if !jsonArr.IsArray() {
		return false, fmt.Sprintf("expected JSON array, got %s", TypeToString(jsonArr.Type))
	}

	jsonItems := jsonArr.Array()
	if len(jsonItems) != sliceVal.Len() {
		return false, fmt.Sprintf("array length mismatch: expected %d, got %d", sliceVal.Len(), len(jsonItems))
	}

	for i := 0; i < sliceVal.Len(); i++ {
		expectedItem := sliceVal.Index(i).Interface()
		jsonItem := jsonItems[i]

		itemVal := reflect.ValueOf(expectedItem)
		if itemVal.Kind() == reflect.Struct {
			ok, msg := compareObject(jsonItem, expectedItem, mode)
			if !ok {
				return false, fmt.Sprintf("index %d: %s", i, msg)
			}
		} else {
			ok, msg := Compare(jsonItem, expectedItem)
			if !ok {
				return false, fmt.Sprintf("index %d: %s", i, msg)
			}
		}
	}

	return true, ""
}

func FindInArray(jsonArr gjson.Result, expected any) (int, gjson.Result) {
	if !jsonArr.IsArray() {
		return -1, gjson.Result{}
	}

	for i, item := range jsonArr.Array() {
		ok, _ := CompareObjectPartial(item, expected)
		if ok {
			return i, item
		}
	}

	return -1, gjson.Result{}
}

func FindInArrayExact(jsonArr gjson.Result, expected any) (int, gjson.Result) {
	if !jsonArr.IsArray() {
		return -1, gjson.Result{}
	}

	for i, item := range jsonArr.Array() {
		ok, _ := CompareObjectExact(item, expected)
		if ok {
			return i, item
		}
	}

	return -1, gjson.Result{}
}
