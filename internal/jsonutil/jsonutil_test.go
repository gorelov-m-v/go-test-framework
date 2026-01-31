package jsonutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		path   string
		expect bool
	}{
		{"non-existent field", `{"a": 1}`, "b", true},
		{"null value", `{"a": null}`, "a", true},
		{"empty string", `{"a": ""}`, "a", true},
		{"whitespace string", `{"a": "   "}`, "a", true},
		{"non-empty string", `{"a": "hello"}`, "a", false},
		{"empty array", `{"a": []}`, "a", true},
		{"non-empty array", `{"a": [1,2]}`, "a", false},
		{"empty object", `{"a": {}}`, "a", true},
		{"non-empty object", `{"a": {"b": 1}}`, "a", false},
		{"number zero", `{"a": 0}`, "a", false},
		{"number non-zero", `{"a": 42}`, "a", false},
		{"boolean false", `{"a": false}`, "a", false},
		{"boolean true", `{"a": true}`, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Get(tt.json, tt.path)
			got := IsEmpty(res)
			if got != tt.expect {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestTypeToString(t *testing.T) {
	tests := []struct {
		typ    gjson.Type
		expect string
	}{
		{gjson.Null, "null"},
		{gjson.False, "boolean"},
		{gjson.True, "boolean"},
		{gjson.Number, "number"},
		{gjson.String, "string"},
		{gjson.JSON, "object/array"},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			got := TypeToString(tt.typ)
			if got != tt.expect {
				t.Errorf("TypeToString(%v) = %v, want %v", tt.typ, got, tt.expect)
			}
		})
	}
}

func TestDebugValue(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		expect string
	}{
		{"string", `"hello"`, `"hello"`},
		{"number", `42`, `42`},
		{"null", `null`, `null`},
		{"bool", `true`, `true`},
		{"object", `{"a":1}`, `{"a":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			got := DebugValue(res)
			if got != tt.expect {
				t.Errorf("DebugValue() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		expect int64
		ok     bool
	}{
		{"int", int(42), 42, true},
		{"int8", int8(42), 42, true},
		{"int16", int16(42), 42, true},
		{"int32", int32(42), 42, true},
		{"int64", int64(42), 42, true},
		{"negative int", int(-10), -10, true},
		{"string", "42", 0, false},
		{"float64", float64(42), 0, false},
		{"uint", uint(42), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ToInt64(tt.input)
			if ok != tt.ok {
				t.Errorf("ToInt64() ok = %v, want %v", ok, tt.ok)
			}
			if ok && got != tt.expect {
				t.Errorf("ToInt64() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestToUint64(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		expect uint64
		ok     bool
	}{
		{"uint", uint(42), 42, true},
		{"uint8", uint8(42), 42, true},
		{"uint16", uint16(42), 42, true},
		{"uint32", uint32(42), 42, true},
		{"uint64", uint64(42), 42, true},
		{"string", "42", 0, false},
		{"int", int(42), 0, false},
		{"float64", float64(42), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ToUint64(tt.input)
			if ok != tt.ok {
				t.Errorf("ToUint64() ok = %v, want %v", ok, tt.ok)
			}
			if ok && got != tt.expect {
				t.Errorf("ToUint64() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		expect float64
		ok     bool
	}{
		{"float32", float32(3.14), float64(float32(3.14)), true},
		{"float64", float64(3.14159), 3.14159, true},
		{"string", "3.14", 0, false},
		{"int", int(42), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ToFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("ToFloat64() ok = %v, want %v", ok, tt.ok)
			}
			if ok && got != tt.expect {
				t.Errorf("ToFloat64() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	t.Run("string match", func(t *testing.T) {
		res := gjson.Parse(`"hello"`)
		ok, _ := Compare(res, "hello")
		assert.True(t, ok)
	})

	t.Run("string mismatch", func(t *testing.T) {
		res := gjson.Parse(`"hello"`)
		ok, msg := Compare(res, "world")
		assert.False(t, ok)
		assert.Contains(t, msg, "hello")
	})

	t.Run("string type mismatch", func(t *testing.T) {
		res := gjson.Parse(`42`)
		ok, msg := Compare(res, "hello")
		assert.False(t, ok)
		assert.Contains(t, msg, "string")
	})

	t.Run("bool true match", func(t *testing.T) {
		res := gjson.Parse(`true`)
		ok, _ := Compare(res, true)
		assert.True(t, ok)
	})

	t.Run("bool false match", func(t *testing.T) {
		res := gjson.Parse(`false`)
		ok, _ := Compare(res, false)
		assert.True(t, ok)
	})

	t.Run("bool mismatch", func(t *testing.T) {
		res := gjson.Parse(`true`)
		ok, _ := Compare(res, false)
		assert.False(t, ok)
	})

	t.Run("bool type mismatch", func(t *testing.T) {
		res := gjson.Parse(`"true"`)
		ok, msg := Compare(res, true)
		assert.False(t, ok)
		assert.Contains(t, msg, "boolean")
	})

	t.Run("int match", func(t *testing.T) {
		res := gjson.Parse(`42`)
		ok, _ := Compare(res, 42)
		assert.True(t, ok)
	})

	t.Run("int mismatch", func(t *testing.T) {
		res := gjson.Parse(`42`)
		ok, _ := Compare(res, 100)
		assert.False(t, ok)
	})

	t.Run("int type mismatch", func(t *testing.T) {
		res := gjson.Parse(`"42"`)
		ok, msg := Compare(res, 42)
		assert.False(t, ok)
		assert.Contains(t, msg, "number")
	})

	t.Run("int vs float mismatch", func(t *testing.T) {
		res := gjson.Parse(`42.5`)
		ok, msg := Compare(res, 42)
		assert.False(t, ok)
		assert.Contains(t, msg, "float")
	})

	t.Run("uint match", func(t *testing.T) {
		res := gjson.Parse(`42`)
		ok, _ := Compare(res, uint(42))
		assert.True(t, ok)
	})

	t.Run("uint mismatch", func(t *testing.T) {
		res := gjson.Parse(`42`)
		ok, _ := Compare(res, uint(100))
		assert.False(t, ok)
	})

	t.Run("float64 match", func(t *testing.T) {
		res := gjson.Parse(`3.14`)
		ok, _ := Compare(res, 3.14)
		assert.True(t, ok)
	})

	t.Run("float64 mismatch", func(t *testing.T) {
		res := gjson.Parse(`3.14`)
		ok, _ := Compare(res, 2.71)
		assert.False(t, ok)
	})

	t.Run("nil match null", func(t *testing.T) {
		res := gjson.Parse(`null`)
		ok, _ := Compare(res, nil)
		assert.True(t, ok)
	})

	t.Run("nil vs non-null", func(t *testing.T) {
		res := gjson.Parse(`"hello"`)
		ok, msg := Compare(res, nil)
		assert.False(t, ok)
		assert.Contains(t, msg, "null")
	})

	t.Run("nil pointer match null", func(t *testing.T) {
		res := gjson.Parse(`null`)
		var p *string
		ok, _ := Compare(res, p)
		assert.True(t, ok)
	})

	t.Run("nil pointer vs non-null", func(t *testing.T) {
		res := gjson.Parse(`"hello"`)
		var p *string
		ok, msg := Compare(res, p)
		assert.False(t, ok)
		assert.Contains(t, msg, "null")
	})

	t.Run("non-nil pointer dereferenced", func(t *testing.T) {
		res := gjson.Parse(`"hello"`)
		s := "hello"
		ok, _ := Compare(res, &s)
		assert.True(t, ok)
	})

	t.Run("[]string match", func(t *testing.T) {
		res := gjson.Parse(`["a", "b", "c"]`)
		ok, _ := Compare(res, []string{"a", "b", "c"})
		assert.True(t, ok)
	})

	t.Run("[]string wrong length", func(t *testing.T) {
		res := gjson.Parse(`["a", "b"]`)
		ok, msg := Compare(res, []string{"a", "b", "c"})
		assert.False(t, ok)
		assert.Contains(t, msg, "length")
	})

	t.Run("[]string missing element", func(t *testing.T) {
		res := gjson.Parse(`["a", "b", "d"]`)
		ok, msg := Compare(res, []string{"a", "b", "c"})
		assert.False(t, ok)
		assert.Contains(t, msg, "missing")
	})

	t.Run("[]string not array", func(t *testing.T) {
		res := gjson.Parse(`"hello"`)
		ok, msg := Compare(res, []string{"a"})
		assert.False(t, ok)
		assert.Contains(t, msg, "array")
	})

	t.Run("map match", func(t *testing.T) {
		res := gjson.Parse(`{"a": 1, "b": 2}`)
		ok, _ := Compare(res, map[string]int{"a": 1, "b": 2})
		assert.True(t, ok)
	})

	t.Run("map not object", func(t *testing.T) {
		res := gjson.Parse(`[1, 2]`)
		ok, msg := Compare(res, map[string]int{"a": 1})
		assert.False(t, ok)
		assert.Contains(t, msg, "object")
	})

	t.Run("unsupported type", func(t *testing.T) {
		res := gjson.Parse(`42`)
		ok, msg := Compare(res, struct{ X int }{X: 42})
		assert.False(t, ok)
		assert.Contains(t, msg, "unsupported")
	})

	t.Run("nil expected field not exists", func(t *testing.T) {
		res := gjson.Get(`{"a": 1}`, "b")
		ok, msg := Compare(res, nil)
		assert.False(t, ok)
		assert.Contains(t, msg, "does not exist")
	})
}

func TestIsZeroValue(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		expect bool
	}{
		{"empty array", [0]int{}, true},
		{"non-empty array", [1]int{1}, false},
		{"nil slice", ([]int)(nil), true},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1}, false},
		{"nil map", (map[string]int)(nil), true},
		{"empty map", map[string]int{}, true},
		{"non-empty map", map[string]int{"a": 1}, false},
		{"bool false", false, true},
		{"bool true", true, false},
		{"int zero", int(0), true},
		{"int non-zero", int(42), false},
		{"int8 zero", int8(0), true},
		{"int16 zero", int16(0), true},
		{"int32 zero", int32(0), true},
		{"int64 zero", int64(0), true},
		{"uint zero", uint(0), true},
		{"uint non-zero", uint(42), false},
		{"uint8 zero", uint8(0), true},
		{"uint16 zero", uint16(0), true},
		{"uint32 zero", uint32(0), true},
		{"uint64 zero", uint64(0), true},
		{"float32 zero", float32(0), true},
		{"float32 non-zero", float32(3.14), false},
		{"float64 zero", float64(0), true},
		{"float64 non-zero", float64(3.14), false},
		{"string empty", "", true},
		{"string non-empty", "hello", false},
		{"nil pointer", (*int)(nil), true},
		{"non-nil pointer", new(int), false},
		{"struct", struct{ X int }{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.value)
			got := isZeroValue(v)
			if got != tt.expect {
				t.Errorf("isZeroValue(%v) = %v, want %v", tt.value, got, tt.expect)
			}
		})
	}
}

func TestCompareSlice(t *testing.T) {
	t.Run("match simple slice", func(t *testing.T) {
		jsonArr := gjson.Parse(`[1, 2, 3]`)
		ok, _ := compareSlice(jsonArr, reflect.ValueOf([]int{1, 2, 3}))
		assert.True(t, ok)
	})

	t.Run("length mismatch", func(t *testing.T) {
		jsonArr := gjson.Parse(`[1, 2]`)
		ok, msg := compareSlice(jsonArr, reflect.ValueOf([]int{1, 2, 3}))
		assert.False(t, ok)
		assert.Contains(t, msg, "length")
	})

	t.Run("not array", func(t *testing.T) {
		jsonArr := gjson.Parse(`"hello"`)
		ok, msg := compareSlice(jsonArr, reflect.ValueOf([]int{1}))
		assert.False(t, ok)
		assert.Contains(t, msg, "array")
	})

	t.Run("element mismatch", func(t *testing.T) {
		jsonArr := gjson.Parse(`[1, 5, 3]`)
		ok, msg := compareSlice(jsonArr, reflect.ValueOf([]int{1, 2, 3}))
		assert.False(t, ok)
		assert.Contains(t, msg, "index")
	})

	t.Run("slice of structs", func(t *testing.T) {
		type Item struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
		}
		jsonArr := gjson.Parse(`[{"id": 1, "name": "a"}, {"id": 2, "name": "b"}]`)
		ok, _ := compareSlice(jsonArr, reflect.ValueOf([]Item{{Id: 1, Name: "a"}, {Id: 2, Name: "b"}}))
		assert.True(t, ok)
	})
}

func TestCompareMap(t *testing.T) {
	t.Run("match map", func(t *testing.T) {
		jsonObj := gjson.Parse(`{"a": 1, "b": 2}`)
		ok, _ := compareMap(jsonObj, reflect.ValueOf(map[string]int{"a": 1, "b": 2}))
		assert.True(t, ok)
	})

	t.Run("key not found", func(t *testing.T) {
		jsonObj := gjson.Parse(`{"a": 1}`)
		ok, msg := compareMap(jsonObj, reflect.ValueOf(map[string]int{"a": 1, "c": 3}))
		assert.False(t, ok)
		assert.Contains(t, msg, "not found")
	})

	t.Run("value mismatch", func(t *testing.T) {
		jsonObj := gjson.Parse(`{"a": 1, "b": 99}`)
		ok, msg := compareMap(jsonObj, reflect.ValueOf(map[string]int{"a": 1, "b": 2}))
		assert.False(t, ok)
		assert.Contains(t, msg, "b")
	})

	t.Run("not object", func(t *testing.T) {
		jsonObj := gjson.Parse(`[1, 2]`)
		ok, msg := compareMap(jsonObj, reflect.ValueOf(map[string]int{"a": 1}))
		assert.False(t, ok)
		assert.Contains(t, msg, "object")
	})
}

func TestCompareModeConstants(t *testing.T) {
	assert.Equal(t, CompareMode(0), ModePartial)
	assert.Equal(t, CompareMode(1), ModeExact)
}

func TestCompareObjectPartial_SimpleStruct(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
		City string `json:"city"`
	}

	jsonData := `{"name": "John", "age": 30, "city": "NYC", "extra": "field"}`
	jsonObj := gjson.Parse(jsonData)

	ok, msg := CompareObjectPartial(jsonObj, Person{
		Name: "John",
		Age:  30,
	})
	assert.True(t, ok, "Should match: %s", msg)

	ok, msg = CompareObjectPartial(jsonObj, Person{
		Name: "John",
	})
	assert.True(t, ok, "Should match with only Name: %s", msg)

	ok, msg = CompareObjectPartial(jsonObj, Person{
		Name: "Jane",
	})
	assert.False(t, ok, "Should not match")
	assert.Contains(t, msg, "name")
}

func TestCompareObjectPartial_WithMap(t *testing.T) {
	type Category struct {
		Id    string            `json:"id"`
		Names map[string]string `json:"names"`
		Type  string            `json:"type"`
	}

	jsonData := `{
		"id": "123",
		"names": {"ru": "Категория", "en": "Category"},
		"type": "category",
		"extra": "ignored"
	}`
	jsonObj := gjson.Parse(jsonData)

	ok, msg := CompareObjectPartial(jsonObj, Category{
		Id:    "123",
		Names: map[string]string{"ru": "Категория", "en": "Category"},
	})
	assert.True(t, ok, "Should match: %s", msg)

	ok, msg = CompareObjectPartial(jsonObj, Category{
		Id: "123",
	})
	assert.True(t, ok, "Should match with only Id: %s", msg)

	ok, msg = CompareObjectPartial(jsonObj, Category{
		Names: map[string]string{"ru": "Другое"},
	})
	assert.False(t, ok, "Should not match")
}

func TestCompareObjectPartial_WithBool(t *testing.T) {
	type Item struct {
		Name      string `json:"name"`
		IsDefault bool   `json:"isDefault"`
		PassToCms bool   `json:"passToCms"`
	}

	jsonData := `{"name": "Test", "isDefault": false, "passToCms": true}`
	jsonObj := gjson.Parse(jsonData)

	ok, msg := CompareObjectPartial(jsonObj, Item{
		Name:      "Test",
		PassToCms: true,
	})
	assert.True(t, ok, "Should match: %s", msg)
}

func TestFindInArray(t *testing.T) {
	type Item struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	jsonData := `[
		{"id": "1", "name": "First"},
		{"id": "2", "name": "Second"},
		{"id": "3", "name": "Third"}
	]`
	jsonArr := gjson.Parse(jsonData)

	idx, item := FindInArray(jsonArr, Item{Id: "2"})
	assert.Equal(t, 1, idx)
	assert.Equal(t, "Second", item.Get("name").String())

	idx, item = FindInArray(jsonArr, Item{Name: "Third"})
	assert.Equal(t, 2, idx)
	assert.Equal(t, "3", item.Get("id").String())

	idx, _ = FindInArray(jsonArr, Item{Id: "999"})
	assert.Equal(t, -1, idx)
}

func TestCompareObjectPartial_NestedStruct(t *testing.T) {
	type Address struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}
	type Person struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	jsonData := `{
		"name": "John",
		"address": {
			"city": "NYC",
			"country": "USA"
		}
	}`
	jsonObj := gjson.Parse(jsonData)

	ok, msg := CompareObjectPartial(jsonObj, Person{
		Name: "John",
		Address: Address{
			City: "NYC",
		},
	})
	assert.True(t, ok, "Should match nested struct: %s", msg)

	ok, msg = CompareObjectPartial(jsonObj, Person{
		Address: Address{
			City: "LA",
		},
	})
	assert.False(t, ok, "Should not match wrong city")
}

func TestToJSONFieldName(t *testing.T) {
	type TestStruct struct {
		Name          string `json:"name"`
		UserID        string `json:"userId"`
		NoTag         string
		CustomName    string `json:"custom_name"`
		IgnoredField  string `json:"-"`
		OmitEmpty     string `json:"omitEmpty,omitempty"`
		PascalCase    string
		CamelCaseTest string
	}

	tests := []struct {
		fieldName string
		expected  string
	}{
		{"Name", "name"},
		{"UserID", "userId"},
		{"NoTag", "noTag"},
		{"CustomName", "custom_name"},
		{"OmitEmpty", "omitEmpty"},
		{"PascalCase", "pascalCase"},
		{"CamelCaseTest", "camelCaseTest"},
	}

	typ := reflect.TypeOf(TestStruct{})
	for _, tt := range tests {
		field, _ := typ.FieldByName(tt.fieldName)
		result := toJSONFieldName(field)
		assert.Equal(t, tt.expected, result, "Field %s", tt.fieldName)
	}
}

func TestCompareObjectExact_ZeroValues(t *testing.T) {
	type Item struct {
		Name      string `json:"name"`
		Count     int    `json:"count"`
		IsDefault bool   `json:"isDefault"`
		ParentId  string `json:"parentId"`
	}

	jsonData := `{"name": "Test", "count": 0, "isDefault": false, "parentId": ""}`
	jsonObj := gjson.Parse(jsonData)

	ok, msg := CompareObjectExact(jsonObj, Item{
		Name:      "Test",
		Count:     0,
		IsDefault: false,
		ParentId:  "",
	})
	assert.True(t, ok, "Should match with exact zero values: %s", msg)

	jsonDataNonZero := `{"name": "Test", "count": 5, "isDefault": true, "parentId": "abc"}`
	jsonObjNonZero := gjson.Parse(jsonDataNonZero)

	ok, msg = CompareObjectExact(jsonObjNonZero, Item{
		Name:      "Test",
		Count:     0,
		IsDefault: false,
		ParentId:  "",
	})
	assert.False(t, ok, "Should not match - count differs")
	assert.Contains(t, msg, "count")
}

func TestCompareObjectExact_vs_Partial(t *testing.T) {
	type Category struct {
		Id         string `json:"id"`
		Name       string `json:"name"`
		GamesCount int    `json:"gamesCount"`
		IsDefault  bool   `json:"isDefault"`
	}

	jsonData := `{"id": "123", "name": "Sports", "gamesCount": 10, "isDefault": true}`
	jsonObj := gjson.Parse(jsonData)

	expected := Category{
		Id:         "123",
		Name:       "Sports",
		GamesCount: 0,
		IsDefault:  false,
	}

	ok, _ := CompareObjectPartial(jsonObj, expected)
	assert.True(t, ok, "Partial should match - zero values skipped")

	ok, msg := CompareObjectExact(jsonObj, expected)
	assert.False(t, ok, "Exact should not match - gamesCount and isDefault differ")
	assert.True(t, msg != "", "Should have error message")
}

func TestFindInArrayExact(t *testing.T) {
	type Item struct {
		Id        string `json:"id"`
		Name      string `json:"name"`
		IsDefault bool   `json:"isDefault"`
	}

	jsonData := `[
		{"id": "1", "name": "First", "isDefault": false},
		{"id": "2", "name": "Second", "isDefault": true},
		{"id": "3", "name": "Third", "isDefault": false}
	]`
	jsonArr := gjson.Parse(jsonData)

	idx, item := FindInArrayExact(jsonArr, Item{Id: "1", Name: "First", IsDefault: false})
	assert.Equal(t, 0, idx)
	assert.Equal(t, "1", item.Get("id").String())

	idx, item = FindInArrayExact(jsonArr, Item{Id: "2", Name: "Second", IsDefault: true})
	assert.Equal(t, 1, idx)
	assert.Equal(t, "2", item.Get("id").String())

	idx, _ = FindInArrayExact(jsonArr, Item{Id: "2", Name: "Second", IsDefault: false})
	assert.Equal(t, -1, idx, "Exact should not find - isDefault differs")

	idx, _ = FindInArray(jsonArr, Item{Id: "2", Name: "Second", IsDefault: false})
	assert.Equal(t, 1, idx, "Partial should find it because IsDefault=false is zero value and skipped")
}

func TestCompareObjectExact_WithMap(t *testing.T) {
	type Category struct {
		Id    string            `json:"id"`
		Names map[string]string `json:"names"`
		Type  string            `json:"type"`
	}

	jsonData := `{
		"id": "123",
		"names": {"ru": "Категория", "en": "Category"},
		"type": "category"
	}`
	jsonObj := gjson.Parse(jsonData)

	ok, msg := CompareObjectExact(jsonObj, Category{
		Id:    "123",
		Names: map[string]string{"ru": "Категория", "en": "Category"},
		Type:  "category",
	})
	assert.True(t, ok, "Should match: %s", msg)

	ok, msg = CompareObjectExact(jsonObj, Category{
		Id:    "123",
		Names: map[string]string{"ru": "Другое"},
		Type:  "category",
	})
	assert.False(t, ok, "Should not match - names differ")
}

func TestCompareObjectExact_NestedStruct(t *testing.T) {
	type Address struct {
		City    string `json:"city"`
		Country string `json:"country"`
	}
	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	jsonData := `{
		"name": "John",
		"age": 30,
		"address": {
			"city": "NYC",
			"country": "USA"
		}
	}`
	jsonObj := gjson.Parse(jsonData)

	ok, msg := CompareObjectExact(jsonObj, Person{
		Name: "John",
		Age:  30,
		Address: Address{
			City:    "NYC",
			Country: "USA",
		},
	})
	assert.True(t, ok, "Should match nested struct: %s", msg)

	ok, msg = CompareObjectExact(jsonObj, Person{
		Name: "John",
		Age:  30,
		Address: Address{
			City:    "LA",
			Country: "USA",
		},
	})
	assert.False(t, ok, "Should not match wrong city")
	assert.Contains(t, msg, "address")
}

func TestCompareObjectExact_MapStringInterface(t *testing.T) {
	jsonData := `{"id": "123", "names": {"ru": "Test", "en": "Test"}, "passToCms": false}`
	jsonObj := gjson.Parse(jsonData)

	type Category struct {
		Id        string                 `json:"id"`
		Names     map[string]interface{} `json:"names"`
		PassToCms interface{}            `json:"passToCms"`
	}

	expected := Category{
		Id:        "123",
		Names:     map[string]interface{}{"ru": "Test", "en": "Test"},
		PassToCms: false,
	}

	ok, msg := CompareObjectExact(jsonObj, expected)
	assert.True(t, ok, "Should match map[string]interface{} and interface{}: %s", msg)
}

func TestCompareObjectExact_PointerString(t *testing.T) {
	jsonData := `{"id": "123", "name": "TestName", "names": {"ru": "Test"}, "passToCms": false}`
	jsonObj := gjson.Parse(jsonData)

	type Category struct {
		Id        string                 `json:"id"`
		Name      *string                `json:"name"`
		Names     map[string]interface{} `json:"names"`
		PassToCms interface{}            `json:"passToCms"`
	}

	name := "TestName"
	expected := Category{
		Id:        "123",
		Name:      &name,
		Names:     map[string]interface{}{"ru": "Test"},
		PassToCms: false,
	}

	ok, msg := CompareObjectExact(jsonObj, expected)
	assert.True(t, ok, "Should match *string fields: %s", msg)
}

func TestValidateBytes(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid JSON object", []byte(`{"key": "value"}`), false},
		{"valid JSON array", []byte(`[1, 2, 3]`), false},
		{"valid JSON string", []byte(`"hello"`), false},
		{"valid JSON number", []byte(`42`), false},
		{"valid JSON null", []byte(`null`), false},
		{"empty bytes", []byte(``), true},
		{"invalid JSON", []byte(`{invalid}`), true},
		{"truncated JSON", []byte(`{"key":`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBytes(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid JSON object", `{"key": "value"}`, false},
		{"valid JSON array", `[1, 2, 3]`, false},
		{"empty string", ``, true},
		{"invalid JSON", `{invalid}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetField(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		path      string
		wantValue string
		wantErr   bool
	}{
		{"simple field", []byte(`{"name": "John"}`), "name", "John", false},
		{"nested field", []byte(`{"user": {"name": "John"}}`), "user.name", "John", false},
		{"array index", []byte(`{"items": [1, 2, 3]}`), "items.1", "2", false},
		{"non-existent field", []byte(`{"a": 1}`), "b", "", false},
		{"invalid JSON", []byte(`{invalid}`), "a", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetField(tt.input, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantValue != "" {
					assert.Equal(t, tt.wantValue, result.String())
				}
			}
		})
	}
}

func TestGetFieldFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		path      string
		wantValue string
		wantErr   bool
	}{
		{"simple field", `{"name": "John"}`, "name", "John", false},
		{"nested field", `{"user": {"name": "John"}}`, "user.name", "John", false},
		{"invalid JSON", `{invalid}`, "a", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetFieldFromString(tt.input, tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantValue != "" {
					assert.Equal(t, tt.wantValue, result.String())
				}
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid JSON object", []byte(`{"key": "value"}`), false},
		{"valid JSON array", []byte(`[1, 2, 3]`), false},
		{"invalid JSON", []byte(`{invalid}`), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Parse(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, result.Exists())
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid JSON object", `{"key": "value"}`, false},
		{"valid JSON array", `[1, 2, 3]`, false},
		{"invalid JSON", `{invalid}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseString(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, result.Exists())
			}
		})
	}
}

func TestIsNull(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		path   string
		expect bool
	}{
		{"null value", `{"a": null}`, "a", true},
		{"string value", `{"a": "hello"}`, "a", false},
		{"number value", `{"a": 42}`, "a", false},
		{"empty string", `{"a": ""}`, "a", false},
		{"boolean false", `{"a": false}`, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Get(tt.json, tt.path)
			got := IsNull(res)
			assert.Equal(t, tt.expect, got)
		})
	}
}
