package jsonutil

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"unicode"

	"github.com/tidwall/gjson"
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
	// Handle pointers - dereference if not nil
	val := reflect.ValueOf(expected)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			// nil pointer means we expect null in JSON
			if !res.Exists() {
				return false, "field does not exist (expected null)"
			}
			if res.Type != gjson.Null {
				return false, fmt.Sprintf("expected null, got %s: %s", TypeToString(res.Type), DebugValue(res))
			}
			return true, ""
		}
		// Dereference the pointer and compare with the actual value
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

	default:
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

// toJSONFieldName converts Go struct field name to JSON field name (camelCase).
// It uses the `json` tag if present, otherwise converts PascalCase to camelCase.
func toJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag != "" && jsonTag != "-" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] != "" {
			return parts[0]
		}
	}
	// Convert PascalCase to camelCase
	name := field.Name
	if len(name) == 0 {
		return name
	}
	runes := []rune(name)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
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
		return false // Structs are never considered zero for partial match
	default:
		return false
	}
}

// CompareObjectPartial compares a JSON object with an expected struct using partial matching.
// Only non-zero fields in the expected struct are compared.
// Returns (true, "") if all non-zero fields match, or (false, reason) if not.
func CompareObjectPartial(jsonObj gjson.Result, expected any) (bool, string) {
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

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip zero values (partial match)
		if isZeroValue(fieldVal) {
			continue
		}

		jsonFieldName := toJSONFieldName(field)
		jsonField := jsonObj.Get(jsonFieldName)

		if !jsonField.Exists() {
			return false, fmt.Sprintf("field '%s' not found in JSON", jsonFieldName)
		}

		// Handle pointers to structs
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				// nil pointer - already skipped by isZeroValue, but check just in case
				continue
			}
			// Dereference and handle as struct
			derefVal := fieldVal.Elem()
			if derefVal.Kind() == reflect.Struct {
				ok, msg := CompareObjectPartial(jsonField, derefVal.Interface())
				if !ok {
					return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
				}
				continue
			}
			// For non-struct pointers, use the dereferenced value
			fieldVal = derefVal
		}

		// Handle nested structs
		if fieldVal.Kind() == reflect.Struct {
			ok, msg := CompareObjectPartial(jsonField, fieldVal.Interface())
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		// Handle maps (e.g., map[string]string for localized names)
		if fieldVal.Kind() == reflect.Map {
			ok, msg := compareMap(jsonField, fieldVal)
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		// Handle slices
		if fieldVal.Kind() == reflect.Slice {
			ok, msg := compareSlice(jsonField, fieldVal)
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		// Handle primitive types
		ok, msg := Compare(jsonField, fieldVal.Interface())
		if !ok {
			return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
		}
	}

	return true, ""
}

// compareMap compares a JSON object with an expected map.
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

// compareSlice compares a JSON array with an expected slice.
func compareSlice(jsonArr gjson.Result, sliceVal reflect.Value) (bool, string) {
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

		// Check if it's a struct for partial matching
		itemVal := reflect.ValueOf(expectedItem)
		if itemVal.Kind() == reflect.Struct {
			ok, msg := CompareObjectPartial(jsonItem, expectedItem)
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

// FindInArray searches for an object in a JSON array that matches the expected struct (partial match).
// Returns the index of the first matching element, or -1 if not found.
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

// CompareObjectExact compares a JSON object with an expected struct using exact matching.
// ALL fields in the expected struct are compared, including zero values.
// Fields that don't exist in JSON are skipped (handles contract mismatches).
// Returns (true, "") if all fields match, or (false, reason) if not.
func CompareObjectExact(jsonObj gjson.Result, expected any) (bool, string) {
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

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonFieldName := toJSONFieldName(field)
		jsonField := jsonObj.Get(jsonFieldName)

		// Skip fields that don't exist in JSON (contract mismatch)
		if !jsonField.Exists() {
			continue
		}

		// Handle pointers to structs
		if fieldVal.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				// nil pointer - check if JSON field is null or doesn't exist
				if jsonField.Exists() && jsonField.Type != gjson.Null {
					return false, fmt.Sprintf("field '%s': expected null, got %s: %s", jsonFieldName, TypeToString(jsonField.Type), DebugValue(jsonField))
				}
				continue
			}
			// Dereference and handle as struct
			derefVal := fieldVal.Elem()
			if derefVal.Kind() == reflect.Struct {
				ok, msg := CompareObjectExact(jsonField, derefVal.Interface())
				if !ok {
					return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
				}
				continue
			}
			// For non-struct pointers, use the dereferenced value
			fieldVal = derefVal
		}

		// Handle nested structs
		if fieldVal.Kind() == reflect.Struct {
			ok, msg := CompareObjectExact(jsonField, fieldVal.Interface())
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		// Handle maps (e.g., map[string]string for localized names)
		if fieldVal.Kind() == reflect.Map {
			ok, msg := compareMap(jsonField, fieldVal)
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		// Handle slices
		if fieldVal.Kind() == reflect.Slice {
			ok, msg := compareSlice(jsonField, fieldVal)
			if !ok {
				return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
			}
			continue
		}

		// Handle primitive types
		ok, msg := Compare(jsonField, fieldVal.Interface())
		if !ok {
			return false, fmt.Sprintf("field '%s': %s", jsonFieldName, msg)
		}
	}

	return true, ""
}

// FindInArrayExact searches for an object in a JSON array that matches the expected struct (exact match).
// Returns the index of the first matching element, or -1 if not found.
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
