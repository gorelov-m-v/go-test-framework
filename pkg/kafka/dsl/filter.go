package dsl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tidwall/gjson"
)

// With adds a filter to match messages where the JSON field at key equals value.
// Multiple With calls use AND logic. Supports GJSON path syntax.
func (q *Query[T]) With(key string, value interface{}) *Query[T] {
	if value != nil {
		q.filters[key] = formatFilterValue(value)
	}
	return q
}

// WithContains adds a filter to match messages where the JSON array at key contains value.
func (q *Query[T]) WithContains(key string, value interface{}) *Query[T] {
	if value != nil {
		q.containsFilters[key] = fmt.Sprintf("%v", value)
	}
	return q
}

// Unique ensures only one matching message exists within the default window.
// Test fails if duplicates are found.
func (q *Query[T]) Unique() *Query[T] {
	q.unique = true
	q.duplicateWindow = q.client.GetUniqueWindow()
	return q
}

// UniqueWithWindow ensures only one matching message exists within the specified time window.
func (q *Query[T]) UniqueWithWindow(window time.Duration) *Query[T] {
	q.unique = true
	q.duplicateWindow = window
	return q
}

// ExpectCount expects exactly count messages matching the filters.
func (q *Query[T]) ExpectCount(count int) *Query[T] {
	q.expectedCount = count
	return q
}

func (q *Query[T]) matchesFilter(jsonValue []byte) bool {
	if len(jsonValue) == 0 {
		return len(q.filters) == 0 && len(q.containsFilters) == 0
	}

	if len(q.filters) == 0 && len(q.containsFilters) == 0 {
		return true
	}

	if !gjson.ValidBytes(jsonValue) {
		return false
	}

	for path, expectedValue := range q.filters {
		result := gjson.GetBytes(jsonValue, path)

		if !result.Exists() {
			return false
		}

		if result.IsArray() {
			if !compareArraysAsSet(result, expectedValue) {
				return false
			}
		} else {
			actualValue := result.String()
			if actualValue != expectedValue {
				return false
			}
		}
	}

	for path, expectedValue := range q.containsFilters {
		result := gjson.GetBytes(jsonValue, path)

		if !result.Exists() {
			return false
		}

		if !result.IsArray() {
			return false
		}

		found := false
		for _, item := range result.Array() {
			if item.String() == expectedValue {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func formatFilterValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case []string:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", value)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func compareArraysAsSet(result gjson.Result, expectedJSON string) bool {
	var expectedArr []string
	if err := json.Unmarshal([]byte(expectedJSON), &expectedArr); err != nil {
		return false
	}

	actualArr := result.Array()

	if len(actualArr) != len(expectedArr) {
		return false
	}

	expectedCounts := make(map[string]int)
	for _, v := range expectedArr {
		expectedCounts[v]++
	}

	actualCounts := make(map[string]int)
	for _, item := range actualArr {
		actualCounts[item.String()]++
	}

	for key, expectedCount := range expectedCounts {
		if actualCounts[key] != expectedCount {
			return false
		}
	}

	return true
}
