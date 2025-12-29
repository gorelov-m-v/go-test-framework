package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRequest_PathParams(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method: "GET",
		Path:   "/users/{user_id}/posts/{post_id}",
		PathParams: map[string]string{
			"user_id": "123",
			"post_id": "456",
		},
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	assert.Equal(t, "GET", httpReq.Method)
	assert.Equal(t, "https://api.example.com/users/123/posts/456", httpReq.URL.String())

	eff, err := BuildEffectiveURL(client.BaseURL, req.Path, req.PathParams, req.QueryParams)
	require.NoError(t, err)
	assert.Equal(t, httpReq.URL.String(), eff)
}

func TestBuildRequest_PathParams_URLEncoding(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method: "GET",
		Path:   "/search/{query}",
		PathParams: map[string]string{
			"query": "hello world",
		},
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)
	assert.Contains(t, httpReq.URL.String(), "hello%20world")
}

func TestBuildRequest_QueryParams(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method: "GET",
		Path:   "/users",
		QueryParams: map[string]string{
			"page":  "2",
			"limit": "10",
			"sort":  "name",
		},
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	query := httpReq.URL.Query()
	assert.Equal(t, "2", query.Get("page"))
	assert.Equal(t, "10", query.Get("limit"))
	assert.Equal(t, "name", query.Get("sort"))
}

func TestBuildRequest_QueryParams_Encoding(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method: "GET",
		Path:   "/search",
		QueryParams: map[string]string{
			"q":      "hello world",
			"filter": "type=user&status=active",
		},
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)
	assert.Contains(t, httpReq.URL.RawQuery, url.QueryEscape("hello world"))
	assert.Contains(t, httpReq.URL.RawQuery, url.QueryEscape("type=user&status=active"))
}

func TestBuildRequest_JSONBody(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	type RequestBody struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	body := RequestBody{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	req := &Request[RequestBody]{
		Method:  http.MethodPost,
		Path:    "/users",
		Body:    &body,
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, httpReq.Method)
	assert.Equal(t, "application/json", httpReq.Header.Get("Content-Type"))

	bodyBytes, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	assert.Contains(t, string(bodyBytes), `"name":"John Doe"`)
	assert.Contains(t, string(bodyBytes), `"email":"john@example.com"`)
}

func TestBuildRequest_RawBody(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	rawData := []byte("raw content here")

	req := &Request[any]{
		Method:  http.MethodPost,
		Path:    "/data",
		RawBody: rawData,
		Headers: map[string]string{
			"Content-Type": "text/plain",
		},
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	bodyBytes, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)
	assert.Equal(t, rawData, bodyBytes)
	assert.Equal(t, "text/plain", httpReq.Header.Get("Content-Type"))
}

func TestBuildRequest_Headers(t *testing.T) {
	client := New(Config{
		BaseURL: "https://api.example.com",
		DefaultHeaders: map[string]string{
			"User-Agent": "TestClient/1.0",
			"Accept":     "application/json",
		},
	})

	req := &Request[any]{
		Method: http.MethodGet,
		Path:   "/users",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"X-Request-ID":  "req-123",
		},
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)
	assert.Equal(t, "TestClient/1.0", httpReq.Header.Get("User-Agent"))
	assert.Equal(t, "application/json", httpReq.Header.Get("Accept"))
	assert.Equal(t, "Bearer token123", httpReq.Header.Get("Authorization"))
	assert.Equal(t, "req-123", httpReq.Header.Get("X-Request-ID"))
}

func TestBuildRequest_HeadersOverride(t *testing.T) {
	client := New(Config{
		BaseURL: "https://api.example.com",
		DefaultHeaders: map[string]string{
			"Accept": "application/json",
		},
	})

	req := &Request[any]{
		Method: http.MethodGet,
		Path:   "/users",
		Headers: map[string]string{
			"Accept": "application/xml", // override
		},
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)
	assert.Equal(t, "application/xml", httpReq.Header.Get("Accept"))
}

func TestBuildRequest_Multipart(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method: http.MethodPost,
		Path:   "/upload",
		Multipart: &MultipartForm{
			Fields: map[string]string{
				"title":       "Test Upload",
				"description": "Test file",
			},
			Files: []MultipartFile{
				{
					FieldName: "file",
					FileName:  "test.txt",
					Content:   []byte("file content here"),
				},
			},
		},
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	assert.Equal(t, http.MethodPost, httpReq.Method)

	contentType := httpReq.Header.Get("Content-Type")
	assert.True(t, strings.HasPrefix(contentType, "multipart/form-data; boundary="))

	bodyBytes, err := io.ReadAll(httpReq.Body)
	require.NoError(t, err)

	bodyStr := string(bodyBytes)
	assert.Contains(t, bodyStr, "title")
	assert.Contains(t, bodyStr, "Test Upload")
	assert.Contains(t, bodyStr, "file")
	assert.Contains(t, bodyStr, "test.txt")
	assert.Contains(t, bodyStr, "file content here")
}

func TestBuildRequest_EmptyBody(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method:  http.MethodGet,
		Path:    "/users",
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	assert.Equal(t, http.MethodGet, httpReq.Method)
	assert.True(t, httpReq.Body == nil || httpReq.Body == http.NoBody)
	assert.Empty(t, httpReq.Header.Get("Content-Type"))
}

func TestBuildRequest_InvalidBaseURL_ParseError(t *testing.T) {
	client := New(Config{BaseURL: "://invalid-url"})

	req := &Request[any]{
		Method:  http.MethodGet,
		Path:    "/users",
		Headers: make(map[string]string),
	}

	_, err := buildRequest(context.Background(), client, req)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "invalid base URL")
}

func TestBuildRequest_InvalidBaseURL_MissingSchemeOrHost(t *testing.T) {
	client := New(Config{BaseURL: "api.example.com"})

	req := &Request[any]{
		Method:  http.MethodGet,
		Path:    "/users",
		Headers: make(map[string]string),
	}

	_, err := buildRequest(context.Background(), client, req)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "base URL must include scheme and host")
}

func TestBuildRequest_Context(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	req := &Request[any]{
		Method:  http.MethodGet,
		Path:    "/users",
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(ctx, client, req)
	require.NoError(t, err)

	assert.Equal(t, ctx, httpReq.Context())
}

func TestBuildRequest_BaseURLWithPathPrefix(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com/api"})

	req := &Request[any]{
		Method:  http.MethodGet,
		Path:    "/v1/users",
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	assert.Equal(t, "https://api.example.com/api/v1/users", httpReq.URL.String())
}

func TestBuildEffectiveURL_MatchesBuildRequest(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com/api"})

	req := &Request[any]{
		Method: "GET",
		Path:   "/v1/users/{id}",
		PathParams: map[string]string{
			"id": "42",
		},
		QueryParams: map[string]string{
			"q": "hello world",
		},
		Headers: make(map[string]string),
	}

	httpReq, err := buildRequest(context.Background(), client, req)
	require.NoError(t, err)

	eff, err := BuildEffectiveURL(client.BaseURL, req.Path, req.PathParams, req.QueryParams)
	require.NoError(t, err)

	assert.Equal(t, httpReq.URL.String(), eff)
}

func TestBuildRequest_ValidationErrors(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})
	ctx := context.Background()

	tests := []struct {
		name        string
		client      *Client
		req         *Request[any]
		expectedErr string
	}{
		{
			name:        "Nil httpclient",
			client:      nil,
			req:         &Request[any]{Method: "GET", Path: "/"},
			expectedErr: "httpclient is nil",
		},
		{
			name:        "Nil request",
			client:      client,
			req:         nil,
			expectedErr: "request is nil",
		},
		{
			name:        "Empty Base URL",
			client:      New(Config{BaseURL: "   "}),
			req:         &Request[any]{Method: "GET", Path: "/"},
			expectedErr: "base URL is empty",
		},
		{
			name:        "Empty Method",
			client:      client,
			req:         &Request[any]{Method: "", Path: "/"},
			expectedErr: "request method is empty",
		},
		{
			name:        "Empty Path",
			client:      client,
			req:         &Request[any]{Method: "GET", Path: "   "},
			expectedErr: "request path is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildRequest(ctx, tt.client, tt.req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestBuildRequest_AbsolutePathError(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method: "GET",
		Path:   "https://evil.com/hack",
	}

	_, err := buildRequest(context.Background(), client, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request path must be relative")
}

func TestBuildRequest_JSONMarshalError(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	type InvalidBody struct {
		Ch chan int `json:"ch"`
	}

	body := InvalidBody{Ch: make(chan int)}

	req := &Request[InvalidBody]{
		Method: "POST",
		Path:   "/users",
		Body:   &body,
	}

	_, err := buildRequest(context.Background(), client, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal JSON body")
}

func TestBuildRequest_InvalidHTTPMethod(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})

	req := &Request[any]{
		Method: "GET / HTTP/1.1", // Невалидный токен метода
		Path:   "/users",
	}

	_, err := buildRequest(context.Background(), client, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create HTTP request")
}

func TestBuildEffectiveURL_ErrorPropagation(t *testing.T) {
	_, err := BuildEffectiveURL("://bad-url", "/path", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid base URL")
}
