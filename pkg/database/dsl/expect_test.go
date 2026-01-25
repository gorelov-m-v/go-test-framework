package dsl

import (
	"database/sql"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestModel struct {
	ID        int64          `db:"id"`
	Name      string         `db:"name"`
	Email     sql.NullString `db:"email"`
	Status    int16          `db:"status_id"`
	IsActive  bool           `db:"is_active"`
	IsDeleted sql.NullBool   `db:"is_deleted"`
	Score     sql.NullInt64  `db:"score"`
	Data      json.RawMessage `db:"data"`
}

func TestGetFieldValueByColumnName_Found(t *testing.T) {
	model := TestModel{ID: 123, Name: "test"}
	result, err := getFieldValueByColumnName(model, "id")

	require.NoError(t, err)
	assert.Equal(t, int64(123), result)
}

func TestGetFieldValueByColumnName_String(t *testing.T) {
	model := TestModel{Name: "hello"}
	result, err := getFieldValueByColumnName(model, "name")

	require.NoError(t, err)
	assert.Equal(t, "hello", result)
}

func TestGetFieldValueByColumnName_NotFound(t *testing.T) {
	model := TestModel{ID: 123}
	_, err := getFieldValueByColumnName(model, "nonexistent")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetFieldValueByColumnName_EmptyColumnName(t *testing.T) {
	model := TestModel{ID: 123}
	_, err := getFieldValueByColumnName(model, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestGetFieldValueByColumnName_NilTarget(t *testing.T) {
	_, err := getFieldValueByColumnName(nil, "id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}

func TestGetFieldValueByColumnName_Pointer(t *testing.T) {
	model := &TestModel{ID: 456, Name: "pointer"}
	result, err := getFieldValueByColumnName(model, "id")

	require.NoError(t, err)
	assert.Equal(t, int64(456), result)
}

func TestGetFieldValueByColumnName_NilPointer(t *testing.T) {
	var model *TestModel
	_, err := getFieldValueByColumnName(model, "id")

	assert.Error(t, err)
}

func TestGetFieldValueByColumnName_NullString_Valid(t *testing.T) {
	model := TestModel{Email: sql.NullString{String: "test@test.com", Valid: true}}
	result, err := getFieldValueByColumnName(model, "email")

	require.NoError(t, err)
	assert.Equal(t, sql.NullString{String: "test@test.com", Valid: true}, result)
}

func TestGetFieldValueByColumnName_NullString_Invalid(t *testing.T) {
	model := TestModel{Email: sql.NullString{Valid: false}}
	result, err := getFieldValueByColumnName(model, "email")

	require.NoError(t, err)
	assert.Equal(t, sql.NullString{Valid: false}, result)
}

func TestGetFieldValueByColumnName_Int16(t *testing.T) {
	model := TestModel{Status: 2}
	result, err := getFieldValueByColumnName(model, "status_id")

	require.NoError(t, err)
	assert.Equal(t, int16(2), result)
}

func TestGetFieldValueByColumnName_Bool(t *testing.T) {
	model := TestModel{IsActive: true}
	result, err := getFieldValueByColumnName(model, "is_active")

	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestMakeFoundExpectation_Success(t *testing.T) {
	exp := makeFoundExpectation[TestModel]()
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeFoundExpectation_NoRows(t *testing.T) {
	exp := makeFoundExpectation[TestModel]()
	var model TestModel

	result := exp.Check(sql.ErrNoRows, model)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
	assert.Contains(t, result.Reason, "ErrNoRows")
}

func TestMakeFoundExpectation_OtherError(t *testing.T) {
	exp := makeFoundExpectation[TestModel]()
	var model TestModel

	result := exp.Check(sql.ErrConnDone, model)

	assert.False(t, result.Ok)
	assert.False(t, result.Retryable)
}

func TestMakeNotFoundExpectation_Success(t *testing.T) {
	exp := makeNotFoundExpectation[TestModel]()
	var model TestModel

	result := exp.Check(sql.ErrNoRows, model)

	assert.True(t, result.Ok)
}

func TestMakeNotFoundExpectation_Found(t *testing.T) {
	exp := makeNotFoundExpectation[TestModel]()
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestMakeNotFoundExpectation_OtherError(t *testing.T) {
	exp := makeNotFoundExpectation[TestModel]()
	var model TestModel

	result := exp.Check(sql.ErrConnDone, model)

	assert.False(t, result.Ok)
	assert.False(t, result.Retryable)
}

func TestMakeColumnEqualsExpectation_IntMatch(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("id", int64(123))
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnEqualsExpectation_IntMismatch(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("id", int64(456))
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnEqualsExpectation_IntCrossType(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("id", 123)
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnEqualsExpectation_Int16(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("status_id", 2)
	model := TestModel{Status: 2}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnEqualsExpectation_StringMatch(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("name", "test")
	model := TestModel{Name: "test"}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnEqualsExpectation_StringMismatch(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("name", "expected")
	model := TestModel{Name: "actual"}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnEqualsExpectation_NullStringMatch(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("email", "test@test.com")
	model := TestModel{Email: sql.NullString{String: "test@test.com", Valid: true}}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnEqualsExpectation_NoRows(t *testing.T) {
	exp := makeColumnEqualsExpectation[TestModel]("id", 123)
	var model TestModel

	result := exp.Check(sql.ErrNoRows, model)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestMakeColumnNotEqualsExpectation_Success(t *testing.T) {
	exp := makeColumnNotEqualsExpectation[TestModel]("id", int64(456))
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnNotEqualsExpectation_Failure(t *testing.T) {
	exp := makeColumnNotEqualsExpectation[TestModel]("id", int64(123))
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnNotEmptyExpectation_Success(t *testing.T) {
	exp := makeColumnNotEmptyExpectation[TestModel]("name")
	model := TestModel{Name: "not empty"}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnNotEmptyExpectation_EmptyString(t *testing.T) {
	exp := makeColumnNotEmptyExpectation[TestModel]("name")
	model := TestModel{Name: ""}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnNotEmptyExpectation_NonZeroInt(t *testing.T) {
	exp := makeColumnNotEmptyExpectation[TestModel]("id")
	model := TestModel{ID: 123}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnEmptyExpectation_EmptyString(t *testing.T) {
	exp := makeColumnEmptyExpectation[TestModel]("name")
	model := TestModel{Name: ""}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnEmptyExpectation_NonEmpty(t *testing.T) {
	exp := makeColumnEmptyExpectation[TestModel]("name")
	model := TestModel{Name: "not empty"}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnIsNullExpectation_Success(t *testing.T) {
	exp := makeColumnIsNullExpectation[TestModel]("email")
	model := TestModel{Email: sql.NullString{Valid: false}}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnIsNullExpectation_NotNull(t *testing.T) {
	exp := makeColumnIsNullExpectation[TestModel]("email")
	model := TestModel{Email: sql.NullString{String: "test", Valid: true}}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnIsNotNullExpectation_Success(t *testing.T) {
	exp := makeColumnIsNotNullExpectation[TestModel]("email")
	model := TestModel{Email: sql.NullString{String: "test", Valid: true}}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnIsNotNullExpectation_IsNull(t *testing.T) {
	exp := makeColumnIsNotNullExpectation[TestModel]("email")
	model := TestModel{Email: sql.NullString{Valid: false}}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnTrueExpectation_Success(t *testing.T) {
	exp := makeColumnTrueExpectation[TestModel]("is_active")
	model := TestModel{IsActive: true}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnTrueExpectation_False(t *testing.T) {
	exp := makeColumnTrueExpectation[TestModel]("is_active")
	model := TestModel{IsActive: false}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnFalseExpectation_Success(t *testing.T) {
	exp := makeColumnFalseExpectation[TestModel]("is_active")
	model := TestModel{IsActive: false}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnFalseExpectation_True(t *testing.T) {
	exp := makeColumnFalseExpectation[TestModel]("is_active")
	model := TestModel{IsActive: true}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnJSONEqualsExpectation_Success(t *testing.T) {
	exp := makeColumnJSONEqualsExpectation[TestModel]("data", map[string]interface{}{"key": "value"})
	model := TestModel{Data: json.RawMessage(`{"key": "value"}`)}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

func TestMakeColumnJSONEqualsExpectation_Mismatch(t *testing.T) {
	exp := makeColumnJSONEqualsExpectation[TestModel]("data", map[string]interface{}{"key": "expected"})
	model := TestModel{Data: json.RawMessage(`{"key": "actual"}`)}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnJSONEqualsExpectation_MissingKey(t *testing.T) {
	exp := makeColumnJSONEqualsExpectation[TestModel]("data", map[string]interface{}{"missing": "value"})
	model := TestModel{Data: json.RawMessage(`{"key": "value"}`)}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnJSONEqualsExpectation_InvalidJSON(t *testing.T) {
	exp := makeColumnJSONEqualsExpectation[TestModel]("data", map[string]interface{}{"key": "value"})
	model := TestModel{Data: json.RawMessage(`not valid json`)}

	result := exp.Check(nil, model)

	assert.False(t, result.Ok)
}

func TestMakeColumnJSONEqualsExpectation_PartialMatch(t *testing.T) {
	exp := makeColumnJSONEqualsExpectation[TestModel]("data", map[string]interface{}{"key": "value"})
	model := TestModel{Data: json.RawMessage(`{"key": "value", "extra": "field"}`)}

	result := exp.Check(nil, model)

	assert.True(t, result.Ok)
}

type SimpleTestModel struct {
	ID     int64  `db:"id"`
	Name   string `db:"name"`
	Status int16  `db:"status_id"`
}

func TestMakeRowExpectation_ExactMatch(t *testing.T) {
	expected := SimpleTestModel{ID: 123, Name: "test", Status: 1}
	exp := makeRowExpectation(expected)
	actual := SimpleTestModel{ID: 123, Name: "test", Status: 1}

	result := exp.Check(nil, actual)

	assert.True(t, result.Ok, "Expected match, got: %s", result.Reason)
}

func TestMakeRowExpectation_Mismatch(t *testing.T) {
	expected := TestModel{ID: 123, Name: "test"}
	exp := makeRowExpectation(expected)
	actual := TestModel{ID: 123, Name: "different"}

	result := exp.Check(nil, actual)

	assert.False(t, result.Ok)
}

func TestMakeRowExpectation_NoRows(t *testing.T) {
	expected := TestModel{ID: 123}
	exp := makeRowExpectation(expected)
	var actual TestModel

	result := exp.Check(sql.ErrNoRows, actual)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
}

func TestMakeRowPartialExpectation_Match(t *testing.T) {
	expected := TestModel{ID: 123}
	exp := makeRowPartialExpectation(expected)
	actual := TestModel{ID: 123, Name: "extra", Status: 2}

	result := exp.Check(nil, actual)

	assert.True(t, result.Ok)
}

func TestMakeRowPartialExpectation_ZeroFieldsIgnored(t *testing.T) {
	expected := TestModel{Name: "test"}
	exp := makeRowPartialExpectation(expected)
	actual := TestModel{ID: 999, Name: "test", Status: 5}

	result := exp.Check(nil, actual)

	assert.True(t, result.Ok)
}

func TestMakeRowPartialExpectation_Mismatch(t *testing.T) {
	expected := TestModel{ID: 123, Name: "expected"}
	exp := makeRowPartialExpectation(expected)
	actual := TestModel{ID: 123, Name: "actual"}

	result := exp.Check(nil, actual)

	assert.False(t, result.Ok)
}

func TestMakeCountAllExpectation_Success(t *testing.T) {
	exp := makeCountAllExpectation[TestModel](3)
	results := []TestModel{{ID: 1}, {ID: 2}, {ID: 3}}

	result := exp.Check(nil, results)

	assert.True(t, result.Ok)
}

func TestMakeCountAllExpectation_Mismatch(t *testing.T) {
	exp := makeCountAllExpectation[TestModel](3)
	results := []TestModel{{ID: 1}, {ID: 2}}

	result := exp.Check(nil, results)

	assert.False(t, result.Ok)
	assert.Contains(t, result.Reason, "expected 3 rows")
}

func TestMakeCountAllExpectation_Empty(t *testing.T) {
	exp := makeCountAllExpectation[TestModel](0)
	results := []TestModel{}

	result := exp.Check(nil, results)

	assert.True(t, result.Ok)
}

func TestMakeNotEmptyAllExpectation_Success(t *testing.T) {
	exp := makeNotEmptyAllExpectation[TestModel]()
	results := []TestModel{{ID: 1}}

	result := exp.Check(nil, results)

	assert.True(t, result.Ok)
}

func TestMakeNotEmptyAllExpectation_Empty(t *testing.T) {
	exp := makeNotEmptyAllExpectation[TestModel]()
	results := []TestModel{}

	result := exp.Check(nil, results)

	assert.False(t, result.Ok)
	assert.Contains(t, result.Reason, "0 rows")
}

func TestCompareStructsExact_Equal(t *testing.T) {
	expected := SimpleTestModel{ID: 1, Name: "test", Status: 2}
	actual := SimpleTestModel{ID: 1, Name: "test", Status: 2}

	ok, msg := compareStructsExact(expected, actual)

	assert.True(t, ok, "Expected match, got: %s", msg)
}

func TestCompareStructsExact_NotEqual(t *testing.T) {
	expected := TestModel{ID: 1, Name: "test"}
	actual := TestModel{ID: 1, Name: "different"}

	ok, msg := compareStructsExact(expected, actual)

	assert.False(t, ok)
	assert.Contains(t, msg, "name")
}

func TestCompareStructsPartial_Match(t *testing.T) {
	expected := TestModel{ID: 1}
	actual := TestModel{ID: 1, Name: "ignored", Status: 99}

	ok, _ := compareStructsPartial(expected, actual)

	assert.True(t, ok)
}

func TestCompareStructsPartial_ZeroFieldsIgnored(t *testing.T) {
	expected := TestModel{}
	actual := TestModel{ID: 999, Name: "anything"}

	ok, _ := compareStructsPartial(expected, actual)

	assert.True(t, ok)
}

func TestCompareStructsPartial_Mismatch(t *testing.T) {
	expected := TestModel{Name: "expected"}
	actual := TestModel{Name: "actual"}

	ok, _ := compareStructsPartial(expected, actual)

	assert.False(t, ok)
}

func TestGetDBColumnName_WithTag(t *testing.T) {
	typ := TestModel{}
	field, _ := reflect.TypeOf(typ).FieldByName("ID")
	name := getDBColumnName(field)

	assert.Equal(t, "id", name)
}

func TestGetDBColumnName_WithoutTag(t *testing.T) {
	type NoTagModel struct {
		ID int64
	}
	field, _ := reflect.TypeOf(NoTagModel{}).FieldByName("ID")
	name := getDBColumnName(field)

	assert.Equal(t, "ID", name)
}
