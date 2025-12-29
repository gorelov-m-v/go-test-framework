package httpdsl

import (
	"testing"
	"time"

	"go-test-framework/pkg/httpclient"

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
	client := httpclient.New(httpclient.Config{BaseURL: "https://api.example.com"})

	body := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	req := &httpclient.Request[map[string]interface{}]{
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

	attachRequest(mockCtx, client, req)

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
	assert.Contains(t, content, "  Authorization: Bearer ***")

	assert.Contains(t, content, "Body (json):")
	assert.Contains(t, content, `"name": "John"`)
	assert.Contains(t, content, `"age": 30`)
}

func TestAttachRequest_RawBody_Truncated(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	client := httpclient.New(httpclient.Config{BaseURL: "https://api.example.com"})

	largeBody := make([]byte, 1050)
	for i := range largeBody {
		largeBody[i] = 'X'
	}

	req := &httpclient.Request[any]{
		Method:  "POST",
		Path:    "/upload",
		RawBody: largeBody,
	}

	attachRequest(mockCtx, client, req)

	content := string(mockCtx.Attachments[0].Content)

	assert.Contains(t, content, "Body (raw):")
	assert.Contains(t, content, "(raw body, 1050 bytes)")
	assert.Contains(t, content, "XXXX")
	assert.Contains(t, content, "...")
}

func TestAttachRequest_Multipart(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	client := httpclient.New(httpclient.Config{})

	req := &httpclient.Request[any]{
		Method: "POST",
		Path:   "/files",
		Multipart: &httpclient.MultipartForm{
			Fields: map[string]string{"type": "avatar"},
			Files: []httpclient.MultipartFile{
				{FieldName: "file", FileName: "image.png", Content: []byte{1, 2, 3}},
			},
		},
	}

	attachRequest(mockCtx, client, req)

	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Body (multipart/form-data):")
	assert.Contains(t, content, "  type: avatar")
	assert.Contains(t, content, "  file: image.png (3 bytes)")
}

func TestAttachRequest_EmptyBody(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}
	client := httpclient.New(httpclient.Config{})

	req := &httpclient.Request[any]{Method: "GET", Path: "/"}

	attachRequest(mockCtx, client, req)
	content := string(mockCtx.Attachments[0].Content)
	assert.Contains(t, content, "Body: (empty)")
}

func TestAttachRequest_InvalidBaseURL(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	client := httpclient.New(httpclient.Config{BaseURL: "://invalid"})
	req := &httpclient.Request[any]{Method: "GET", Path: "/"}

	attachRequest(mockCtx, client, req)
	content := string(mockCtx.Attachments[0].Content)

	assert.Contains(t, content, "Effective URL: (failed to resolve:")
}

func TestAttachResponse_Success_JSON(t *testing.T) {
	mockCtx := &MockStepCtxAttachment{}

	rawJSON := []byte(`{"status":"ok", "id": 1}`)
	resp := &httpclient.Response[any]{
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

	resp := &httpclient.Response[any]{
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

	resp := &httpclient.Response[any]{
		StatusCode: 400,
		Error: &httpclient.ErrorResponse{
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

	resp := &httpclient.Response[any]{
		StatusCode: 500,
		Error: &httpclient.ErrorResponse{
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

	resp := &httpclient.Response[any]{
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
	resp := &httpclient.Response[any]{
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
