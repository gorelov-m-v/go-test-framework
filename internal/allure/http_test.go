package allure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	httpClient "github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

func TestToHTTPRequestDTO(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		dto := ToHTTPRequestDTO[any](nil)

		assert.Equal(t, HTTPRequestDTO{}, dto)
	})

	t.Run("request with typed body", func(t *testing.T) {
		body := struct{ Name string }{Name: "test"}
		req := &httpClient.Request[struct{ Name string }]{
			Method:      "POST",
			Path:        "/api/users",
			PathParams:  map[string]string{"id": "123"},
			QueryParams: map[string]string{"page": "1"},
			Headers:     map[string]string{"Content-Type": "application/json"},
			Body:        &body,
		}

		dto := ToHTTPRequestDTO(req)

		assert.Equal(t, "POST", dto.Method)
		assert.Equal(t, "/api/users", dto.Path)
		assert.Equal(t, map[string]string{"id": "123"}, dto.PathParams)
		assert.Equal(t, map[string]string{"page": "1"}, dto.QueryParams)
		assert.Equal(t, map[string]string{"Content-Type": "application/json"}, dto.Headers)
		assert.Equal(t, body, dto.Body)
	})

	t.Run("request with body map", func(t *testing.T) {
		bodyMap := map[string]any{"key": "value"}
		req := &httpClient.Request[any]{
			Method:  "POST",
			Path:    "/api/data",
			BodyMap: bodyMap,
		}

		dto := ToHTTPRequestDTO(req)

		assert.Equal(t, bodyMap, dto.Body)
	})

	t.Run("request with raw body", func(t *testing.T) {
		req := &httpClient.Request[any]{
			Method:  "POST",
			Path:    "/api/raw",
			RawBody: []byte(`{"raw": true}`),
		}

		dto := ToHTTPRequestDTO(req)

		assert.Equal(t, []byte(`{"raw": true}`), dto.RawBody)
	})
}

func TestToHTTPResponseDTO(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		dto := ToHTTPResponseDTO[any](nil)

		assert.Equal(t, HTTPResponseDTO{}, dto)
	})

	t.Run("success response", func(t *testing.T) {
		type respBody struct{ ID string }
		resp := &httpClient.Response[respBody]{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       respBody{ID: "123"},
			RawBody:    []byte(`{"id":"123"}`),
			Duration:   100 * time.Millisecond,
		}

		dto := ToHTTPResponseDTO(resp)

		assert.Equal(t, 200, dto.StatusCode)
		assert.Equal(t, map[string][]string{"Content-Type": {"application/json"}}, dto.Headers)
		assert.NotNil(t, dto.Body)
		assert.Equal(t, []byte(`{"id":"123"}`), dto.RawBody)
		assert.Equal(t, 100*time.Millisecond, dto.Duration)
	})

	t.Run("error response", func(t *testing.T) {
		errResp := &httpClient.ErrorResponse{
			StatusCode: 400,
			Message:    "bad request",
		}
		resp := &httpClient.Response[any]{
			StatusCode:   400,
			Error:        errResp,
			NetworkError: "connection timeout",
		}

		dto := ToHTTPResponseDTO(resp)

		assert.Equal(t, 400, dto.StatusCode)
		assert.Equal(t, errResp, dto.Error)
		assert.Equal(t, "connection timeout", dto.NetworkError)
	})
}
