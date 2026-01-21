package jsonutil

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

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
