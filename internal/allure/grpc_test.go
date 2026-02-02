package allure

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	grpcClient "github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

func TestWriteGRPCError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains []string
	}{
		{
			name: "gRPC status error",
			err:  status.Error(codes.NotFound, "user not found"),
			contains: []string{
				"Error:",
				"Code: NotFound",
				"Message: user not found",
			},
		},
		{
			name: "generic error",
			err:  errors.New("some generic error"),
			contains: []string{
				"Error:",
				"Message: some generic error",
			},
		},
		{
			name: "permission denied",
			err:  status.Error(codes.PermissionDenied, "access denied"),
			contains: []string{
				"Code: PermissionDenied",
				"Message: access denied",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewDefaultReporter()
			builder := NewReportBuilder()

			reporter.writeGRPCError(builder, tt.err)

			content := builder.String()

			for _, expected := range tt.contains {
				assert.Contains(t, content, expected)
			}
		})
	}
}

func TestWriteBody(t *testing.T) {
	tests := []struct {
		name     string
		body     any
		contains []string
	}{
		{
			name: "nil body",
			body: nil,
		},
		{
			name: "map body",
			body: map[string]string{"id": "123", "name": "John"},
			contains: []string{
				"Body:",
				`"id": "123"`,
				`"name": "John"`,
			},
		},
		{
			name: "struct body",
			body: struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}{ID: "456", Name: "Jane"},
			contains: []string{
				"Body:",
				`"id": "456"`,
				`"name": "Jane"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewDefaultReporter()
			builder := NewReportBuilder()

			reporter.writeBody(builder, tt.body)

			content := builder.String()

			if tt.body == nil {
				assert.Empty(t, strings.TrimSpace(content))
			} else {
				for _, expected := range tt.contains {
					assert.Contains(t, content, expected)
				}
			}
		})
	}
}

func TestGRPCRequestDTO(t *testing.T) {
	dto := GRPCRequestDTO{
		Target:   "localhost:9090",
		Method:   "/user.UserService/GetUser",
		Metadata: metadata.Pairs("key", "value"),
		Body:     map[string]string{"id": "123"},
	}

	assert.Equal(t, "localhost:9090", dto.Target)
	assert.Equal(t, "/user.UserService/GetUser", dto.Method)
	assert.NotNil(t, dto.Metadata)
	assert.NotNil(t, dto.Body)
}

func TestGRPCResponseDTO(t *testing.T) {
	err := status.Error(codes.NotFound, "not found")
	dto := GRPCResponseDTO{
		StatusCode: 5,
		Status:     "NOT_FOUND",
		Metadata:   metadata.Pairs("key", "value"),
		Body:       map[string]string{"error": "not found"},
		Error:      err,
		Duration:   100 * time.Millisecond,
	}

	assert.Equal(t, 5, dto.StatusCode)
	assert.Equal(t, "NOT_FOUND", dto.Status)
	assert.NotNil(t, dto.Metadata)
	assert.NotNil(t, dto.Body)
	assert.NotNil(t, dto.Error)
	assert.Equal(t, 100*time.Millisecond, dto.Duration)
}

func TestToGRPCRequestDTO(t *testing.T) {
	t.Run("with body", func(t *testing.T) {
		body := &struct{ ID string }{ID: "123"}
		md := metadata.Pairs("auth", "token")
		dto := ToGRPCRequestDTO("localhost:9090", "/svc/Method", body, md)

		assert.Equal(t, "localhost:9090", dto.Target)
		assert.Equal(t, "/svc/Method", dto.Method)
		assert.Equal(t, md, dto.Metadata)
		assert.Equal(t, body, dto.Body)
	})

	t.Run("nil body", func(t *testing.T) {
		dto := ToGRPCRequestDTO[struct{}]("host:80", "/method", nil, nil)

		assert.Equal(t, "host:80", dto.Target)
		assert.Equal(t, "/method", dto.Method)
		assert.Nil(t, dto.Body)
	})
}

func TestToGRPCResponseDTO(t *testing.T) {
	t.Run("nil response", func(t *testing.T) {
		dto := ToGRPCResponseDTO[any](nil)

		assert.Equal(t, GRPCResponseDTO{}, dto)
	})

	t.Run("success response", func(t *testing.T) {
		body := struct{ Name string }{Name: "test"}
		resp := &grpcClient.Response[struct{ Name string }]{
			Body:     &body,
			Duration: 50 * time.Millisecond,
			Metadata: metadata.Pairs("key", "val"),
		}

		dto := ToGRPCResponseDTO(resp)

		assert.Equal(t, "OK", dto.Status)
		assert.Equal(t, 0, dto.StatusCode)
		assert.Equal(t, &body, dto.Body)
		assert.Equal(t, 50*time.Millisecond, dto.Duration)
	})

	t.Run("grpc status error", func(t *testing.T) {
		resp := &grpcClient.Response[any]{
			Error:    status.Error(codes.NotFound, "not found"),
			Duration: 10 * time.Millisecond,
		}

		dto := ToGRPCResponseDTO(resp)

		assert.Equal(t, "NotFound", dto.Status)
		assert.Equal(t, int(codes.NotFound), dto.StatusCode)
		assert.NotNil(t, dto.Error)
	})

	t.Run("generic error", func(t *testing.T) {
		resp := &grpcClient.Response[any]{
			Error: errors.New("connection refused"),
		}

		dto := ToGRPCResponseDTO(resp)

		assert.Equal(t, "UNKNOWN", dto.Status)
		assert.Equal(t, -1, dto.StatusCode)
	})
}
