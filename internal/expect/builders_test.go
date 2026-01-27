package expect

import (
	"errors"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

var errNoRows = errors.New("no rows")
var errQuery = errors.New("query failed")

type testRow struct {
	ID     int
	Name   string
	Active bool
	Data   []byte
}

func getValue(row testRow, column string) (any, error) {
	switch column {
	case "id":
		return row.ID, nil
	case "name":
		return row.Name, nil
	case "active":
		return row.Active, nil
	case "data":
		return row.Data, nil
	default:
		return nil, errors.New("unknown column")
	}
}

func getJSON(row testRow) ([]byte, error) {
	return row.Data, nil
}

func TestCheckColumnError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		errNoRows error
		wantOK    bool
		wantRetry bool
	}{
		{"no error", nil, errNoRows, true, false},
		{"no rows error", errNoRows, errNoRows, false, true},
		{"query error", errQuery, errNoRows, false, false},
		{"no rows with nil errNoRows", errNoRows, nil, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := checkColumnError[testRow](tt.err, tt.errNoRows, "col")
			if ok != tt.wantOK {
				t.Errorf("checkColumnError() ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok && result.Retryable != tt.wantRetry {
				t.Errorf("checkColumnError() Retryable = %v, want %v", result.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestGetColumnValue(t *testing.T) {
	row := testRow{ID: 1, Name: "test", Active: true}

	tests := []struct {
		name    string
		column  string
		wantVal any
		wantOK  bool
	}{
		{"valid column id", "id", 1, true},
		{"valid column name", "name", "test", true},
		{"valid column active", "active", true, true},
		{"invalid column", "unknown", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, _, ok := getColumnValue(row, tt.column, getValue)
			if ok != tt.wantOK {
				t.Errorf("getColumnValue() ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && val != tt.wantVal {
				t.Errorf("getColumnValue() val = %v, want %v", val, tt.wantVal)
			}
		})
	}
}

func TestBuildColumnExpectation_Check(t *testing.T) {
	tests := []struct {
		name      string
		row       testRow
		err       error
		column    string
		check     ValueCheck
		wantOK    bool
		wantRetry bool
	}{
		{
			name:   "success - not empty",
			row:    testRow{Name: "hello"},
			err:    nil,
			column: "name",
			check:  CheckNotEmpty(),
			wantOK: true,
		},
		{
			name:      "fail - empty",
			row:       testRow{Name: ""},
			err:       nil,
			column:    "name",
			check:     CheckNotEmpty(),
			wantOK:    false,
			wantRetry: true,
		},
		{
			name:      "fail - no rows",
			row:       testRow{},
			err:       errNoRows,
			column:    "name",
			check:     CheckNotEmpty(),
			wantOK:    false,
			wantRetry: true,
		},
		{
			name:      "fail - query error",
			row:       testRow{},
			err:       errQuery,
			column:    "name",
			check:     CheckNotEmpty(),
			wantOK:    false,
			wantRetry: false,
		},
		{
			name:      "fail - invalid column",
			row:       testRow{Name: "hello"},
			err:       nil,
			column:    "unknown",
			check:     CheckNotEmpty(),
			wantOK:    false,
			wantRetry: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildColumnExpectation(ColumnExpectationConfig[testRow]{
				ColumnName: tt.column,
				ExpectName: "test",
				GetValue:   getValue,
				ErrNoRows:  errNoRows,
				Check:      tt.check,
			})

			result := exp.Check(tt.err, tt.row)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
			if !result.Ok && result.Retryable != tt.wantRetry {
				t.Errorf("Check() Retryable = %v, want %v", result.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestBuildColumnBoolExpectation_Check(t *testing.T) {
	toBool := func(v any) (bool, bool) {
		if b, ok := v.(bool); ok {
			return b, true
		}
		return false, false
	}

	tests := []struct {
		name         string
		row          testRow
		err          error
		expectedBool bool
		wantOK       bool
	}{
		{
			name:         "success - true matches",
			row:          testRow{Active: true},
			err:          nil,
			expectedBool: true,
			wantOK:       true,
		},
		{
			name:         "success - false matches",
			row:          testRow{Active: false},
			err:          nil,
			expectedBool: false,
			wantOK:       true,
		},
		{
			name:         "fail - true expected false",
			row:          testRow{Active: true},
			err:          nil,
			expectedBool: false,
			wantOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var check ValueCheck
			if tt.expectedBool {
				check = CheckTrue()
			} else {
				check = CheckFalse()
			}

			exp := BuildColumnBoolExpectation(ColumnBoolExpectationConfig[testRow]{
				ColumnName:   "active",
				ExpectName:   "test",
				GetValue:     getValue,
				ErrNoRows:    errNoRows,
				Check:        check,
				ExpectedBool: tt.expectedBool,
				ToBoolFunc:   toBool,
			})

			result := exp.Check(tt.err, tt.row)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildColumnNullExpectation_Check(t *testing.T) {
	isNull := func(v any) bool {
		return v == nil
	}

	tests := []struct {
		name         string
		row          testRow
		expectedNull bool
		wantOK       bool
	}{
		{
			name:         "success - not null when expected",
			row:          testRow{Name: "hello"},
			expectedNull: false,
			wantOK:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildColumnNullExpectation(ColumnNullExpectationConfig[testRow]{
				ColumnName:   "name",
				ExpectName:   "test",
				GetValue:     getValue,
				ErrNoRows:    errNoRows,
				Check:        CheckIsNotNull(),
				ExpectedNull: tt.expectedNull,
				IsNullFunc:   isNull,
			})

			result := exp.Check(nil, tt.row)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildColumnEmptyExpectation_Check(t *testing.T) {
	isEmpty := func(v any) bool {
		if s, ok := v.(string); ok {
			return s == ""
		}
		return false
	}

	tests := []struct {
		name          string
		row           testRow
		expectedEmpty bool
		wantOK        bool
	}{
		{
			name:          "success - not empty when expected",
			row:           testRow{Name: "hello"},
			expectedEmpty: false,
			wantOK:        true,
		},
		{
			name:          "success - empty when expected",
			row:           testRow{Name: ""},
			expectedEmpty: true,
			wantOK:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var check ValueCheck
			if tt.expectedEmpty {
				check = CheckEmpty()
			} else {
				check = CheckNotEmpty()
			}

			exp := BuildColumnEmptyExpectation(ColumnEmptyExpectationConfig[testRow]{
				ColumnName:    "name",
				ExpectName:    "test",
				GetValue:      getValue,
				ErrNoRows:     errNoRows,
				Check:         check,
				ExpectedEmpty: tt.expectedEmpty,
				IsEmptyFunc:   isEmpty,
			})

			result := exp.Check(nil, tt.row)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestCheckJSONField(t *testing.T) {
	tests := []struct {
		name      string
		row       testRow
		path      string
		wantOK    bool
		wantRetry bool
	}{
		{
			name:   "valid json with existing path",
			row:    testRow{Data: []byte(`{"name": "test"}`)},
			path:   "name",
			wantOK: true,
		},
		{
			name:      "valid json with non-existing path",
			row:       testRow{Data: []byte(`{"name": "test"}`)},
			path:      "missing",
			wantOK:    false,
			wantRetry: true,
		},
		{
			name:      "invalid json",
			row:       testRow{Data: []byte(`not json`)},
			path:      "name",
			wantOK:    false,
			wantRetry: true,
		},
		{
			name:      "empty json bytes",
			row:       testRow{Data: []byte{}},
			path:      "name",
			wantOK:    false,
			wantRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := JSONFieldExpectationConfig[testRow]{
				Path:    tt.path,
				GetJSON: getJSON,
			}
			_, result, ok := checkJSONField(cfg, nil, tt.row)
			if ok != tt.wantOK {
				t.Errorf("checkJSONField() ok = %v, want %v, reason: %s", ok, tt.wantOK, result.Reason)
			}
			if !ok && result.Retryable != tt.wantRetry {
				t.Errorf("checkJSONField() Retryable = %v, want %v", result.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestCheckJSONField_WithPreCheck(t *testing.T) {
	preCheckFail := func(err error, row testRow) (polling.CheckResult, bool) {
		return polling.CheckResult{Ok: false, Retryable: true, Reason: "precheck failed"}, false
	}
	preCheckPass := func(err error, row testRow) (polling.CheckResult, bool) {
		return polling.CheckResult{}, true
	}

	t.Run("precheck fails", func(t *testing.T) {
		cfg := JSONFieldExpectationConfig[testRow]{
			Path:     "name",
			GetJSON:  getJSON,
			PreCheck: preCheckFail,
		}
		_, result, ok := checkJSONField(cfg, nil, testRow{Data: []byte(`{"name": "test"}`)})
		if ok {
			t.Error("Expected precheck to fail")
		}
		if result.Reason != "precheck failed" {
			t.Errorf("Expected precheck reason, got: %s", result.Reason)
		}
	})

	t.Run("precheck passes", func(t *testing.T) {
		cfg := JSONFieldExpectationConfig[testRow]{
			Path:     "name",
			GetJSON:  getJSON,
			PreCheck: preCheckPass,
		}
		_, _, ok := checkJSONField(cfg, nil, testRow{Data: []byte(`{"name": "test"}`)})
		if !ok {
			t.Error("Expected precheck to pass")
		}
	})
}

func TestCheckJSONField_GetJSONError(t *testing.T) {
	getJSONError := func(row testRow) ([]byte, error) {
		return nil, errors.New("failed to get json")
	}

	cfg := JSONFieldExpectationConfig[testRow]{
		Path:    "name",
		GetJSON: getJSONError,
	}
	_, result, ok := checkJSONField(cfg, nil, testRow{})
	if ok {
		t.Error("Expected failure when GetJSON returns error")
	}
	if !result.Retryable {
		t.Error("Expected retryable when GetJSON returns error")
	}
}

func TestBuildJSONFieldExpectation_Check(t *testing.T) {
	tests := []struct {
		name   string
		row    testRow
		path   string
		check  JSONCheck
		wantOK bool
	}{
		{
			name:   "success - equals",
			row:    testRow{Data: []byte(`{"name": "test"}`)},
			path:   "name",
			check:  JSONCheckEquals("test"),
			wantOK: true,
		},
		{
			name:   "fail - not equals",
			row:    testRow{Data: []byte(`{"name": "test"}`)},
			path:   "name",
			check:  JSONCheckEquals("other"),
			wantOK: false,
		},
		{
			name:   "success - not empty",
			row:    testRow{Data: []byte(`{"name": "test"}`)},
			path:   "name",
			check:  JSONCheckNotEmpty(),
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildJSONFieldExpectation(JSONFieldExpectationConfig[testRow]{
				Path:       tt.path,
				ExpectName: "test",
				GetJSON:    getJSON,
				Check:      tt.check,
			})

			result := exp.Check(nil, tt.row)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildJSONFieldNullExpectation_Check(t *testing.T) {
	tests := []struct {
		name         string
		row          testRow
		path         string
		expectedNull bool
		wantOK       bool
	}{
		{
			name:         "success - null when expected",
			row:          testRow{Data: []byte(`{"name": null}`)},
			path:         "name",
			expectedNull: true,
			wantOK:       true,
		},
		{
			name:         "success - not null when expected",
			row:          testRow{Data: []byte(`{"name": "test"}`)},
			path:         "name",
			expectedNull: false,
			wantOK:       true,
		},
		{
			name:         "fail - not null when null expected",
			row:          testRow{Data: []byte(`{"name": "test"}`)},
			path:         "name",
			expectedNull: true,
			wantOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildJSONFieldNullExpectation(JSONFieldNullExpectationConfig[testRow]{
				Path:         tt.path,
				ExpectName:   "test",
				GetJSON:      getJSON,
				ExpectedNull: tt.expectedNull,
			})

			result := exp.Check(nil, tt.row)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildBytesJSONFieldExpectation_Check(t *testing.T) {
	tests := []struct {
		name   string
		bytes  []byte
		path   string
		check  JSONCheck
		wantOK bool
	}{
		{
			name:   "success",
			bytes:  []byte(`{"name": "test"}`),
			path:   "name",
			check:  JSONCheckEquals("test"),
			wantOK: true,
		},
		{
			name:   "fail - path not found",
			bytes:  []byte(`{"name": "test"}`),
			path:   "missing",
			check:  JSONCheckEquals("test"),
			wantOK: false,
		},
		{
			name:   "fail - empty bytes",
			bytes:  []byte{},
			path:   "name",
			check:  JSONCheckEquals("test"),
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildBytesJSONFieldExpectation(BytesJSONFieldExpectationConfig{
				Path:       tt.path,
				ExpectName: "test",
				Check:      tt.check,
			})

			result := exp.Check(nil, tt.bytes)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildBytesJSONFieldWithExistsCheck_Check(t *testing.T) {
	tests := []struct {
		name          string
		bytes         []byte
		path          string
		requireExists bool
		check         JSONCheck
		wantOK        bool
	}{
		{
			name:          "success - exists and matches",
			bytes:         []byte(`{"name": "test"}`),
			path:          "name",
			requireExists: true,
			check:         JSONCheckEquals("test"),
			wantOK:        true,
		},
		{
			name:          "success - not exists and not required",
			bytes:         []byte(`{"name": "test"}`),
			path:          "missing",
			requireExists: false,
			check:         JSONCheckEquals("test"),
			wantOK:        true,
		},
		{
			name:          "fail - not exists but required",
			bytes:         []byte(`{"name": "test"}`),
			path:          "missing",
			requireExists: true,
			check:         JSONCheckEquals("test"),
			wantOK:        false,
		},
		{
			name:          "fail - empty bytes",
			bytes:         []byte{},
			path:          "name",
			requireExists: true,
			check:         JSONCheckEquals("test"),
			wantOK:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildBytesJSONFieldWithExistsCheck(BytesJSONFieldExistsCheckConfig{
				Path:          tt.path,
				ExpectName:    "test",
				RequireExists: tt.requireExists,
				Check:         tt.check,
			})

			result := exp.Check(nil, tt.bytes)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildBytesJSONFieldNullCheck_Check(t *testing.T) {
	tests := []struct {
		name         string
		bytes        []byte
		path         string
		expectedNull bool
		wantOK       bool
	}{
		{
			name:         "success - null when expected",
			bytes:        []byte(`{"name": null}`),
			path:         "name",
			expectedNull: true,
			wantOK:       true,
		},
		{
			name:         "success - field missing when null expected",
			bytes:        []byte(`{"other": "value"}`),
			path:         "name",
			expectedNull: true,
			wantOK:       true,
		},
		{
			name:         "success - not null when expected",
			bytes:        []byte(`{"name": "test"}`),
			path:         "name",
			expectedNull: false,
			wantOK:       true,
		},
		{
			name:         "fail - not null when null expected",
			bytes:        []byte(`{"name": "test"}`),
			path:         "name",
			expectedNull: true,
			wantOK:       false,
		},
		{
			name:         "fail - null when not expected",
			bytes:        []byte(`{"name": null}`),
			path:         "name",
			expectedNull: false,
			wantOK:       false,
		},
		{
			name:         "fail - empty bytes",
			bytes:        []byte{},
			path:         "name",
			expectedNull: true,
			wantOK:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildBytesJSONFieldNullCheck(BytesJSONFieldNullCheckConfig{
				Path:         tt.path,
				ExpectName:   "test",
				ExpectedNull: tt.expectedNull,
			})

			result := exp.Check(nil, tt.bytes)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildFullObjectExpectation_Check(t *testing.T) {
	compareExact := func(jsonObj gjson.Result, expected any) (bool, string) {
		return true, ""
	}
	compareFail := func(jsonObj gjson.Result, expected any) (bool, string) {
		return false, "mismatch"
	}

	tests := []struct {
		name      string
		row       testRow
		compare   ObjectCompareFunc
		wantOK    bool
		retryable bool
	}{
		{
			name:    "success",
			row:     testRow{Data: []byte(`{"name": "test"}`)},
			compare: compareExact,
			wantOK:  true,
		},
		{
			name:      "fail - compare fails",
			row:       testRow{Data: []byte(`{"name": "test"}`)},
			compare:   compareFail,
			wantOK:    false,
			retryable: true,
		},
		{
			name:      "fail - invalid json",
			row:       testRow{Data: []byte(`not json`)},
			compare:   compareExact,
			wantOK:    false,
			retryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildFullObjectExpectation(FullObjectExpectationConfig[testRow]{
				ExpectName: "test",
				GetJSON:    getJSON,
				Expected:   map[string]any{"name": "test"},
				Compare:    tt.compare,
				Retryable:  tt.retryable,
			})

			result := exp.Check(nil, tt.row)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestBuildFullObjectExpectation_WithPreCheck(t *testing.T) {
	preCheckFail := func(err error, row testRow) (polling.CheckResult, bool) {
		return polling.CheckResult{Ok: false, Reason: "precheck failed"}, false
	}

	exp := BuildFullObjectExpectation(FullObjectExpectationConfig[testRow]{
		ExpectName: "test",
		GetJSON:    getJSON,
		PreCheck:   preCheckFail,
		Expected:   map[string]any{},
		Compare:    func(jsonObj gjson.Result, expected any) (bool, string) { return true, "" },
	})

	result := exp.Check(nil, testRow{Data: []byte(`{}`)})
	if result.Ok {
		t.Error("Expected precheck to fail")
	}
}

func TestBuildFullObjectExpectation_GetJSONError(t *testing.T) {
	getJSONError := func(row testRow) ([]byte, error) {
		return nil, errors.New("failed")
	}

	exp := BuildFullObjectExpectation(FullObjectExpectationConfig[testRow]{
		ExpectName: "test",
		GetJSON:    getJSONError,
		Expected:   map[string]any{},
		Compare:    func(jsonObj gjson.Result, expected any) (bool, string) { return true, "" },
		Retryable:  true,
	})

	result := exp.Check(nil, testRow{})
	if result.Ok {
		t.Error("Expected failure when GetJSON returns error")
	}
	if !result.Retryable {
		t.Error("Expected retryable")
	}
}

func TestBuildBytesObjectExpectation_Check(t *testing.T) {
	compareExact := func(jsonObj gjson.Result, expected any) (bool, string) {
		return true, ""
	}
	compareFail := func(jsonObj gjson.Result, expected any) (bool, string) {
		return false, "mismatch"
	}

	tests := []struct {
		name    string
		bytes   []byte
		compare ObjectCompareFunc
		wantOK  bool
	}{
		{
			name:    "success",
			bytes:   []byte(`{"name": "test"}`),
			compare: compareExact,
			wantOK:  true,
		},
		{
			name:    "fail - compare fails",
			bytes:   []byte(`{"name": "test"}`),
			compare: compareFail,
			wantOK:  false,
		},
		{
			name:    "fail - empty bytes",
			bytes:   []byte{},
			compare: compareExact,
			wantOK:  false,
		},
		{
			name:    "fail - invalid json",
			bytes:   []byte(`not json`),
			compare: compareExact,
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := BuildBytesObjectExpectation(BytesObjectExpectationConfig{
				ExpectName: "test",
				Expected:   map[string]any{"name": "test"},
				Compare:    tt.compare,
			})

			result := exp.Check(nil, tt.bytes)
			if result.Ok != tt.wantOK {
				t.Errorf("Check() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}
