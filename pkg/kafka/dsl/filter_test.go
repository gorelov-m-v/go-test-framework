package dsl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestFormatFilterValue_String(t *testing.T) {
	result := formatFilterValue("hello")
	assert.Equal(t, "hello", result)
}

func TestFormatFilterValue_Int(t *testing.T) {
	result := formatFilterValue(123)
	assert.Equal(t, "123", result)
}

func TestFormatFilterValue_Int64(t *testing.T) {
	result := formatFilterValue(int64(9876543210))
	assert.Equal(t, "9876543210", result)
}

func TestFormatFilterValue_Float(t *testing.T) {
	result := formatFilterValue(3.14)
	assert.Equal(t, "3.14", result)
}

func TestFormatFilterValue_Bool(t *testing.T) {
	assert.Equal(t, "true", formatFilterValue(true))
	assert.Equal(t, "false", formatFilterValue(false))
}

func TestFormatFilterValue_StringSlice(t *testing.T) {
	result := formatFilterValue([]string{"a", "b", "c"})
	assert.Equal(t, `["a","b","c"]`, result)
}

func TestFormatFilterValue_EmptyStringSlice(t *testing.T) {
	result := formatFilterValue([]string{})
	assert.Equal(t, `[]`, result)
}

func TestFormatFilterValue_Nil(t *testing.T) {
	result := formatFilterValue(nil)
	assert.Equal(t, "", result)
}

func TestCompareArraysAsSet_Equal(t *testing.T) {
	jsonData := `["a", "b", "c"]`
	result := gjson.Parse(jsonData)
	expectedJSON := `["a","b","c"]`

	assert.True(t, compareArraysAsSet(result, expectedJSON))
}

func TestCompareArraysAsSet_DifferentOrder(t *testing.T) {
	jsonData := `["c", "a", "b"]`
	result := gjson.Parse(jsonData)
	expectedJSON := `["a","b","c"]`

	assert.True(t, compareArraysAsSet(result, expectedJSON))
}

func TestCompareArraysAsSet_DifferentLength(t *testing.T) {
	jsonData := `["a", "b"]`
	result := gjson.Parse(jsonData)
	expectedJSON := `["a","b","c"]`

	assert.False(t, compareArraysAsSet(result, expectedJSON))
}

func TestCompareArraysAsSet_DifferentValues(t *testing.T) {
	jsonData := `["a", "b", "d"]`
	result := gjson.Parse(jsonData)
	expectedJSON := `["a","b","c"]`

	assert.False(t, compareArraysAsSet(result, expectedJSON))
}

func TestCompareArraysAsSet_WithDuplicates(t *testing.T) {
	jsonData := `["a", "a", "b"]`
	result := gjson.Parse(jsonData)
	expectedJSON := `["a","a","b"]`

	assert.True(t, compareArraysAsSet(result, expectedJSON))
}

func TestCompareArraysAsSet_DuplicatesMismatch(t *testing.T) {
	jsonData := `["a", "a", "b"]`
	result := gjson.Parse(jsonData)
	expectedJSON := `["a","b","b"]`

	assert.False(t, compareArraysAsSet(result, expectedJSON))
}

func TestCompareArraysAsSet_EmptyArrays(t *testing.T) {
	jsonData := `[]`
	result := gjson.Parse(jsonData)
	expectedJSON := `[]`

	assert.True(t, compareArraysAsSet(result, expectedJSON))
}

func TestCompareArraysAsSet_InvalidExpectedJSON(t *testing.T) {
	jsonData := `["a", "b"]`
	result := gjson.Parse(jsonData)
	expectedJSON := `not valid json`

	assert.False(t, compareArraysAsSet(result, expectedJSON))
}

func TestAbs_PositiveNumber(t *testing.T) {
	assert.Equal(t, int64(5), abs(5))
}

func TestAbs_NegativeNumber(t *testing.T) {
	assert.Equal(t, int64(5), abs(-5))
}

func TestAbs_Zero(t *testing.T) {
	assert.Equal(t, int64(0), abs(0))
}

func TestAbs_MaxInt64(t *testing.T) {
	assert.Equal(t, int64(9223372036854775807), abs(9223372036854775807))
}

func TestMatchesFilter_EmptyFiltersEmptyJSON(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	assert.True(t, q.matchesFilter([]byte{}))
}

func TestMatchesFilter_EmptyFiltersWithJSON(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"id": 123, "name": "test"}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ExactMatch(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"id": "123"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"id": 123, "name": "test"}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ExactMatchString(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"name": "test"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"id": 123, "name": "test"}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_NoMatch(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"id": "456"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"id": 123, "name": "test"}`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_NestedPath(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"user.id": "123"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"user": {"id": 123, "name": "test"}}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_DeepNestedPath(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"data.user.profile.id": "abc"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"data": {"user": {"profile": {"id": "abc"}}}}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_PathNotExists(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"nonexistent": "value"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"id": 123}`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_MultipleFilters(t *testing.T) {
	q := &Query[any]{
		filters: map[string]string{
			"id":   "123",
			"name": "test",
		},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"id": 123, "name": "test", "extra": "field"}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_MultipleFiltersPartialMatch(t *testing.T) {
	q := &Query[any]{
		filters: map[string]string{
			"id":   "123",
			"name": "wrong",
		},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"id": 123, "name": "test"}`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ArrayFilter(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"tags": `["a","b","c"]`},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"tags": ["a", "b", "c"]}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ArrayFilterDifferentOrder(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"tags": `["a","b","c"]`},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"tags": ["c", "a", "b"]}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_InvalidJSON(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"id": "123"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`not valid json`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ContainsFilter(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: map[string]string{"tags": "b"},
	}

	jsonData := []byte(`{"tags": ["a", "b", "c"]}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ContainsFilterNotFound(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: map[string]string{"tags": "d"},
	}

	jsonData := []byte(`{"tags": ["a", "b", "c"]}`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ContainsFilterNotArray(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: map[string]string{"name": "test"},
	}

	jsonData := []byte(`{"name": "test"}`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ContainsFilterPathNotExists(t *testing.T) {
	q := &Query[any]{
		filters:         make(map[string]string),
		containsFilters: map[string]string{"nonexistent": "value"},
	}

	jsonData := []byte(`{"id": 123}`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_MixedFilters(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"id": "123"},
		containsFilters: map[string]string{"tags": "important"},
	}

	jsonData := []byte(`{"id": 123, "tags": ["important", "urgent"]}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_MixedFiltersPartialFail(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"id": "123"},
		containsFilters: map[string]string{"tags": "missing"},
	}

	jsonData := []byte(`{"id": 123, "tags": ["important", "urgent"]}`)
	assert.False(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_ArrayIndex(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"items.0": "first"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"items": ["first", "second", "third"]}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_NestedArrayAccess(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"users.0.name": "Alice"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"users": [{"name": "Alice"}, {"name": "Bob"}]}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_BooleanValue(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"active": "true"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"active": true}`)
	assert.True(t, q.matchesFilter(jsonData))
}

func TestMatchesFilter_NullValue(t *testing.T) {
	q := &Query[any]{
		filters:         map[string]string{"value": "null"},
		containsFilters: make(map[string]string),
	}

	jsonData := []byte(`{"value": null}`)
	assert.False(t, q.matchesFilter(jsonData))
}
