package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		cfg            Config
		wantTimeout    time.Duration
		wantBaseURL    string
		wantValidator  bool
		wantMaskCount  int
		wantErr        bool
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
