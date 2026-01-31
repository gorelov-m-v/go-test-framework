package dsl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

func TestExpectResponseStatus(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		resp           *client.Response[any]
		err            error
		wantOk         bool
		wantRetryable  bool
		wantContains   string
	}{
		{
			name:           "status matches",
			expectedStatus: 200,
			resp:           &client.Response[any]{StatusCode: 200},
			wantOk:         true,
		},
		{
			name:           "status mismatches",
			expectedStatus: 200,
			resp:           &client.Response[any]{StatusCode: 500},
			wantOk:         false,
			wantRetryable:  true,
			wantContains:   "500",
		},
		{
			name:           "request error",
			expectedStatus: 200,
			resp:           &client.Response[any]{StatusCode: 200},
			err:            errors.New("connection refused"),
			wantOk:         false,
			wantRetryable:  true,
			wantContains:   "Request failed",
		},
		{
			name:           "response nil",
			expectedStatus: 200,
			resp:           nil,
			wantOk:         false,
			wantRetryable:  true,
			wantContains:   "nil",
		},
		{
			name:           "network error",
			expectedStatus: 200,
			resp:           &client.Response[any]{StatusCode: 0, NetworkError: "timeout"},
			wantOk:         false,
			wantRetryable:  true,
			wantContains:   "Network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := makeResponseStatusExpectation(tt.expectedStatus)

			result := exp.Check(tt.err, tt.resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectResponseStatus_Name(t *testing.T) {
	exp := makeResponseStatusExpectation(201)

	assert.Contains(t, exp.Name, "201")
	assert.Contains(t, exp.Name, "Created")
}

func TestExpectResponseBodyNotEmpty(t *testing.T) {
	tests := []struct {
		name          string
		resp          *client.Response[any]
		err           error
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:   "body present",
			resp:   &client.Response[any]{RawBody: []byte(`{"id": 1}`)},
			wantOk: true,
		},
		{
			name:          "body empty",
			resp:          &client.Response[any]{RawBody: []byte{}},
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "empty",
		},
		{
			name:          "body nil",
			resp:          &client.Response[any]{RawBody: nil},
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "request error",
			resp:          &client.Response[any]{RawBody: []byte(`{"id": 1}`)},
			err:           errors.New("failed"),
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "response nil",
			resp:          nil,
			wantOk:        false,
			wantRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := makeResponseBodyNotEmptyExpectation()

			result := exp.Check(tt.err, tt.resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectFieldTrue(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:   "value is true",
			path:   "active",
			json:   `{"active": true}`,
			wantOk: true,
		},
		{
			name:          "value is false",
			path:          "active",
			json:          `{"active": false}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "false",
		},
		{
			name:          "value is not boolean",
			path:          "count",
			json:          `{"count": 123}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "boolean",
		},
		{
			name:          "field missing",
			path:          "active",
			json:          `{"name": "test"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "does not exist",
		},
		{
			name:          "invalid JSON",
			path:          "active",
			json:          `not json`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "empty path",
			path:          "",
			json:          `{"active": true}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Invalid JSON path",
		},
		{
			name:   "nested path",
			path:   "user.isActive",
			json:   `{"user": {"isActive": true}}`,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.FieldTrue(tt.path)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectFieldFalse(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:   "value is false",
			path:   "deleted",
			json:   `{"deleted": false}`,
			wantOk: true,
		},
		{
			name:          "value is true",
			path:          "deleted",
			json:          `{"deleted": true}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "true",
		},
		{
			name:          "value is string",
			path:          "status",
			json:          `{"status": "inactive"}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "boolean",
		},
		{
			name:          "field missing",
			path:          "deleted",
			json:          `{"name": "test"}`,
			wantOk:        false,
			wantRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.FieldFalse(tt.path)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectFieldIsNull(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:   "value is null",
			path:   "deletedAt",
			json:   `{"deletedAt": null}`,
			wantOk: true,
		},
		{
			name:          "value is string",
			path:          "deletedAt",
			json:          `{"deletedAt": "2024-01-01"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "string",
		},
		{
			name:          "value is number",
			path:          "count",
			json:          `{"count": 0}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "field missing",
			path:          "deletedAt",
			json:          `{"name": "test"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "does not exist",
		},
		{
			name:          "empty path",
			path:          "",
			json:          `{"deletedAt": null}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Invalid JSON path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.FieldIsNull(tt.path)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectFieldIsNotNull(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:   "value is string",
			path:   "createdAt",
			json:   `{"createdAt": "2024-01-01"}`,
			wantOk: true,
		},
		{
			name:          "value is null",
			path:          "createdAt",
			json:          `{"createdAt": null}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "null",
		},
		{
			name:   "value is zero",
			path:   "count",
			json:   `{"count": 0}`,
			wantOk: true,
		},
		{
			name:   "value is empty string",
			path:   "name",
			json:   `{"name": ""}`,
			wantOk: true,
		},
		{
			name:          "field missing",
			path:          "createdAt",
			json:          `{"name": "test"}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:   "nested path",
			path:   "user.email",
			json:   `{"user": {"email": "test@test.com"}}`,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.FieldIsNotNull(tt.path)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectFieldValue(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expected      any
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "string matches",
			path:     "name",
			expected: "John",
			json:     `{"name": "John"}`,
			wantOk:   true,
		},
		{
			name:          "string mismatches",
			path:          "name",
			expected:      "John",
			json:          `{"name": "Jane"}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:     "int matches",
			path:     "count",
			expected: 42,
			json:     `{"count": 42}`,
			wantOk:   true,
		},
		{
			name:          "int mismatches",
			path:          "count",
			expected:      42,
			json:          `{"count": 100}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:     "bool matches",
			path:     "active",
			expected: true,
			json:     `{"active": true}`,
			wantOk:   true,
		},
		{
			name:          "bool mismatches",
			path:          "active",
			expected:      true,
			json:          `{"active": false}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:     "null matches",
			path:     "deletedAt",
			expected: nil,
			json:     `{"deletedAt": null}`,
			wantOk:   true,
		},
		{
			name:          "null mismatches",
			path:          "deletedAt",
			expected:      nil,
			json:          `{"deletedAt": "2024-01-01"}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "field missing",
			path:          "name",
			expected:      "John",
			json:          `{"other": "value"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "does not exist",
		},
		{
			name:          "empty path",
			path:          "",
			expected:      "value",
			json:          `{"name": "John"}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Invalid JSON path",
		},
		{
			name:     "nested path",
			path:     "user.email",
			expected: "test@test.com",
			json:     `{"user": {"email": "test@test.com"}}`,
			wantOk:   true,
		},
		{
			name:     "array index",
			path:     "items.0",
			expected: "first",
			json:     `{"items": ["first", "second"]}`,
			wantOk:   true,
		},
		{
			name:          "invalid JSON",
			path:          "name",
			expected:      "John",
			json:          `not json`,
			wantOk:        false,
			wantRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.FieldEquals(tt.path, tt.expected)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectFieldNotEmpty(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:   "string not empty",
			path:   "name",
			json:   `{"name": "John"}`,
			wantOk: true,
		},
		{
			name:          "empty string",
			path:          "name",
			json:          `{"name": ""}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "empty",
		},
		{
			name:          "whitespace string",
			path:          "name",
			json:          `{"name": "   "}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "null value",
			path:          "name",
			json:          `{"name": null}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "empty array",
			path:          "items",
			json:          `{"items": []}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:   "non-empty array",
			path:   "items",
			json:   `{"items": [1, 2, 3]}`,
			wantOk: true,
		},
		{
			name:          "empty object",
			path:          "data",
			json:          `{"data": {}}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:   "non-empty object",
			path:   "data",
			json:   `{"data": {"id": 1}}`,
			wantOk: true,
		},
		{
			name:   "zero number",
			path:   "count",
			json:   `{"count": 0}`,
			wantOk: true,
		},
		{
			name:   "false boolean",
			path:   "active",
			json:   `{"active": false}`,
			wantOk: true,
		},
		{
			name:          "field missing",
			path:          "name",
			json:          `{"other": "value"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "does not exist",
		},
		{
			name:          "empty path",
			path:          "",
			json:          `{"name": "John"}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Invalid JSON path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.FieldNotEmpty(tt.path)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectArrayContains(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expected      any
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "found with map",
			path:     "users",
			expected: map[string]any{"name": "John"},
			json:     `{"users": [{"name": "John", "age": 30}, {"name": "Jane", "age": 25}]}`,
			wantOk:   true,
		},
		{
			name:          "not found",
			path:          "users",
			expected:      map[string]any{"name": "Bob"},
			json:          `{"users": [{"name": "John"}, {"name": "Jane"}]}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "No matching",
		},
		{
			name:     "partial match",
			path:     "items",
			expected: map[string]any{"id": 1},
			json:     `{"items": [{"id": 1, "name": "Item1", "price": 100}]}`,
			wantOk:   true,
		},
		{
			name:          "empty array",
			path:          "items",
			expected:      map[string]any{"id": 1},
			json:          `{"items": []}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "not array",
			path:          "user",
			expected:      map[string]any{"name": "John"},
			json:          `{"user": {"name": "John"}}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Expected array",
		},
		{
			name:          "path missing",
			path:          "items",
			expected:      map[string]any{"id": 1},
			json:          `{"other": "value"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "does not exist",
		},
		{
			name:          "empty path",
			path:          "",
			expected:      map[string]any{"id": 1},
			json:          `[{"id": 1}]`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Invalid JSON path",
		},
		{
			name:     "nested path",
			path:     "data.items",
			expected: map[string]any{"id": 1},
			json:     `{"data": {"items": [{"id": 1, "name": "test"}]}}`,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.ArrayContains(tt.path, tt.expected)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok, "Reason: %s", result.Reason)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectArrayContains_WithStruct(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	exp := jsonSource.ArrayContains("items", Item{ID: 2})
	resp := &client.Response[any]{
		RawBody: []byte(`{"items": [{"id": 1, "name": "first"}, {"id": 2, "name": "second"}]}`),
	}

	result := exp.Check(nil, resp)

	assert.True(t, result.Ok)
}

func TestExpectArrayContainsExact(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expected      any
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "found exact match",
			path:     "users",
			expected: map[string]any{"name": "John", "age": 30},
			json:     `{"users": [{"name": "John", "age": 30}, {"name": "Jane", "age": 25}]}`,
			wantOk:   true,
		},
		{
			name:     "partial match with map",
			path:     "users",
			expected: map[string]any{"name": "John"},
			json:     `{"users": [{"name": "John", "age": 30}]}`,
			wantOk:   true,
		},
		{
			name:          "not found",
			path:          "users",
			expected:      map[string]any{"name": "Bob"},
			json:          `{"users": [{"name": "John"}, {"name": "Jane"}]}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "No matching",
		},
		{
			name:          "empty array",
			path:          "items",
			expected:      map[string]any{"id": 1},
			json:          `{"items": []}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "not array",
			path:          "user",
			expected:      map[string]any{"name": "John"},
			json:          `{"user": {"name": "John"}}`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Expected array",
		},
		{
			name:          "path missing",
			path:          "items",
			expected:      map[string]any{"id": 1},
			json:          `{"other": "value"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "does not exist",
		},
		{
			name:          "empty path",
			path:          "",
			expected:      map[string]any{"id": 1},
			json:          `[{"id": 1}]`,
			wantOk:        false,
			wantRetryable: false,
			wantContains:  "Invalid JSON path",
		},
		{
			name:     "nested path",
			path:     "data.items",
			expected: map[string]any{"id": 1, "name": "test"},
			json:     `{"data": {"items": [{"id": 1, "name": "test"}]}}`,
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.ArrayContainsExact(tt.path, tt.expected)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok, "Reason: %s", result.Reason)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectArrayContainsExact_WithStruct(t *testing.T) {
	type Item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name          string
		expected      Item
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "exact match",
			expected: Item{ID: 2, Name: "second"},
			json:     `{"items": [{"id": 1, "name": "first"}, {"id": 2, "name": "second"}]}`,
			wantOk:   true,
		},
		{
			name:          "zero value field mismatch",
			expected:      Item{ID: 2, Name: ""},
			json:          `{"items": [{"id": 2, "name": "second"}]}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "similar object",
		},
		{
			name:          "partial fields with zero value",
			expected:      Item{ID: 2},
			json:          `{"items": [{"id": 1, "name": "first"}, {"id": 2, "name": "second"}]}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "similar object",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.ArrayContainsExact("items", tt.expected)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok, "Reason: %s", result.Reason)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectResponseBody(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name          string
		expected      any
		json          string
		err           error
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "exact match struct",
			expected: User{ID: 1, Name: "John"},
			json:     `{"id": 1, "name": "John"}`,
			wantOk:   true,
		},
		{
			name:          "field mismatch",
			expected:      User{ID: 1, Name: "John"},
			json:          `{"id": 1, "name": "Jane"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "name",
		},
		{
			name:     "extra fields in JSON",
			expected: User{ID: 1, Name: "John"},
			json:     `{"id": 1, "name": "John", "email": "john@test.com"}`,
			wantOk:   true,
		},
		{
			name:          "zero value mismatch",
			expected:      User{ID: 1, Name: ""},
			json:          `{"id": 1, "name": "John"}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:     "map match",
			expected: map[string]any{"id": 1, "name": "John"},
			json:     `{"id": 1, "name": "John", "extra": "field"}`,
			wantOk:   true,
		},
		{
			name:          "map mismatch",
			expected:      map[string]any{"id": 1, "name": "John"},
			json:          `{"id": 1, "name": "Jane"}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "invalid JSON",
			expected:      User{ID: 1},
			json:          `not json`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "Invalid JSON",
		},
		{
			name:          "request error",
			expected:      User{ID: 1},
			json:          "",
			err:           errors.New("connection refused"),
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "empty body",
			expected:      User{ID: 1},
			json:          "",
			wantOk:        false,
			wantRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.BodyEquals(tt.expected)
			var resp *client.Response[any]
			if tt.err == nil {
				resp = &client.Response[any]{RawBody: []byte(tt.json)}
			}

			result := exp.Check(tt.err, resp)

			assert.Equal(t, tt.wantOk, result.Ok, "Reason: %s", result.Reason)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectResponseBody_NestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}
	type User struct {
		ID      int     `json:"id"`
		Address Address `json:"address"`
	}

	tests := []struct {
		name          string
		expected      User
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "nested match",
			expected: User{ID: 1, Address: Address{City: "Moscow"}},
			json:     `{"id": 1, "address": {"city": "Moscow"}}`,
			wantOk:   true,
		},
		{
			name:          "nested mismatch",
			expected:      User{ID: 1, Address: Address{City: "Moscow"}},
			json:          `{"id": 1, "address": {"city": "London"}}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.BodyEquals(tt.expected)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok, "Reason: %s", result.Reason)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectResponseBodyPartial(t *testing.T) {
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	tests := []struct {
		name          string
		expected      any
		json          string
		err           error
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "all fields match",
			expected: User{ID: 1, Name: "John"},
			json:     `{"id": 1, "name": "John", "email": "john@test.com"}`,
			wantOk:   true,
		},
		{
			name:     "partial fields match",
			expected: User{ID: 1},
			json:     `{"id": 1, "name": "John", "email": "john@test.com"}`,
			wantOk:   true,
		},
		{
			name:          "field mismatch",
			expected:      User{ID: 1, Name: "John"},
			json:          `{"id": 1, "name": "Jane"}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "name",
		},
		{
			name:     "zero values ignored",
			expected: User{ID: 1, Name: ""},
			json:     `{"id": 1, "name": "John"}`,
			wantOk:   true,
		},
		{
			name:     "map match",
			expected: map[string]any{"id": 1},
			json:     `{"id": 1, "name": "John", "email": "john@test.com"}`,
			wantOk:   true,
		},
		{
			name:          "map mismatch",
			expected:      map[string]any{"id": 2},
			json:          `{"id": 1, "name": "John"}`,
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "invalid JSON",
			expected:      User{ID: 1},
			json:          `not json`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "Invalid JSON",
		},
		{
			name:          "request error",
			expected:      User{ID: 1},
			json:          "",
			err:           errors.New("connection refused"),
			wantOk:        false,
			wantRetryable: true,
		},
		{
			name:          "empty body",
			expected:      User{ID: 1},
			json:          "",
			wantOk:        false,
			wantRetryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.BodyPartial(tt.expected)
			var resp *client.Response[any]
			if tt.err == nil {
				resp = &client.Response[any]{RawBody: []byte(tt.json)}
			}

			result := exp.Check(tt.err, resp)

			assert.Equal(t, tt.wantOk, result.Ok, "Reason: %s", result.Reason)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectResponseBodyPartial_NestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}
	type User struct {
		ID      int     `json:"id"`
		Address Address `json:"address"`
	}

	tests := []struct {
		name          string
		expected      User
		json          string
		wantOk        bool
		wantRetryable bool
		wantContains  string
	}{
		{
			name:     "nested match",
			expected: User{Address: Address{City: "Moscow"}},
			json:     `{"id": 1, "name": "John", "address": {"city": "Moscow", "street": "Main"}}`,
			wantOk:   true,
		},
		{
			name:          "nested mismatch",
			expected:      User{Address: Address{City: "Moscow"}},
			json:          `{"id": 1, "address": {"city": "London"}}`,
			wantOk:        false,
			wantRetryable: true,
			wantContains:  "address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := jsonSource.BodyPartial(tt.expected)
			resp := &client.Response[any]{RawBody: []byte(tt.json)}

			result := exp.Check(nil, resp)

			assert.Equal(t, tt.wantOk, result.Ok, "Reason: %s", result.Reason)
			if !tt.wantOk {
				assert.Equal(t, tt.wantRetryable, result.Retryable)
				if tt.wantContains != "" {
					assert.Contains(t, result.Reason, tt.wantContains)
				}
			}
		})
	}
}

func TestExpectResponseBodyPartial_FieldMissing(t *testing.T) {
	type User struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
	exp := jsonSource.BodyPartial(User{ID: 1, Email: "test@test.com"})
	resp := &client.Response[any]{
		RawBody: []byte(`{"id": 1, "name": "John"}`),
	}

	result := exp.Check(nil, resp)

	assert.False(t, result.Ok)
	assert.True(t, result.Retryable)
	assert.Contains(t, result.Reason, "email")
}

func TestValidateJSONPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "user.name",
			wantErr: false,
		},
		{
			name:    "valid array path",
			path:    "users.0.name",
			wantErr: false,
		},
		{
			name:    "single field",
			path:    "id",
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := expect.ValidateJSONPath(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot be empty")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
