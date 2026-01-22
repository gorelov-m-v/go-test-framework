package contract

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestSpec() *openapi3.T {
	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: &openapi3.Paths{},
		Components: &openapi3.Components{
			Schemas: openapi3.Schemas{},
		},
	}

	userSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"id": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
				},
				"name": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
				},
				"email": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
				},
			},
			Required: []string{"id", "name"},
		},
	}
	spec.Components.Schemas["User"] = userSchema

	errorSchema := &openapi3.SchemaRef{
		Value: &openapi3.Schema{
			Type: &openapi3.Types{"object"},
			Properties: openapi3.Schemas{
				"code": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
				},
				"message": &openapi3.SchemaRef{
					Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
				},
			},
			Required: []string{"code", "message"},
		},
	}
	spec.Components.Schemas["Error"] = errorSchema

	spec.Paths.Set("/users", &openapi3.PathItem{
		Get: &openapi3.Operation{
			Responses: &openapi3.Responses{},
		},
		Post: &openapi3.Operation{
			Responses: &openapi3.Responses{},
		},
	})

	spec.Paths.Find("/users").Get.Responses.Set("200", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: ptr("List of users"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:  &openapi3.Types{"array"},
							Items: userSchema,
						},
					},
				},
			},
		},
	})

	spec.Paths.Find("/users").Post.Responses.Set("201", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: ptr("Created user"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: userSchema,
				},
			},
		},
	})

	spec.Paths.Find("/users").Post.Responses.Set("400", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: ptr("Bad request"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: errorSchema,
				},
			},
		},
	})

	spec.Paths.Set("/users/{id}", &openapi3.PathItem{
		Get: &openapi3.Operation{
			Responses: &openapi3.Responses{},
		},
	})

	spec.Paths.Find("/users/{id}").Get.Responses.Set("200", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: ptr("User details"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: userSchema,
				},
			},
		},
	})

	spec.Paths.Find("/users/{id}").Get.Responses.Set("404", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: ptr("Not found"),
			Content: openapi3.Content{
				"application/json": &openapi3.MediaType{
					Schema: errorSchema,
				},
			},
		},
	})

	return spec
}

func ptr(s string) *string {
	return &s
}

func TestValidateResponse(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		statusCode   int
		body         []byte
		wantErr      bool
		wantErrType  ErrorType
		wantContains string
	}{
		{
			name:       "valid response",
			method:     "POST",
			path:       "/users",
			statusCode: 201,
			body:       []byte(`{"id": 1, "name": "John", "email": "john@test.com"}`),
			wantErr:    false,
		},
		{
			name:         "missing required field",
			method:       "POST",
			path:         "/users",
			statusCode:   201,
			body:         []byte(`{"id": 1}`),
			wantErr:      true,
			wantErrType:  ErrSchemaValidation,
			wantContains: "name",
		},
		{
			name:        "wrong type",
			method:      "POST",
			path:        "/users",
			statusCode:  201,
			body:        []byte(`{"id": "not-a-number", "name": "John"}`),
			wantErr:     true,
			wantErrType: ErrSchemaValidation,
		},
		{
			name:        "path not found",
			method:      "GET",
			path:        "/unknown",
			statusCode:  200,
			body:        []byte(`{}`),
			wantErr:     true,
			wantErrType: ErrPathNotFound,
		},
		{
			name:        "method not found",
			method:      "DELETE",
			path:        "/users",
			statusCode:  200,
			body:        []byte(`{}`),
			wantErr:     true,
			wantErrType: ErrOperationNotFound,
		},
		{
			name:        "response not defined",
			method:      "POST",
			path:        "/users",
			statusCode:  500,
			body:        []byte(`{}`),
			wantErr:     true,
			wantErrType: ErrResponseNotDefined,
		},
		{
			name:       "path with param",
			method:     "GET",
			path:       "/users/123",
			statusCode: 200,
			body:       []byte(`{"id": 1, "name": "John"}`),
			wantErr:    false,
		},
		{
			name:        "invalid JSON",
			method:      "POST",
			path:        "/users",
			statusCode:  201,
			body:        []byte(`not json`),
			wantErr:     true,
			wantErrType: ErrInvalidJSON,
		},
		{
			name:       "empty body",
			method:     "POST",
			path:       "/users",
			statusCode: 201,
			body:       []byte{},
			wantErr:    false,
		},
		{
			name:       "array response valid",
			method:     "GET",
			path:       "/users",
			statusCode: 200,
			body:       []byte(`[{"id": 1, "name": "John"}, {"id": 2, "name": "Jane"}]`),
			wantErr:    false,
		},
		{
			name:        "array item invalid",
			method:      "GET",
			path:        "/users",
			statusCode:  200,
			body:        []byte(`[{"id": 1, "name": "John"}, {"id": "invalid"}]`),
			wantErr:     true,
			wantErrType: ErrSchemaValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidatorFromSpec(createTestSpec())

			err := v.ValidateResponse(tt.method, tt.path, tt.statusCode, tt.body)

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.wantErrType, validationErr.Type)
				if tt.wantContains != "" {
					assert.Contains(t, validationErr.Message, tt.wantContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateResponseBySchema(t *testing.T) {
	tests := []struct {
		name        string
		schema      string
		body        []byte
		wantErr     bool
		wantErrType ErrorType
	}{
		{
			name:    "valid",
			schema:  "User",
			body:    []byte(`{"id": 1, "name": "John"}`),
			wantErr: false,
		},
		{
			name:        "missing required",
			schema:      "User",
			body:        []byte(`{"id": 1}`),
			wantErr:     true,
			wantErrType: ErrSchemaValidation,
		},
		{
			name:        "schema not found",
			schema:      "Unknown",
			body:        []byte(`{}`),
			wantErr:     true,
			wantErrType: ErrSchemaNotFound,
		},
		{
			name:        "invalid JSON",
			schema:      "User",
			body:        []byte(`not json`),
			wantErr:     true,
			wantErrType: ErrInvalidJSON,
		},
		{
			name:    "empty body",
			schema:  "User",
			body:    []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidatorFromSpec(createTestSpec())

			err := v.ValidateResponseBySchema(tt.schema, tt.body)

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.wantErrType, validationErr.Type)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPathMatches(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		{
			name:     "exact path match",
			pattern:  "/users",
			path:     "/users",
			expected: true,
		},
		{
			name:     "exact path no match",
			pattern:  "/users",
			path:     "/accounts",
			expected: false,
		},
		{
			name:     "single param match",
			pattern:  "/users/{id}",
			path:     "/users/123",
			expected: true,
		},
		{
			name:     "multiple params match",
			pattern:  "/users/{id}/posts/{postId}",
			path:     "/users/1/posts/42",
			expected: true,
		},
		{
			name:     "extra segment no match",
			pattern:  "/users/{id}",
			path:     "/users/123/extra",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pathMatches(tt.pattern, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "without leading slash",
			input:    "users",
			expected: "/users",
		},
		{
			name:     "with leading slash",
			input:    "/users",
			expected: "/users",
		},
		{
			name:     "with query string",
			input:    "/users?page=1",
			expected: "/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
