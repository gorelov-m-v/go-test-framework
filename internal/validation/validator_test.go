package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockStepBreaker struct {
	brokenMessage string
	brokenNow     bool
}

func (m *mockStepBreaker) Break(args ...interface{}) {
	if len(args) > 0 {
		m.brokenMessage = fmt.Sprint(args...)
	}
}

func (m *mockStepBreaker) BrokenNow() {
	m.brokenNow = true
}

func newMockStepBreaker() *mockStepBreaker {
	return &mockStepBreaker{}
}

func TestNewWithBreaker(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	assert.NotNil(t, v)
	assert.Equal(t, "HTTP", v.dslName)
	assert.False(t, v.failed)
}

func TestRequireNotNil_Success(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	value := "test"
	result := v.RequireNotNil(&value, "pointer")

	assert.True(t, result)
	assert.False(t, mock.brokenNow)
	assert.Empty(t, mock.brokenMessage)
}

func TestRequireNotNil_NilValue(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireNotNil(nil, "client")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "HTTP DSL Error")
	assert.Contains(t, mock.brokenMessage, "client is nil")
}

func TestRequireNotNil_NilPointer(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	var ptr *string = nil
	result := v.RequireNotNil(ptr, "pointer")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "pointer is nil")
}

func TestRequireNotNil_AlreadyFailed(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")
	v.failed = true

	result := v.RequireNotNil("value", "field")

	assert.False(t, result)
	assert.Empty(t, mock.brokenMessage)
}

func TestRequireNotEmpty_Success(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireNotEmpty("value", "field")

	assert.True(t, result)
	assert.False(t, mock.brokenNow)
}

func TestRequireNotEmpty_Empty(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireNotEmpty("", "method")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "HTTP DSL Error")
	assert.Contains(t, mock.brokenMessage, "method is not set")
}

func TestRequireNotEmpty_Whitespace(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireNotEmpty("   ", "path")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "path is not set")
}

func TestRequireNotEmpty_AlreadyFailed(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")
	v.failed = true

	result := v.RequireNotEmpty("value", "field")

	assert.False(t, result)
}

func TestRequireNotEmptyWithHint_Success(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireNotEmptyWithHint("value", "field", "Use .GET() method")

	assert.True(t, result)
	assert.False(t, mock.brokenNow)
}

func TestRequireNotEmptyWithHint_Empty(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireNotEmptyWithHint("", "method", "Use .GET(), .POST() etc.")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "HTTP DSL Error")
	assert.Contains(t, mock.brokenMessage, "method is not set")
	assert.Contains(t, mock.brokenMessage, "Use .GET(), .POST() etc.")
}

func TestRequireNotEmptyWithHint_AlreadyFailed(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")
	v.failed = true

	result := v.RequireNotEmptyWithHint("", "field", "hint")

	assert.False(t, result)
}

func TestRequireStruct_Success(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	type Request struct{ ID int }
	result := v.RequireStruct(Request{ID: 1}, "request")

	assert.True(t, result)
	assert.False(t, mock.brokenNow)
}

func TestRequireStruct_Nil(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireStruct(nil, "request")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "must be a struct")
	assert.Contains(t, mock.brokenMessage, "got nil")
}

func TestRequireStruct_NotStruct(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.RequireStruct("string value", "request")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "must be a struct")
}

func TestRequireStruct_Pointer(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	type Request struct{ ID int }
	req := &Request{ID: 1}
	result := v.RequireStruct(req, "request")

	assert.False(t, result)
	assert.Contains(t, mock.brokenMessage, "must be a struct")
}

func TestRequireStruct_AlreadyFailed(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")
	v.failed = true

	type Request struct{ ID int }
	result := v.RequireStruct(Request{ID: 1}, "request")

	assert.False(t, result)
}

func TestRequire_Success(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.Require(true, "something must be true")

	assert.True(t, result)
	assert.False(t, mock.brokenNow)
}

func TestRequire_Failure(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	result := v.Require(false, "timeout must be positive")

	assert.False(t, result)
	assert.True(t, mock.brokenNow)
	assert.Contains(t, mock.brokenMessage, "HTTP DSL Error")
	assert.Contains(t, mock.brokenMessage, "timeout must be positive")
}

func TestRequire_AlreadyFailed(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")
	v.failed = true

	result := v.Require(true, "message")

	assert.False(t, result)
}

func TestValidator_ChainedValidation(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "HTTP")

	v.RequireNotEmpty("GET", "method")
	v.RequireNotEmpty("/api", "path")
	v.Require(true, "valid config")

	assert.False(t, v.failed)
	assert.False(t, mock.brokenNow)
}

func TestValidator_ChainedValidation_FirstFails(t *testing.T) {
	mock := newMockStepBreaker()
	v := NewWithBreaker(mock, "gRPC")

	v.RequireNotEmpty("", "method")
	result := v.RequireNotEmpty("/api", "path")

	assert.True(t, v.failed)
	assert.False(t, result)
}

func TestValidator_DifferentDSLNames(t *testing.T) {
	tests := []struct {
		dslName  string
		expected string
	}{
		{"HTTP", "HTTP DSL Error"},
		{"gRPC", "gRPC DSL Error"},
		{"Kafka", "Kafka DSL Error"},
		{"Redis", "Redis DSL Error"},
		{"Database", "Database DSL Error"},
	}

	for _, tt := range tests {
		t.Run(tt.dslName, func(t *testing.T) {
			mock := newMockStepBreaker()
			v := NewWithBreaker(mock, tt.dslName)

			v.RequireNotEmpty("", "field")

			assert.Contains(t, mock.brokenMessage, tt.expected)
		})
	}
}
