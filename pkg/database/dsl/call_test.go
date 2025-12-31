package dsl

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestUser struct {
	ID        int            `db:"id"`
	Username  string         `db:"username"`
	Email     sql.NullString `db:"email"`
	IsActive  int64          `db:"is_active"`
	CreatedAt time.Time      `db:"created_at"`
	DeletedAt *time.Time     `db:"deleted_at"`
}

func TestGetFieldValueByColumnName_Success(t *testing.T) {
	user := TestUser{
		ID:       1,
		Username: "testuser",
		Email:    sql.NullString{String: "test@example.com", Valid: true},
		IsActive: 1,
	}

	val, err := getFieldValueByColumnName(user, "id")
	require.NoError(t, err)
	assert.Equal(t, 1, val)

	val, err = getFieldValueByColumnName(user, "username")
	require.NoError(t, err)
	assert.Equal(t, "testuser", val)

	val, err = getFieldValueByColumnName(user, "email")
	require.NoError(t, err)
	nullStr, ok := val.(sql.NullString)
	require.True(t, ok)
	assert.True(t, nullStr.Valid)
	assert.Equal(t, "test@example.com", nullStr.String)

	val, err = getFieldValueByColumnName(user, "is_active")
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)
}

func TestGetFieldValueByColumnName_PointerStruct(t *testing.T) {
	user := &TestUser{
		ID:       2,
		Username: "ptruser",
	}

	val, err := getFieldValueByColumnName(user, "id")
	require.NoError(t, err)
	assert.Equal(t, 2, val)

	val, err = getFieldValueByColumnName(user, "username")
	require.NoError(t, err)
	assert.Equal(t, "ptruser", val)
}

func TestGetFieldValueByColumnName_NotFound(t *testing.T) {
	user := TestUser{ID: 1}

	_, err := getFieldValueByColumnName(user, "nonexistent_column")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no field with db tag 'nonexistent_column' found")
}

func TestGetFieldValueByColumnName_NullPointer(t *testing.T) {
	now := time.Now()
	user := TestUser{
		ID:        1,
		DeletedAt: &now,
	}

	val, err := getFieldValueByColumnName(user, "deleted_at")
	require.NoError(t, err)
	deletedAt, ok := val.(*time.Time)
	require.True(t, ok)
	assert.NotNil(t, deletedAt)
	assert.Equal(t, now, *deletedAt)

	user.DeletedAt = nil
	val, err = getFieldValueByColumnName(user, "deleted_at")
	require.NoError(t, err)
	deletedAt, ok = val.(*time.Time)
	require.True(t, ok)
	assert.Nil(t, deletedAt)
}

func TestGetFieldValueByColumnName_NullStringInvalid(t *testing.T) {
	user := TestUser{
		ID:    1,
		Email: sql.NullString{Valid: false},
	}

	val, err := getFieldValueByColumnName(user, "email")
	require.NoError(t, err)
	nullStr, ok := val.(sql.NullString)
	require.True(t, ok)
	assert.False(t, nullStr.Valid)
}

func TestGetFieldValueByColumnName_InvalidTarget(t *testing.T) {
	_, err := getFieldValueByColumnName("not a struct", "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target is not a struct")

	_, err = getFieldValueByColumnName(123, "id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target is not a struct")
}

func TestGetFieldValueByColumnName_TagWithOptions(t *testing.T) {
	type UserWithOptions struct {
		ID   int    `db:"id,omitempty"`
		Name string `db:"name,primarykey"`
	}

	user := UserWithOptions{
		ID:   42,
		Name: "test",
	}

	val, err := getFieldValueByColumnName(user, "id")
	require.NoError(t, err)
	assert.Equal(t, 42, val)

	val, err = getFieldValueByColumnName(user, "name")
	require.NoError(t, err)
	assert.Equal(t, "test", val)
}

func TestGetFirstLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line",
			input:    "SELECT * FROM users",
			expected: "SELECT * FROM users",
		},
		{
			name:     "multi line",
			input:    "SELECT *\nFROM users\nWHERE id = 1",
			expected: "SELECT *",
		},
		{
			name:     "with leading spaces",
			input:    "  SELECT * FROM users  ",
			expected: "SELECT * FROM users",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFirstLine(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "simple SELECT",
			query:    "SELECT * FROM users",
			expected: "users",
		},
		{
			name:     "SELECT with WHERE",
			query:    "SELECT * FROM users WHERE id = 1",
			expected: "users",
		},
		{
			name:     "SELECT with backticks",
			query:    "SELECT * FROM `users`",
			expected: "users",
		},
		{
			name:     "SELECT with database prefix",
			query:    "SELECT * FROM database.users",
			expected: "users",
		},
		{
			name:     "SELECT with database prefix and backticks",
			query:    "SELECT * FROM `beta-09_core`.game_category",
			expected: "game_category",
		},
		{
			name:     "INSERT INTO",
			query:    "INSERT INTO users (name) VALUES ('test')",
			expected: "users",
		},
		{
			name:     "INSERT INTO with backticks",
			query:    "INSERT INTO `users` (name) VALUES ('test')",
			expected: "users",
		},
		{
			name:     "UPDATE",
			query:    "UPDATE users SET name = 'test'",
			expected: "users",
		},
		{
			name:     "UPDATE with backticks",
			query:    "UPDATE `users` SET name = 'test'",
			expected: "users",
		},
		{
			name:     "UPDATE with database prefix",
			query:    "UPDATE database.users SET name = 'test'",
			expected: "users",
		},
		{
			name:     "DELETE FROM",
			query:    "DELETE FROM users WHERE id = 1",
			expected: "users",
		},
		{
			name:     "multi-line query",
			query:    "SELECT * FROM users WHERE id = 1",
			expected: "users",
		},
		{
			name:     "query with JOIN",
			query:    "SELECT * FROM users JOIN orders ON users.id = orders.user_id",
			expected: "users",
		},
		{
			name:     "unknown query without FROM",
			query:    "SHOW TABLES",
			expected: "query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTableName(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractOperation(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "SELECT",
			query:    "SELECT * FROM users",
			expected: "SELECT",
		},
		{
			name:     "select lowercase",
			query:    "select * from users",
			expected: "SELECT",
		},
		{
			name:     "INSERT",
			query:    "INSERT INTO users (name) VALUES ('test')",
			expected: "INSERT",
		},
		{
			name:     "UPDATE",
			query:    "UPDATE users SET name = 'test'",
			expected: "UPDATE",
		},
		{
			name:     "DELETE",
			query:    "DELETE FROM users WHERE id = 1",
			expected: "DELETE",
		},
		{
			name:     "with leading spaces",
			query:    "  SELECT * FROM users",
			expected: "SELECT",
		},
		{
			name:     "multi-line",
			query:    "SELECT *\nFROM users",
			expected: "SELECT",
		},
		{
			name:     "unknown operation",
			query:    "EXPLAIN SELECT * FROM users",
			expected: "EXEC",
		},
		{
			name:     "TRUNCATE",
			query:    "TRUNCATE TABLE users",
			expected: "EXEC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractOperation(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}
