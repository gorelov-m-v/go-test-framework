package dsl

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEqualsLoose_BoolLikeVsNumericLike(t *testing.T) {
	tests := []struct {
		name         string
		expected     any
		actual       any
		wantEqual    bool
		wantRetry    bool
		wantContains string
	}{
		{
			name:      "sql.NullBool(true) vs int64(1) - should match",
			expected:  sql.NullBool{Valid: true, Bool: true},
			actual:    int64(1),
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:      "sql.NullBool(false) vs int64(0) - should match",
			expected:  sql.NullBool{Valid: true, Bool: false},
			actual:    int64(0),
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:         "sql.NullBool(true) vs int64(0) - should not match",
			expected:     sql.NullBool{Valid: true, Bool: true},
			actual:       int64(0),
			wantEqual:    false,
			wantRetry:    true,
			wantContains: "expected true, got false",
		},

		{
			name:      "int(1) vs bool(true) - should match",
			expected:  int(1),
			actual:    true,
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:      "int(0) vs bool(false) - should match",
			expected:  int(0),
			actual:    false,
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:         "int(1) vs bool(false) - should not match",
			expected:     int(1),
			actual:       false,
			wantEqual:    false,
			wantRetry:    true,
			wantContains: "expected true, got false",
		},
		{
			name:      "bool(true) vs sql.NullBool(true) - should match",
			expected:  true,
			actual:    sql.NullBool{Valid: true, Bool: true},
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:      "bool(true) vs int64(1) - should match",
			expected:  true,
			actual:    int64(1),
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:      "*sql.NullBool(true) vs int(1) - should match",
			expected:  &sql.NullBool{Valid: true, Bool: true},
			actual:    int(1),
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:      "int32(1) vs bool(true) - should match",
			expected:  int32(1),
			actual:    true,
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:      "uint64(0) vs bool(false) - should match",
			expected:  uint64(0),
			actual:    false,
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:         "int(2) vs bool(true) - should not match (2 is not bool-like)",
			expected:     int(2),
			actual:       true,
			wantEqual:    false,
			wantRetry:    false,
			wantContains: "type mismatch",
		},
		{
			name:         "bool(true) vs int(5) - should not match (5 is not bool-like)",
			expected:     true,
			actual:       int(5),
			wantEqual:    false,
			wantRetry:    false,
			wantContains: "type mismatch",
		},
		{
			name:      "sql.NullInt64(1) vs bool(true) - should match",
			expected:  sql.NullInt64{Valid: true, Int64: 1},
			actual:    true,
			wantEqual: true,
			wantRetry: true,
		},
		{
			name:      "bool(false) vs sql.NullInt64(0) - should match",
			expected:  false,
			actual:    sql.NullInt64{Valid: true, Int64: 0},
			wantEqual: true,
			wantRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEqual, gotRetry, reason := equalsLoose(tt.expected, tt.actual)

			assert.Equal(t, tt.wantEqual, gotEqual, "equal mismatch: %s", reason)
			assert.Equal(t, tt.wantRetry, gotRetry, "retryable mismatch: %s", reason)

			if tt.wantContains != "" {
				assert.Contains(t, reason, tt.wantContains, "reason should contain expected substring")
			}
		})
	}
}

func TestEqualsLoose_BoolComparisonPriority(t *testing.T) {
	t.Run("bool-like comparison takes precedence over numeric", func(t *testing.T) {
		equal, retryable, reason := equalsLoose(int(1), true)

		assert.True(t, equal, "int(1) should equal true via bool conversion")
		assert.True(t, retryable, "should be retryable")
		assert.NotContains(t, reason, "type mismatch", "should not have type mismatch error")
	})

	t.Run("non-bool-like numerics still work", func(t *testing.T) {
		equal, retryable, _ := equalsLoose(int(42), int64(42))

		assert.True(t, equal, "numeric comparison should work for non-bool values")
		assert.True(t, retryable)
	})

	t.Run("bool-like values that don't match fail gracefully", func(t *testing.T) {
		equal, retryable, reason := equalsLoose(true, int(0))

		assert.False(t, equal, "true should not equal 0")
		assert.True(t, retryable, "should be retryable (not a type error)")
		assert.Contains(t, reason, "expected true, got false")
	})
}
