package dsl

import (
	"testing"
	"time"

	"go-test-framework/pkg/http/client"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/stretchr/testify/assert"
)

type MockStepCtxAttachment struct {
	provider.StepCtx
	Attachments []AttachmentData
}

type AttachmentData struct {
	Name    string
	Mime    allure.MimeType
	Content []byte
}

func (m *MockStepCtxAttachment) WithNewAttachment(name string, mime allure.MimeType, content []byte) {
	m.Attachments = append(m.Attachments, AttachmentData{
		Name:    name,
		Mime:    mime,
		Content: content,
	})
}

func TestAttachRequest_FullJSON(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	httpClient := client.New(client.Config{BaseURL: "https://api.example.com"})

	body := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	req := &client.Request[map[string]interface{}]{
		Method: "POST",
		Path:   "/users/{id}",
		PathParams: map[string]string{
			"id": "123",
		},
		QueryParams: map[string]string{
			"force": "true",
		},
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer secret_token",
		},
		Body: &body,
	}

	attachRequest(mockCtx, httpClient, req)

	assert.Len(t, mockCtx.Attachments, 1)
	att := mockCtx.Attachments[0]
	assert.Equal(t, "HTTP Request", att.Name)
	assert.Equal(t, allure.Text, att.Mime)

	content := string(att.Content)

	assert.Contains(t, content, "Method: POST")
	assert.Contains(t, content, "Path: /users/{id}")
	assert.Contains(t, content, "Effective URL: https://api.example.com/users/123")

	assert.Contains(t, content, "Path Params:")
	assert.Contains(t, content, "  id: 123")
	assert.Contains(t, content, "Query Params:")
	assert.Contains(t, content, "  force: true")

	assert.Contains(t, content, "Headers:")
	assert.Contains(t, content, "  Content-Type: application/json")
	assert.Contains(t, content, "  Authorization: Bearer ***MASKED***")

	assert.Contains(t, content, "Body (json):")
	assert.Contains(t, content, `"name": "John"`)
	assert.Contains(t, content, `"age": 30`)
}

func TestAttachRequest_RawBody_Truncated(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	httpClient := client.New(client.Config{BaseURL: "https://api.example.com"})

	largeBody := make([]byte, 1050)
	for i := range largeBody {
		largeBody[i] = 'X'
	}

	req := &client.Request[any]{
		Method:  "POST",
		Path:    "/upload",
		RawBody: largeBody,
	}

	attachRequest(mockCtx, httpClient, req)

	content := string(mockCtx.Attachments[0].Content)

	assert.Contains(t, content, "Body (raw):")
	assert.Contains(t, content, "(raw body, 1050 bytes)")
	assert.Contains(t, content, "XXXX")
	assert.Contains(t, content, "...")
}

func TestAttachRequest_Multipart(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	httpClient := client.New(client.Config{})

	req := &client.Request[any]{
		Method: "POST",
		Path:   "/files",
		Multipart: &client.MultipartForm{
			Fields: map[string]string{"type": "avatar"},
			Files: []client.MultipartFile{
				{FieldName: "file", FileName: "image.png", Content: []byte{1, 2, 3}},
			},
		},
	}

	attachRequest(mockCtx, httpClient, req)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Body (multipart/form-data):")
	assert.Contains(t, content, "  type: avatar")
	assert.Contains(t, content, "  file: image.png (3 bytes)")
}

func TestAttachRequest_EmptyBody(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	httpClient := client.New(client.Config{})

	req := &client.Request[any]{Method: "GET", Path: "/"}

	attachRequest(mockCtx, httpClient, req)
	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Body: (empty)")
}

func TestAttachRequest_InvalidBaseURL(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	httpClient := client.New(client.Config{BaseURL: "://invalid"})
	req := &client.Request[any]{Method: "GET", Path: "/"}

	attachRequest(mockCtx, httpClient, req)
	content := string(mockCtx.Attachments[0].Content)

	assert.Contains(t, content, "Effective URL: (failed to resolve:")
}

func TestAttachResponse_Success_JSON(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	rawJSON := []byte(`{"status":"ok", "id": 1}`)
	resp := &client.Response[any]{
		StatusCode: 200,
		Duration:   150 * time.Millisecond,
		Headers: map[string][]string{
			"Server": {"Nginx"},
		},
		RawBody: rawJSON,
	}

	attachResponse(mockCtx, nil, resp)

	assert.Len(t, mockCtx.Attachments, 1)
	att := mockCtx.Attachments[0]
	assert.Equal(t, "HTTP Response", att.Name)

	content := string(att.Content)

	assert.Contains(t, content, "Status: 200")
	assert.Contains(t, content, "Duration: 150ms")
	assert.Contains(t, content, "Headers:")
	assert.Contains(t, content, "  Server: Nginx")
	assert.Contains(t, content, "Body:")
	assert.Contains(t, content, `"status": "ok"`)
	assert.Contains(t, content, `"id": 1`)
}

func TestAttachResponse_NetworkError(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	resp := &client.Response[any]{
		NetworkError: "connection refused",
		Duration:     10 * time.Millisecond,
	}

	attachResponse(mockCtx, nil, resp)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Network Error: connection refused")
	assert.Contains(t, content, "Duration: 10ms")
	assert.NotContains(t, content, "Status:") // Статуса нет при ошибке сети
}

func TestAttachResponse_APIErrorStruct(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	resp := &client.Response[any]{
		StatusCode: 400,
		Error: &client.ErrorResponse{
			Message: "Validation Failed",
			Errors: map[string][]string{
				"email": {"invalid format"},
			},
		},
		RawBody: []byte(`{"message":"..."}`),
	}

	attachResponse(mockCtx, nil, resp)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Status: 400")
	assert.Contains(t, content, "Error:")
	assert.Contains(t, content, "  Message: Validation Failed")
	assert.Contains(t, content, "  Errors:")
	assert.Contains(t, content, "    email: invalid format")
}

func TestAttachResponse_APIErrorRaw(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	resp := &client.Response[any]{
		StatusCode: 500,
		Error: &client.ErrorResponse{
			Body: "Internal Server Error HTML",
		},
		RawBody: []byte("Internal Server Error HTML"),
	}

	attachResponse(mockCtx, nil, resp)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Status: 500")
	assert.Contains(t, content, "Error:")
	assert.Contains(t, content, "  Body: Internal Server Error HTML")
}

func TestAttachResponse_RawBody_NonJSON(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	resp := &client.Response[any]{
		StatusCode: 200,
		RawBody:    []byte("Just some plain text response"),
	}

	attachResponse(mockCtx, nil, resp)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Body:")
	assert.Contains(t, content, "Just some plain text response")
}

func TestAttachResponse_BrokenJSON(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	raw := []byte(`{broken: json`)
	resp := &client.Response[any]{
		StatusCode: 200,
		RawBody:    raw,
	}

	attachResponse(mockCtx, nil, resp)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "{broken: json")
}

func TestAttachResponse_NilResponse(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	attachResponse[any](mockCtx, nil, nil)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Response: <nil>")
}

func TestSanitizeHeaders(t *testing.T) {
	httpClient := client.New(client.Config{
		BaseURL: "https://example.com",
	})

	t.Run("masks Authorization Bearer token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer secret_token_123",
			"Content-Type":  "application/json",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "Bearer ***MASKED***", sanitized["Authorization"])
		assert.Equal(t, "application/json", sanitized["Content-Type"])
	})

	t.Run("masks Authorization Basic token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Basic dXNlcjpwYXNz",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "Basic ***MASKED***", sanitized["Authorization"])
	})

	t.Run("masks plain Authorization token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "secret_token",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "***MASKED***", sanitized["Authorization"])
	})

	t.Run("masks Cookie header", func(t *testing.T) {
		headers := map[string]string{
			"Cookie": "session=abc123; user=john",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "***MASKED***", sanitized["Cookie"])
	})

	t.Run("masks Api-Key header", func(t *testing.T) {
		headers := map[string]string{
			"Api-Key": "my_secret_key",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "***MASKED***", sanitized["Api-Key"])
	})

	t.Run("preserves non-secret headers", func(t *testing.T) {
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Accept":        "application/json",
			"User-Agent":    "Test Client",
			"X-Request-ID":  "12345",
			"Cache-Control": "no-cache",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, headers, sanitized)
	})

	t.Run("handles empty headers", func(t *testing.T) {
		headers := map[string]string{}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Empty(t, sanitized)
	})

	t.Run("masks multiple secret headers", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token123",
			"Cookie":        "session=xyz",
			"X-Api-Key":     "key456",
			"Content-Type":  "application/json",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "Bearer ***MASKED***", sanitized["Authorization"])
		assert.Equal(t, "***MASKED***", sanitized["Cookie"])
		assert.Equal(t, "***MASKED***", sanitized["X-Api-Key"])
		assert.Equal(t, "application/json", sanitized["Content-Type"])
	})

	t.Run("masks secrets case-insensitively (lowercase keys)", func(t *testing.T) {
		headers := map[string]string{
			"authorization": "Bearer token123",
			"cookie":        "session=xyz",
			"x-api-key":     "key456",
			"content-type":  "application/json",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "Bearer ***MASKED***", sanitized["authorization"])
		assert.Equal(t, "***MASKED***", sanitized["cookie"])
		assert.Equal(t, "***MASKED***", sanitized["x-api-key"])
		assert.Equal(t, "application/json", sanitized["content-type"])
	})

	t.Run("masks secrets case-insensitively (mixed-case keys)", func(t *testing.T) {
		headers := map[string]string{
			"AuThOrIzAtIoN": "Bearer token123",
			"cOoKiE":        "session=xyz",
			"X-aPi-kEy":     "key456",
		}

		sanitized := sanitizeHeaders(httpClient, headers)

		assert.Equal(t, "Bearer ***MASKED***", sanitized["AuThOrIzAtIoN"])
		assert.Equal(t, "***MASKED***", sanitized["cOoKiE"])
		assert.Equal(t, "***MASKED***", sanitized["X-aPi-kEy"])
	})
}

func TestMaskHeaderValue(t *testing.T) {
	t.Run("masks Bearer token", func(t *testing.T) {
		result := maskHeaderValue("Authorization", "Bearer abc123")
		assert.Equal(t, "Bearer ***MASKED***", result)
	})

	t.Run("masks Basic auth", func(t *testing.T) {
		result := maskHeaderValue("Authorization", "Basic xyz789")
		assert.Equal(t, "Basic ***MASKED***", result)
	})

	t.Run("masks plain authorization", func(t *testing.T) {
		result := maskHeaderValue("Authorization", "token123")
		assert.Equal(t, "***MASKED***", result)
	})

	t.Run("masks cookie", func(t *testing.T) {
		result := maskHeaderValue("Cookie", "session=abc; user=john")
		assert.Equal(t, "***MASKED***", result)
	})

	t.Run("masks custom secret header", func(t *testing.T) {
		result := maskHeaderValue("X-Secret-Key", "my_secret")
		assert.Equal(t, "***MASKED***", result)
	})

	t.Run("authorization masking is case-insensitive", func(t *testing.T) {
		result := maskHeaderValue("aUtHoRiZaTiOn", "Bearer abc123")
		assert.Equal(t, "Bearer ***MASKED***", result)
	})
}

func TestCustomSecretHeaders(t *testing.T) {
	httpClient := client.New(client.Config{
		BaseURL: "https://example.com",
		SecretHeaders: []string{
			"X-Custom-Secret",
			"X-API-Token",
		},
	})

	headers := map[string]string{
		"X-Custom-Secret": "secret123",
		"X-API-Token":     "token456",
		"Authorization":   "Bearer should_not_be_masked",
		"Content-Type":    "application/json",
	}

	sanitized := sanitizeHeaders(httpClient, headers)

	assert.Equal(t, "***MASKED***", sanitized["X-Custom-Secret"])
	assert.Equal(t, "***MASKED***", sanitized["X-API-Token"])
	assert.Equal(t, "Bearer should_not_be_masked", sanitized["Authorization"])
	assert.Equal(t, "application/json", sanitized["Content-Type"])
}

func TestCustomSecretHeaders_CaseInsensitive(t *testing.T) {
	httpClient := client.New(client.Config{
		BaseURL: "https://example.com",
		SecretHeaders: []string{
			"X-Custom-Secret",
		},
	})

	headers := map[string]string{
		"x-custom-secret": "secret123",
		"X-CUSTOM-SECRET": "secret456",
	}

	sanitized := sanitizeHeaders(httpClient, headers)

	assert.Equal(t, "***MASKED***", sanitized["x-custom-secret"])
	assert.Equal(t, "***MASKED***", sanitized["X-CUSTOM-SECRET"])
}

func TestSanitizeHeaders_EdgeCases(t *testing.T) {
	httpClient := client.New(client.Config{BaseURL: "https://example.com"})

	t.Run("does not mutate original map", func(t *testing.T) {
		original := map[string]string{
			"Authorization": "Bearer secret",
			"Public":        "value",
		}

		originalCopy := map[string]string{
			"Authorization": "Bearer secret",
			"Public":        "value",
		}

		sanitized := sanitizeHeaders(httpClient, original)

		assert.Equal(t, "Bearer ***MASKED***", sanitized["Authorization"])
		assert.Equal(t, originalCopy, original)
	})

	t.Run("handles nil input map", func(t *testing.T) {
		var headers map[string]string = nil
		sanitized := sanitizeHeaders(httpClient, headers)

		assert.NotNil(t, sanitized)
		assert.Empty(t, sanitized)
	})

	t.Run("handles nil client", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token",
			"Content-Type":  "application/json",
		}

		sanitized := sanitizeHeaders(nil, headers)

		assert.Equal(t, "Bearer token", sanitized["Authorization"])
		assert.Equal(t, "application/json", sanitized["Content-Type"])
	})
}

func TestMaskHeaderValue_AuthorizationEdgeCases(t *testing.T) {
	t.Run("empty value", func(t *testing.T) {
		assert.Equal(t, "***MASKED***", maskHeaderValue("Authorization", ""))
	})

	t.Run("only spaces", func(t *testing.T) {
		assert.Equal(t, "***MASKED***", maskHeaderValue("Authorization", "   "))
	})

	t.Run("scheme without token", func(t *testing.T) {
		assert.Equal(t, "***MASKED***", maskHeaderValue("Authorization", "Bearer"))
	})

	t.Run("scheme with trailing space only", func(t *testing.T) {
		assert.Equal(t, "***MASKED***", maskHeaderValue("Authorization", "Bearer "))
	})

	t.Run("multiple spaces between scheme and token", func(t *testing.T) {
		assert.Equal(t, "Bearer ***MASKED***", maskHeaderValue("Authorization", "Bearer   token"))
	})
}
