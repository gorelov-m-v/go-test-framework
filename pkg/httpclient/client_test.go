package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_DoTyped_SuccessJSON(t *testing.T) {
	type reqBody struct {
		Name string `json:"name"`
	}
	type respBody struct {
		OK bool `json:"ok"`
		ID int  `json:"id"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/users/42", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Contains(t, r.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "req-123", r.Header.Get("X-Request-ID"))

		var b reqBody
		require.NoError(t, json.NewDecoder(r.Body).Decode(&b))
		assert.Equal(t, "John", b.Name)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respBody{OK: true, ID: 7})
	}))
	defer srv.Close()

	client := New(Config{
		BaseURL: srv.URL,
		Timeout: 5 * time.Second,
		DefaultHeaders: map[string]string{
			"Accept": "application/json",
		},
	})

	req := &Request[reqBody]{
		Method: http.MethodPost,
		Path:   "/users/{id}",
		PathParams: map[string]string{
			"id": "42",
		},
		QueryParams: map[string]string{
			"page": "2",
		},
		Headers: map[string]string{
			"X-Request-ID": "req-123",
		},
		Body: &reqBody{Name: "John"},
	}

	resp, err := DoTyped[reqBody, respBody](context.Background(), client, req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Empty(t, resp.NetworkError)
	assert.Nil(t, resp.Error)
	assert.NotEmpty(t, resp.RawBody)
	assert.True(t, resp.Duration > 0)
	assert.True(t, resp.Body.OK)
	assert.Equal(t, 7, resp.Body.ID)
}

func TestClient_DoTyped_ErrorEnvelope(t *testing.T) {
	type errEnvelope struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(errEnvelope{
			Message: "validation failed",
			Errors: map[string][]string{
				"field": {"must be positive"},
			},
		})
	}))
	defer srv.Close()

	client := New(Config{BaseURL: srv.URL})

	req := &Request[any]{
		Method:  http.MethodGet,
		Path:    "/fail",
		Headers: map[string]string{},
	}

	resp, err := DoTyped[any, map[string]any](context.Background(), client, req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "validation failed", resp.Error.Message)
	assert.Contains(t, resp.Error.Errors["field"], "must be positive")
	assert.NotEmpty(t, resp.RawBody)
	assert.Empty(t, resp.NetworkError)
}

func TestClient_Config_SecretHeaders(t *testing.T) {
	c1 := New(Config{})
	assert.True(t, c1.IsSecretHeader("Authorization"))
	assert.True(t, c1.IsSecretHeader("authorization"), "Must be case insensitive")
	assert.True(t, c1.IsSecretHeader("X-Api-Key"))
	assert.False(t, c1.IsSecretHeader("Content-Type"))

	c2 := New(Config{
		SecretHeaders: []string{"X-My-Secret", "Authorization"},
	})
	assert.True(t, c2.IsSecretHeader("x-my-secret"))
	assert.True(t, c2.IsSecretHeader("Authorization"))
	assert.False(t, c2.IsSecretHeader("Cookie"))
}

func TestClient_Config_Timeout(t *testing.T) {
	c1 := New(Config{})
	assert.Equal(t, 30*time.Second, c1.HTTPClient.Timeout)

	c2 := New(Config{Timeout: 10 * time.Second})
	assert.Equal(t, 10*time.Second, c2.HTTPClient.Timeout)
}

func TestClient_DoTyped_NetworkError(t *testing.T) {
	client := New(Config{BaseURL: "http://127.0.0.1:1"}) // Порт 1 закрыт

	req := &Request[any]{
		Method: http.MethodGet,
		Path:   "/",
	}

	resp, err := DoTyped[any, any](context.Background(), client, req)

	require.Error(t, err)

	assert.Contains(t, err.Error(), "refused")
	assert.NotNil(t, resp)
	assert.Contains(t, resp.NetworkError, "request failed")
	assert.True(t, resp.Duration >= 0)
}

func TestClient_DoTyped_ContextTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := New(Config{BaseURL: srv.URL})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req := &Request[any]{
		Method: http.MethodGet,
		Path:   "/",
	}

	resp, err := DoTyped[any, any](ctx, client, req)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.NetworkError, "request failed")
}

func TestClient_DoTyped_BuildRequestError(t *testing.T) {
	client := New(Config{BaseURL: "https://example.com"})

	resp, err := DoTyped[any, any](context.Background(), client, nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request is nil")
	assert.NotNil(t, resp)
	assert.Contains(t, resp.NetworkError, "failed to build request")
}

func TestClient_Do_Wrapper(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"foo":"bar"}`))
	}))
	defer srv.Close()

	client := New(Config{BaseURL: srv.URL})

	req := &Request[any]{
		Method: "GET",
		Path:   "/",
	}

	resp, err := client.Do(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_DoTyped_DefaultHeadersOverriddenByRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/xml", r.Header.Get("Accept"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	client := New(Config{
		BaseURL: srv.URL,
		DefaultHeaders: map[string]string{
			"Accept": "application/json",
		},
	})

	req := &Request[any]{
		Method: http.MethodGet,
		Path:   "/",
		Headers: map[string]string{
			"Accept": "application/xml",
		},
	}

	resp, err := DoTyped[any, any](context.Background(), client, req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestClient_DoTyped_NonJSONResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("OK"))
	}))
	defer srv.Close()

	client := New(Config{BaseURL: srv.URL})

	req := &Request[any]{
		Method:  http.MethodGet,
		Path:    "/plain",
		Headers: map[string]string{},
	}

	resp, err := DoTyped[any, map[string]any](context.Background(), client, req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []byte("OK"), resp.RawBody)
	assert.Empty(t, resp.NetworkError)
}
