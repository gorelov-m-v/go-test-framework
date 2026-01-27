package expect

import (
	"database/sql"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

func TestCheckEquals(t *testing.T) {
	simpleEquals := func(expected, actual any) (bool, bool, string) {
		if expected == actual {
			return true, false, ""
		}
		return false, true, "values not equal"
	}

	tests := []struct {
		name     string
		expected any
		actual   any
		column   string
		wantOK   bool
	}{
		{"equal strings", "hello", "hello", "col", true},
		{"not equal strings", "hello", "world", "col", false},
		{"equal ints", 42, 42, "col", true},
		{"not equal ints", 42, 100, "col", false},
		{"equal nil", nil, nil, "col", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckEquals(tt.expected, simpleEquals)
			result := check(tt.actual, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckEquals() Ok = %v, want %v", result.Ok, tt.wantOK)
			}
			if !result.Ok && result.Reason == "" {
				t.Error("CheckEquals() should have reason when not ok")
			}
		})
	}
}

func TestCheckNotEquals(t *testing.T) {
	simpleEquals := func(expected, actual any) (bool, bool, string) {
		if expected == actual {
			return true, false, ""
		}
		return false, true, "values not equal"
	}

	tests := []struct {
		name        string
		notExpected any
		actual      any
		column      string
		wantOK      bool
	}{
		{"different strings", "hello", "world", "col", true},
		{"same strings", "hello", "hello", "col", false},
		{"different ints", 42, 100, "col", true},
		{"same ints", 42, 42, "col", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckNotEquals(tt.notExpected, simpleEquals)
			result := check(tt.actual, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckNotEquals() Ok = %v, want %v", result.Ok, tt.wantOK)
			}
		})
	}
}

func TestCheckNotEmpty(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		column string
		wantOK bool
	}{
		{"non-empty string", "hello", "col", true},
		{"empty string", "", "col", false},
		{"whitespace string", "   ", "col", false},
		{"nil", nil, "col", false},
		{"non-empty slice", []int{1, 2}, "col", true},
		{"empty slice", []int{}, "col", false},
		{"sql.NullString valid", sql.NullString{Valid: true, String: "hello"}, "col", true},
		{"sql.NullString invalid", sql.NullString{Valid: false}, "col", false},
		{"non-zero int", 42, "col", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckNotEmpty()
			result := check(tt.value, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckNotEmpty() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestCheckEmpty(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		column string
		wantOK bool
	}{
		{"empty string", "", "col", true},
		{"whitespace string", "   ", "col", true},
		{"non-empty string", "hello", "col", false},
		{"nil", nil, "col", true},
		{"empty slice", []int{}, "col", true},
		{"non-empty slice", []int{1, 2}, "col", false},
		{"sql.NullString invalid", sql.NullString{Valid: false}, "col", true},
		{"sql.NullString valid non-empty", sql.NullString{Valid: true, String: "hello"}, "col", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckEmpty()
			result := check(tt.value, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckEmpty() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestCheckIsNull(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		column string
		wantOK bool
	}{
		{"nil", nil, "col", true},
		{"sql.NullString invalid", sql.NullString{Valid: false}, "col", true},
		{"sql.NullString valid", sql.NullString{Valid: true, String: ""}, "col", false},
		{"sql.NullInt64 invalid", sql.NullInt64{Valid: false}, "col", true},
		{"sql.NullInt64 valid", sql.NullInt64{Valid: true, Int64: 0}, "col", false},
		{"nil pointer", (*int)(nil), "col", true},
		{"non-nil pointer", ptrInt(42), "col", false},
		{"string value", "hello", "col", false},
		{"int value", 42, "col", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckIsNull()
			result := check(tt.value, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckIsNull() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestCheckIsNotNull(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		column string
		wantOK bool
	}{
		{"nil", nil, "col", false},
		{"sql.NullString invalid", sql.NullString{Valid: false}, "col", false},
		{"sql.NullString valid", sql.NullString{Valid: true, String: ""}, "col", true},
		{"sql.NullInt64 valid", sql.NullInt64{Valid: true, Int64: 42}, "col", true},
		{"non-nil pointer", ptrInt(42), "col", true},
		{"string value", "hello", "col", true},
		{"int value", 42, "col", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckIsNotNull()
			result := check(tt.value, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckIsNotNull() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestCheckTrue(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		column    string
		wantOK    bool
		wantRetry bool
	}{
		{"bool true", true, "col", true, false},
		{"bool false", false, "col", false, true},
		{"int 1", int(1), "col", true, false},
		{"int 0", int(0), "col", false, true},
		{"int 2", int(2), "col", false, false},
		{"sql.NullBool valid true", sql.NullBool{Valid: true, Bool: true}, "col", true, false},
		{"sql.NullBool valid false", sql.NullBool{Valid: true, Bool: false}, "col", false, true},
		{"sql.NullBool invalid", sql.NullBool{Valid: false}, "col", false, true},
		{"nil", nil, "col", false, true},
		{"string", "true", "col", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckTrue()
			result := check(tt.value, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckTrue() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
			if !result.Ok && result.Retryable != tt.wantRetry {
				t.Errorf("CheckTrue() Retryable = %v, want %v", result.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestCheckFalse(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		column    string
		wantOK    bool
		wantRetry bool
	}{
		{"bool false", false, "col", true, false},
		{"bool true", true, "col", false, true},
		{"int 0", int(0), "col", true, false},
		{"int 1", int(1), "col", false, true},
		{"int 2", int(2), "col", false, false},
		{"sql.NullBool valid false", sql.NullBool{Valid: true, Bool: false}, "col", true, false},
		{"sql.NullBool valid true", sql.NullBool{Valid: true, Bool: true}, "col", false, true},
		{"sql.NullBool invalid", sql.NullBool{Valid: false}, "col", false, true},
		{"nil", nil, "col", false, true},
		{"string", "false", "col", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := CheckFalse()
			result := check(tt.value, tt.column)
			if result.Ok != tt.wantOK {
				t.Errorf("CheckFalse() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
			if !result.Ok && result.Retryable != tt.wantRetry {
				t.Errorf("CheckFalse() Retryable = %v, want %v", result.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestJSONCheckEquals(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		path     string
		expected any
		wantOK   bool
	}{
		{"equal string", `"hello"`, "", "hello", true},
		{"not equal string", `"hello"`, "", "world", false},
		{"equal int", `42`, "", 42, true},
		{"equal int as float", `42`, "", float64(42), true},
		{"not equal int", `42`, "", 100, false},
		{"equal bool true", `true`, "", true, true},
		{"equal bool false", `false`, "", false, true},
		{"not equal bool", `true`, "", false, false},
		{"equal null", `null`, "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			check := JSONCheckEquals(tt.expected)
			result := check(res, tt.path)
			if result.Ok != tt.wantOK {
				t.Errorf("JSONCheckEquals() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestJSONCheckNotEmpty(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		path   string
		wantOK bool
	}{
		{"non-empty string", `"hello"`, "", true},
		{"empty string", `""`, "", false},
		{"non-empty array", `[1,2,3]`, "", true},
		{"empty array", `[]`, "", false},
		{"non-empty object", `{"a":1}`, "", true},
		{"empty object", `{}`, "", false},
		{"number", `42`, "", true},
		{"zero", `0`, "", true},
		{"null", `null`, "", false},
		{"bool true", `true`, "", true},
		{"bool false", `false`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			check := JSONCheckNotEmpty()
			result := check(res, tt.path)
			if result.Ok != tt.wantOK {
				t.Errorf("JSONCheckNotEmpty() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestJSONCheckEmpty(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		path   string
		wantOK bool
	}{
		{"empty string", `""`, "", true},
		{"non-empty string", `"hello"`, "", false},
		{"empty array", `[]`, "", true},
		{"non-empty array", `[1,2,3]`, "", false},
		{"empty object", `{}`, "", true},
		{"non-empty object", `{"a":1}`, "", false},
		{"null", `null`, "", true},
		{"number", `42`, "", false},
		{"zero", `0`, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			check := JSONCheckEmpty()
			result := check(res, tt.path)
			if result.Ok != tt.wantOK {
				t.Errorf("JSONCheckEmpty() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestJSONCheckIsNull(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		wantOK bool
	}{
		{"null", `null`, true},
		{"string", `"hello"`, false},
		{"empty string", `""`, false},
		{"number", `42`, false},
		{"zero", `0`, false},
		{"bool true", `true`, false},
		{"bool false", `false`, false},
		{"array", `[]`, false},
		{"object", `{}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			check := JSONCheckIsNull()
			result := check(res, "path")
			if result.Ok != tt.wantOK {
				t.Errorf("JSONCheckIsNull() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestJSONCheckIsNotNull(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		wantOK bool
	}{
		{"null", `null`, false},
		{"string", `"hello"`, true},
		{"empty string", `""`, true},
		{"number", `42`, true},
		{"zero", `0`, true},
		{"bool true", `true`, true},
		{"bool false", `false`, true},
		{"array", `[]`, true},
		{"object", `{}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			check := JSONCheckIsNotNull()
			result := check(res, "path")
			if result.Ok != tt.wantOK {
				t.Errorf("JSONCheckIsNotNull() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
		})
	}
}

func TestJSONCheckTrue(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantOK    bool
		wantRetry bool
	}{
		{"true", `true`, true, false},
		{"false", `false`, false, true},
		{"string", `"true"`, false, false},
		{"number", `1`, false, false},
		{"null", `null`, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			check := JSONCheckTrue()
			result := check(res, "path")
			if result.Ok != tt.wantOK {
				t.Errorf("JSONCheckTrue() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
			if !result.Ok && result.Retryable != tt.wantRetry {
				t.Errorf("JSONCheckTrue() Retryable = %v, want %v", result.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestJSONCheckFalse(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantOK    bool
		wantRetry bool
	}{
		{"false", `false`, true, false},
		{"true", `true`, false, true},
		{"string", `"false"`, false, false},
		{"number", `0`, false, false},
		{"null", `null`, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := gjson.Parse(tt.json)
			check := JSONCheckFalse()
			result := check(res, "path")
			if result.Ok != tt.wantOK {
				t.Errorf("JSONCheckFalse() Ok = %v, want %v, reason: %s", result.Ok, tt.wantOK, result.Reason)
			}
			if !result.Ok && result.Retryable != tt.wantRetry {
				t.Errorf("JSONCheckFalse() Retryable = %v, want %v", result.Retryable, tt.wantRetry)
			}
		})
	}
}

func TestCheckResultRetryable(t *testing.T) {
	t.Run("CheckEquals not retryable on type mismatch", func(t *testing.T) {
		equalsWithTypeMismatch := func(expected, actual any) (bool, bool, string) {
			return false, false, "type mismatch"
		}
		check := CheckEquals("hello", equalsWithTypeMismatch)
		result := check(42, "col")
		if result.Retryable {
			t.Error("Expected non-retryable result for type mismatch")
		}
	})

	t.Run("CheckNotEquals always retryable", func(t *testing.T) {
		simpleEquals := func(expected, actual any) (bool, bool, string) {
			return expected == actual, false, ""
		}
		check := CheckNotEquals("hello", simpleEquals)
		result := check("hello", "col")
		if !result.Retryable {
			t.Error("CheckNotEquals should be retryable")
		}
	})
}

func TestCheckResultReason(t *testing.T) {
	tests := []struct {
		name         string
		check        ValueCheck
		value        any
		column       string
		wantContains string
	}{
		{"CheckNotEmpty reason", CheckNotEmpty(), "", "myCol", "myCol"},
		{"CheckEmpty reason", CheckEmpty(), "value", "myCol", "myCol"},
		{"CheckIsNull reason", CheckIsNull(), "value", "myCol", "myCol"},
		{"CheckIsNotNull reason", CheckIsNotNull(), nil, "myCol", "myCol"},
		{"CheckTrue reason for false", CheckTrue(), false, "myCol", "myCol"},
		{"CheckFalse reason for true", CheckFalse(), true, "myCol", "myCol"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.check(tt.value, tt.column)
			if result.Ok {
				t.Error("Expected check to fail")
				return
			}
			if !contains(result.Reason, tt.wantContains) {
				t.Errorf("Reason %q should contain %q", result.Reason, tt.wantContains)
			}
		})
	}
}

func ptrInt(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestCheckResultFields(t *testing.T) {
	t.Run("successful check has Ok=true", func(t *testing.T) {
		check := CheckNotEmpty()
		result := check("hello", "col")
		if !result.Ok {
			t.Error("Expected Ok=true")
		}
		if result.Reason != "" {
			t.Error("Expected empty reason for successful check")
		}
	})

	t.Run("failed check has Ok=false and reason", func(t *testing.T) {
		check := CheckNotEmpty()
		result := check("", "col")
		if result.Ok {
			t.Error("Expected Ok=false")
		}
		if result.Reason == "" {
			t.Error("Expected non-empty reason for failed check")
		}
	})
}

var _ polling.CheckResult
