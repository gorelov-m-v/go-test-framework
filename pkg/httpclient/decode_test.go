package httpclient

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserResp struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestDecodeResponse_Success_JSON(t *testing.T) {
	jsonBody := `{"id": 1, "name": "Alice"}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(jsonBody)),
	}

	result, err := decodeResponse[UserResp](resp, 100*time.Millisecond)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Equal(t, 1, result.Body.ID)
	assert.Equal(t, "Alice", result.Body.Name)
	assert.Equal(t, jsonBody, string(result.RawBody))
	assert.Empty(t, result.NetworkError)
	assert.Nil(t, result.Error)
}

func TestDecodeResponse_Success_NonJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(bytes.NewBufferString("some plain text")),
	}

	result, err := decodeResponse[UserResp](resp, 100*time.Millisecond)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.Empty(t, result.Body.Name) // Поля пустые (Zero Value)
	assert.Equal(t, "some plain text", string(result.RawBody))
	assert.Empty(t, result.NetworkError)
}

func TestDecodeResponse_MalformedJSON_SetsNetworkError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(`{"id": 1, "name": "inc`)), // Обрезанный JSON
	}

	result, err := decodeResponse[UserResp](resp, 100*time.Millisecond)

	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, result.StatusCode)
	assert.NotEmpty(t, result.NetworkError)
	assert.Contains(t, result.NetworkError, "failed to decode response body")
}

func TestDecodeResponse_ErrorStatus_ParsedEnvelope(t *testing.T) {
	errorJSON := `{
		"message": "Validation failed",
		"errors": {
			"email": ["invalid format"]
		}
	}`
	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(errorJSON)),
	}

	result, err := decodeResponse[UserResp](resp, 100*time.Millisecond)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, result.StatusCode)
	assert.NotNil(t, result.Error)
	assert.Equal(t, "Validation failed", result.Error.Message)
	assert.Equal(t, []string{"invalid format"}, result.Error.Errors["email"])
	assert.Equal(t, 0, result.Body.ID)
}

func TestDecodeResponse_ErrorStatus_RawBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Header:     http.Header{"Content-Type": []string{"text/html"}},
		Body:       io.NopCloser(bytes.NewBufferString("<html>Internal Server Error</html>")),
	}

	result, err := decodeResponse[UserResp](resp, 100*time.Millisecond)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	assert.NotNil(t, result.Error)
	assert.Empty(t, result.Error.Message)
	assert.Equal(t, "<html>Internal Server Error</html>", result.Error.Body)
}

func TestDecodeResponse_EmptyBody(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusNoContent, // 204
		Header:     http.Header{},
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	result, err := decodeResponse[UserResp](resp, 100*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, result.StatusCode)
	assert.Empty(t, result.RawBody)
}

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}
func (e *errReader) Close() error { return nil }

func TestDecodeResponse_ReadError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       &errReader{},
	}

	result, err := decodeResponse[UserResp](resp, 100*time.Millisecond)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated read error")
	assert.NotNil(t, result)
	assert.Contains(t, result.NetworkError, "failed to read response body")
}

func TestIsJSONContentType(t *testing.T) {
	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json", true},
		{"APPLICATION/JSON", true},
		{"application/json; charset=utf-8", true},
		{"application/vnd.api+json", true},
		{"text/plain", false},
		{"text/html", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			assert.Equal(t, tt.expected, isJSONContentType(tt.contentType))
		})
	}
}
