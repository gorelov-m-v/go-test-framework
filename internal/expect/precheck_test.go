package expect

import (
	"errors"
	"testing"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type mockResponse struct {
	body         []byte
	networkError string
	err          error
	isNil        bool
}

func TestBuildPreCheck_ErrorHandling(t *testing.T) {
	cfg := PreCheckConfig[*mockResponse]{
		IsNil: func(r *mockResponse) bool { return r == nil || r.isNil },
	}
	preCheck := BuildPreCheck(cfg)

	t.Run("returns failure when err is not nil", func(t *testing.T) {
		result, ok := preCheck(errors.New("some error"), &mockResponse{})
		if ok {
			t.Error("Expected preCheck to fail when error is provided")
		}
		if !result.Retryable {
			t.Error("Expected result to be retryable")
		}
		if result.Reason != "Request failed" {
			t.Errorf("Expected reason 'Request failed', got '%s'", result.Reason)
		}
	})

	t.Run("returns failure when response is nil", func(t *testing.T) {
		result, ok := preCheck(nil, nil)
		if ok {
			t.Error("Expected preCheck to fail when response is nil")
		}
		if result.Reason != "Response is nil" {
			t.Errorf("Expected reason 'Response is nil', got '%s'", result.Reason)
		}
	})

	t.Run("returns success when no errors", func(t *testing.T) {
		_, ok := preCheck(nil, &mockResponse{})
		if !ok {
			t.Error("Expected preCheck to succeed with valid response")
		}
	})
}

func TestBuildPreCheck_WithNetworkError(t *testing.T) {
	cfg := PreCheckConfig[*mockResponse]{
		IsNil:           func(r *mockResponse) bool { return r == nil },
		GetNetworkError: func(r *mockResponse) string { return r.networkError },
	}
	preCheck := BuildPreCheck(cfg)

	t.Run("returns failure on network error", func(t *testing.T) {
		result, ok := preCheck(nil, &mockResponse{networkError: "connection refused"})
		if ok {
			t.Error("Expected preCheck to fail on network error")
		}
		if result.Reason != "Network error occurred" {
			t.Errorf("Expected reason 'Network error occurred', got '%s'", result.Reason)
		}
	})

	t.Run("returns success without network error", func(t *testing.T) {
		_, ok := preCheck(nil, &mockResponse{networkError: ""})
		if !ok {
			t.Error("Expected preCheck to succeed without network error")
		}
	})
}

func TestBuildPreCheck_WithHasError(t *testing.T) {
	cfg := PreCheckConfig[*mockResponse]{
		IsNil:    func(r *mockResponse) bool { return r == nil },
		HasError: func(r *mockResponse) error { return r.err },
	}
	preCheck := BuildPreCheck(cfg)

	t.Run("returns failure when response has error", func(t *testing.T) {
		result, ok := preCheck(nil, &mockResponse{err: errors.New("internal error")})
		if ok {
			t.Error("Expected preCheck to fail when response has error")
		}
		if result.Reason != "Response contains error" {
			t.Errorf("Expected reason 'Response contains error', got '%s'", result.Reason)
		}
	})

	t.Run("returns success when response has no error", func(t *testing.T) {
		_, ok := preCheck(nil, &mockResponse{err: nil})
		if !ok {
			t.Error("Expected preCheck to succeed when response has no error")
		}
	})
}

func TestBuildPreCheckWithBody(t *testing.T) {
	cfg := PreCheckConfig[*mockResponse]{
		IsNil:          func(r *mockResponse) bool { return r == nil },
		EmptyBodyCheck: func(r *mockResponse) bool { return len(r.body) == 0 },
	}
	preCheckWithBody := BuildPreCheckWithBody(cfg)

	t.Run("returns failure when body is empty", func(t *testing.T) {
		result, ok := preCheckWithBody(nil, &mockResponse{body: []byte{}})
		if ok {
			t.Error("Expected preCheckWithBody to fail when body is empty")
		}
		if result.Reason != "Response body is empty" {
			t.Errorf("Expected reason 'Response body is empty', got '%s'", result.Reason)
		}
	})

	t.Run("returns success when body is not empty", func(t *testing.T) {
		_, ok := preCheckWithBody(nil, &mockResponse{body: []byte(`{"data": "test"}`)})
		if !ok {
			t.Error("Expected preCheckWithBody to succeed when body is not empty")
		}
	})

	t.Run("base preCheck runs first", func(t *testing.T) {
		result, ok := preCheckWithBody(errors.New("error"), &mockResponse{body: []byte(`{"data": "test"}`)})
		if ok {
			t.Error("Expected preCheckWithBody to fail when base preCheck fails")
		}
		if result.Reason != "Request failed" {
			t.Errorf("Expected reason from base preCheck, got '%s'", result.Reason)
		}
	})
}

func TestBuildSimplePreCheck(t *testing.T) {
	cfg := SimplePreCheckConfig[*mockResponse]{
		IsNil:    func(r *mockResponse) bool { return r == nil },
		HasError: func(r *mockResponse) error { return r.err },
	}
	preCheck := BuildSimplePreCheck(cfg)

	t.Run("returns failure when response is nil", func(t *testing.T) {
		_, ok := preCheck(nil, nil)
		if ok {
			t.Error("Expected preCheck to fail when response is nil")
		}
	})

	t.Run("returns failure when has error", func(t *testing.T) {
		_, ok := preCheck(nil, &mockResponse{err: errors.New("error")})
		if ok {
			t.Error("Expected preCheck to fail when response has error")
		}
	})

	t.Run("returns success for valid response", func(t *testing.T) {
		_, ok := preCheck(nil, &mockResponse{})
		if !ok {
			t.Error("Expected preCheck to succeed")
		}
	})
}

type mockRedisResult struct {
	key    string
	exists bool
	err    error
}

func TestBuildKeyExistsPreCheck(t *testing.T) {
	basePreCheck := func(err error, r *mockRedisResult) (polling.CheckResult, bool) {
		if err != nil {
			return polling.CheckResult{Ok: false, Retryable: true, Reason: "Query failed"}, false
		}
		if r == nil {
			return polling.CheckResult{Ok: false, Retryable: true, Reason: "Result is nil"}, false
		}
		return polling.CheckResult{}, true
	}

	cfg := KeyExistsPreCheckConfig[*mockRedisResult]{
		BasePreCheck: basePreCheck,
		KeyExists:    func(r *mockRedisResult) bool { return r.exists },
		GetKey:       func(r *mockRedisResult) string { return r.key },
	}
	preCheck := BuildKeyExistsPreCheck(cfg)

	t.Run("returns failure when key does not exist", func(t *testing.T) {
		result, ok := preCheck(nil, &mockRedisResult{key: "test:key", exists: false})
		if ok {
			t.Error("Expected preCheck to fail when key does not exist")
		}
		expectedReason := "Key 'test:key' does not exist"
		if result.Reason != expectedReason {
			t.Errorf("Expected reason '%s', got '%s'", expectedReason, result.Reason)
		}
	})

	t.Run("returns success when key exists", func(t *testing.T) {
		_, ok := preCheck(nil, &mockRedisResult{key: "test:key", exists: true})
		if !ok {
			t.Error("Expected preCheck to succeed when key exists")
		}
	})

	t.Run("base preCheck runs first", func(t *testing.T) {
		result, ok := preCheck(errors.New("redis error"), &mockRedisResult{exists: true})
		if ok {
			t.Error("Expected preCheck to fail when base preCheck fails")
		}
		if result.Reason != "Query failed" {
			t.Errorf("Expected reason from base preCheck, got '%s'", result.Reason)
		}
	})
}

func TestBuildKeyExistsPreCheck_WithoutGetKey(t *testing.T) {
	cfg := KeyExistsPreCheckConfig[*mockRedisResult]{
		KeyExists: func(r *mockRedisResult) bool { return r.exists },
	}
	preCheck := BuildKeyExistsPreCheck(cfg)

	t.Run("returns generic message when GetKey is nil", func(t *testing.T) {
		result, ok := preCheck(nil, &mockRedisResult{key: "test:key", exists: false})
		if ok {
			t.Error("Expected preCheck to fail when key does not exist")
		}
		expectedReason := "Key does not exist"
		if result.Reason != expectedReason {
			t.Errorf("Expected reason '%s', got '%s'", expectedReason, result.Reason)
		}
	})
}

func TestFormatKeyNotExistsReason(t *testing.T) {
	t.Run("with key", func(t *testing.T) {
		reason := formatKeyNotExistsReason("user:123")
		expected := "Key 'user:123' does not exist"
		if reason != expected {
			t.Errorf("Expected '%s', got '%s'", expected, reason)
		}
	})

	t.Run("without key", func(t *testing.T) {
		reason := formatKeyNotExistsReason("")
		expected := "Key does not exist"
		if reason != expected {
			t.Errorf("Expected '%s', got '%s'", expected, reason)
		}
	})
}
