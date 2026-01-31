package client

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		cfg           Config
		wantTimeout   time.Duration
		wantBaseURL   string
		wantValidator bool
		wantMaskCount int
		wantErr       bool
	}{
		{
			name: "default timeout",
			cfg: Config{
				BaseURL: "https://api.example.com",
			},
			wantTimeout: 30 * time.Second,
			wantBaseURL: "https://api.example.com",
		},
		{
			name: "custom timeout",
			cfg: Config{
				BaseURL: "https://api.example.com",
				Timeout: 10 * time.Second,
			},
			wantTimeout: 10 * time.Second,
			wantBaseURL: "https://api.example.com",
		},
		{
			name: "with mask headers",
			cfg: Config{
				BaseURL:     "https://api.example.com",
				MaskHeaders: "Authorization, Cookie, X-Api-Key",
			},
			wantTimeout:   30 * time.Second,
			wantMaskCount: 3,
		},
		{
			name: "with default headers",
			cfg: Config{
				BaseURL: "https://api.example.com",
				DefaultHeaders: map[string]string{
					"Accept": "application/json",
				},
			},
			wantTimeout: 30 * time.Second,
		},
		{
			name: "invalid contract spec",
			cfg: Config{
				BaseURL:      "https://api.example.com",
				ContractSpec: "nonexistent.yaml",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)
			assert.Equal(t, tt.wantTimeout, client.HTTPClient.Timeout)
			if tt.wantBaseURL != "" {
				assert.Equal(t, tt.wantBaseURL, client.BaseURL)
			}
			if tt.wantMaskCount > 0 {
				assert.Len(t, client.maskHeaders, tt.wantMaskCount)
			}
		})
	}
}

func TestShouldMaskHeader(t *testing.T) {
	tests := []struct {
		name        string
		maskHeaders string
		headerName  string
		want        bool
	}{
		{
			name:        "mask authorization",
			maskHeaders: "Authorization,Cookie",
			headerName:  "Authorization",
			want:        true,
		},
		{
			name:        "mask case insensitive",
			maskHeaders: "Authorization,Cookie",
			headerName:  "authorization",
			want:        true,
		},
		{
			name:        "mask with spaces",
			maskHeaders: "Authorization, Cookie",
			headerName:  "Cookie",
			want:        true,
		},
		{
			name:        "not masked",
			maskHeaders: "Authorization,Cookie",
			headerName:  "Content-Type",
			want:        false,
		},
		{
			name:        "no mask configured",
			maskHeaders: "",
			headerName:  "Authorization",
			want:        false,
		},
		{
			name:        "header with spaces",
			maskHeaders: "Authorization",
			headerName:  "  Authorization  ",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(Config{
				BaseURL:     "https://api.example.com",
				MaskHeaders: tt.maskHeaders,
			})
			require.NoError(t, err)

			result := client.ShouldMaskHeader(tt.headerName)

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestBuildEffectiveURL(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		path        string
		pathParams  map[string]string
		queryParams map[string]string
		want        string
		wantErr     bool
	}{
		{
			name: "simple path",
			base: "https://api.example.com",
			path: "/users",
			want: "https://api.example.com/users",
		},
		{
			name: "path without leading slash",
			base: "https://api.example.com",
			path: "users",
			want: "https://api.example.com/users",
		},
		{
			name: "with path params",
			base: "https://api.example.com",
			path: "/users/{id}",
			pathParams: map[string]string{
				"id": "123",
			},
			want: "https://api.example.com/users/123",
		},
		{
			name: "with multiple path params",
			base: "https://api.example.com",
			path: "/users/{userId}/posts/{postId}",
			pathParams: map[string]string{
				"userId": "1",
				"postId": "42",
			},
			want: "https://api.example.com/users/1/posts/42",
		},
		{
			name: "with query params",
			base: "https://api.example.com",
			path: "/users",
			queryParams: map[string]string{
				"page":  "1",
				"limit": "10",
			},
			want: "https://api.example.com/users?limit=10&page=1",
		},
		{
			name: "with base path",
			base: "https://api.example.com/v1",
			path: "/users",
			want: "https://api.example.com/v1/users",
		},
		{
			name: "path param with special chars",
			base: "https://api.example.com",
			path: "/search/{query}",
			pathParams: map[string]string{
				"query": "hello world",
			},
			want: "https://api.example.com/search/hello%20world",
		},
		{
			name:    "invalid base URL",
			base:    "not-a-url",
			path:    "/users",
			wantErr: true,
		},
		{
			name:    "empty base URL",
			base:    "",
			path:    "/users",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildEffectiveURL(tt.base, tt.path, tt.pathParams, tt.queryParams)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestIsJSONContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{
			name:        "application/json",
			contentType: "application/json",
			want:        true,
		},
		{
			name:        "application/json with charset",
			contentType: "application/json; charset=utf-8",
			want:        true,
		},
		{
			name:        "text/json",
			contentType: "text/json",
			want:        true,
		},
		{
			name:        "application/vnd.api+json",
			contentType: "application/vnd.api+json",
			want:        true,
		},
		{
			name:        "text/html",
			contentType: "text/html",
			want:        false,
		},
		{
			name:        "text/plain",
			contentType: "text/plain",
			want:        false,
		},
		{
			name:        "empty",
			contentType: "",
			want:        false,
		},
		{
			name:        "uppercase JSON",
			contentType: "application/JSON",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJSONContentType(tt.contentType)

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		body       []byte
		statusCode int
		wantMsg    string
		wantErrors map[string][]string
	}{
		{
			name:       "empty body",
			body:       []byte{},
			statusCode: 400,
			wantMsg:    "",
		},
		{
			name:       "json with message",
			body:       []byte(`{"message": "Bad request"}`),
			statusCode: 400,
			wantMsg:    "Bad request",
		},
		{
			name:       "json with errors",
			body:       []byte(`{"message": "Validation failed", "errors": {"email": ["invalid format"]}}`),
			statusCode: 422,
			wantMsg:    "Validation failed",
			wantErrors: map[string][]string{"email": {"invalid format"}},
		},
		{
			name:       "invalid json",
			body:       []byte(`not json`),
			statusCode: 500,
			wantMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseErrorResponse(tt.body, tt.statusCode)

			require.NotNil(t, result)
			assert.Equal(t, tt.statusCode, result.StatusCode)
			assert.Equal(t, tt.wantMsg, result.Message)
			if tt.wantErrors != nil {
				assert.Equal(t, tt.wantErrors, result.Errors)
			}
		})
	}
}

func TestValidateBuildInput(t *testing.T) {
	validClient := &Client{BaseURL: "https://api.example.com"}
	validRequest := &Request[any]{Method: "GET", Path: "/users"}

	tests := []struct {
		name    string
		client  *Client
		req     *Request[any]
		wantErr string
	}{
		{
			name:    "valid input",
			client:  validClient,
			req:     validRequest,
			wantErr: "",
		},
		{
			name:    "nil client",
			client:  nil,
			req:     validRequest,
			wantErr: "httpclient is nil",
		},
		{
			name:    "nil request",
			client:  validClient,
			req:     nil,
			wantErr: "request is nil",
		},
		{
			name:    "empty base URL",
			client:  &Client{BaseURL: ""},
			req:     validRequest,
			wantErr: "base URL is empty",
		},
		{
			name:    "whitespace base URL",
			client:  &Client{BaseURL: "   "},
			req:     validRequest,
			wantErr: "base URL is empty",
		},
		{
			name:    "empty method",
			client:  validClient,
			req:     &Request[any]{Method: "", Path: "/users"},
			wantErr: "request method is empty",
		},
		{
			name:    "whitespace method",
			client:  validClient,
			req:     &Request[any]{Method: "  ", Path: "/users"},
			wantErr: "request method is empty",
		},
		{
			name:    "empty path",
			client:  validClient,
			req:     &Request[any]{Method: "GET", Path: ""},
			wantErr: "request path is empty",
		},
		{
			name:    "whitespace path",
			client:  validClient,
			req:     &Request[any]{Method: "GET", Path: "   "},
			wantErr: "request path is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBuildInput(tt.client, tt.req)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestBuildBody(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name            string
		req             *Request[testStruct]
		wantContentType string
		wantBody        string
		wantErr         string
	}{
		{
			name:            "no body",
			req:             &Request[testStruct]{},
			wantContentType: "",
			wantBody:        "",
		},
		{
			name: "json body from struct",
			req: &Request[testStruct]{
				Body: &testStruct{Name: "test", Value: 42},
			},
			wantContentType: "application/json",
			wantBody:        `{"name":"test","value":42}`,
		},
		{
			name: "json body from map",
			req: &Request[testStruct]{
				BodyMap: map[string]interface{}{"key": "value", "num": 123},
			},
			wantContentType: "application/json",
			wantBody:        `{"key":"value","num":123}`,
		},
		{
			name: "raw body",
			req: &Request[testStruct]{
				RawBody: []byte("raw content"),
			},
			wantContentType: "",
			wantBody:        "raw content",
		},
		{
			name: "multipart with fields only",
			req: &Request[testStruct]{
				Multipart: &MultipartForm{
					Fields: map[string]string{"field1": "value1"},
				},
			},
			wantContentType: "multipart/form-data",
		},
		{
			name: "multipart with files",
			req: &Request[testStruct]{
				Multipart: &MultipartForm{
					Fields: map[string]string{"name": "test"},
					Files: []MultipartFile{
						{FieldName: "file", FileName: "test.txt", Content: []byte("file content")},
					},
				},
			},
			wantContentType: "multipart/form-data",
		},
		{
			name: "multiple body types error - body and bodymap",
			req: &Request[testStruct]{
				Body:    &testStruct{Name: "test"},
				BodyMap: map[string]interface{}{"key": "value"},
			},
			wantErr: "only one body type can be set",
		},
		{
			name: "multiple body types error - body and raw",
			req: &Request[testStruct]{
				Body:    &testStruct{Name: "test"},
				RawBody: []byte("raw"),
			},
			wantErr: "only one body type can be set",
		},
		{
			name: "multiple body types error - body and multipart",
			req: &Request[testStruct]{
				Body:      &testStruct{Name: "test"},
				Multipart: &MultipartForm{Fields: map[string]string{"f": "v"}},
			},
			wantErr: "only one body type can be set",
		},
		{
			name: "multiple body types error - all four",
			req: &Request[testStruct]{
				Body:      &testStruct{Name: "test"},
				BodyMap:   map[string]interface{}{"key": "value"},
				RawBody:   []byte("raw"),
				Multipart: &MultipartForm{Fields: map[string]string{"f": "v"}},
			},
			wantErr: "only one body type can be set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, contentType, err := buildBody(tt.req)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			require.NoError(t, err)

			if tt.wantContentType == "multipart/form-data" {
				assert.Contains(t, contentType, "multipart/form-data")
			} else {
				assert.Equal(t, tt.wantContentType, contentType)
			}

			if tt.wantBody != "" {
				body, err := io.ReadAll(reader)
				require.NoError(t, err)
				assert.Equal(t, tt.wantBody, string(body))
			} else if reader == nil {
				assert.Empty(t, tt.wantBody)
			}
		})
	}
}

func TestCountTrue(t *testing.T) {
	tests := []struct {
		name  string
		flags []bool
		want  int
	}{
		{
			name:  "all false",
			flags: []bool{false, false, false},
			want:  0,
		},
		{
			name:  "one true",
			flags: []bool{false, true, false},
			want:  1,
		},
		{
			name:  "all true",
			flags: []bool{true, true, true},
			want:  3,
		},
		{
			name:  "empty",
			flags: []bool{},
			want:  0,
		},
		{
			name:  "mixed",
			flags: []bool{true, false, true, false, true},
			want:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countTrue(tt.flags...)

			assert.Equal(t, tt.want, result)
		})
	}
}
