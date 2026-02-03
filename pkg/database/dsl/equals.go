package dsl

import (
	"fmt"
	"reflect"

	"github.com/gorelov-m-v/go-test-framework/internal/typeconv"
)

func equalsLoose(expected, actual any) (bool, bool, string) {
	if expected == nil && actual == nil {
		return true, true, ""
	}
	if expected == nil || actual == nil {
		if typeconv.IsNull(actual) && expected == nil {
			return true, true, ""
		}
		return false, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	if typeconv.IsNull(actual) {
		return false, true, "column is NULL yet"
	}

	expBool, expIsBool := typeconv.ToBool(expected)
	actBool, actIsBool := typeconv.ToBool(actual)
	if expIsBool && actIsBool {
		if expBool == actBool {
			return true, true, ""
		}
		return false, true, fmt.Sprintf("expected %v, got %v", expBool, actBool)
	}

	expNum, expIsNum := typeconv.ToNumber(expected)
	actNum, actIsNum := typeconv.ToNumber(actual)
	if expIsNum && actIsNum {
		equal := expNum == actNum
		return equal, true, fmt.Sprintf("expected %v, got %v", expNum, actNum)
	}

	if expIsNum != actIsNum {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	expStr, expIsStr := typeconv.ToString(expected)
	actStr, actIsStr := typeconv.ToString(actual)
	if expIsStr && actIsStr {
		equal := expStr == actStr
		return equal, true, fmt.Sprintf("expected %v, got %v", expStr, actStr)
	}

	if expIsStr != actIsStr {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	if expVal.Type() != actVal.Type() {
		return false, false, fmt.Sprintf("type mismatch: expected %T(%v), got %T(%v) - incompatible types", expected, expected, actual, actual)
	}

	if !expVal.Type().Comparable() {
		equal := reflect.DeepEqual(expected, actual)
		return equal, true, fmt.Sprintf("expected %v, got %v", expected, actual)
	}

	equal := expected == actual
	return equal, true, fmt.Sprintf("expected %v, got %v", expected, actual)
}
