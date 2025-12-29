package httpclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeHeaders(t *testing.T) {
	client := New(Config{
		BaseURL: "https://example.com",
	})

	t.Run("masks Authorization Bearer token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer secret_token_123",
			"Content-Type":  "application/json",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "Bearer ***", sanitized["Authorization"])
		assert.Equal(t, "application/json", sanitized["Content-Type"])
	})

	t.Run("masks Authorization Basic token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Basic dXNlcjpwYXNz",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "Basic ***", sanitized["Authorization"])
	})

	t.Run("masks plain Authorization token", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "secret_token",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "***", sanitized["Authorization"])
	})

	t.Run("masks Cookie header", func(t *testing.T) {
		headers := map[string]string{
			"Cookie": "session=abc123; user=john",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "***", sanitized["Cookie"])
	})

	t.Run("masks Api-Key header", func(t *testing.T) {
		headers := map[string]string{
			"Api-Key": "my_secret_key",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "***", sanitized["Api-Key"])
	})

	t.Run("preserves non-secret headers", func(t *testing.T) {
		headers := map[string]string{
			"Content-Type":  "application/json",
			"Accept":        "application/json",
			"User-Agent":    "Test Client",
			"X-Request-ID":  "12345",
			"Cache-Control": "no-cache",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, headers, sanitized)
	})

	t.Run("handles empty headers", func(t *testing.T) {
		headers := map[string]string{}

		sanitized := client.SanitizeHeaders(headers)

		assert.Empty(t, sanitized)
	})

	t.Run("masks multiple secret headers", func(t *testing.T) {
		headers := map[string]string{
			"Authorization": "Bearer token123",
			"Cookie":        "session=xyz",
			"X-Api-Key":     "key456",
			"Content-Type":  "application/json",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "Bearer ***", sanitized["Authorization"])
		assert.Equal(t, "***", sanitized["Cookie"])
		assert.Equal(t, "***", sanitized["X-Api-Key"])
		assert.Equal(t, "application/json", sanitized["Content-Type"])
	})

	t.Run("masks secrets case-insensitively (lowercase keys)", func(t *testing.T) {
		headers := map[string]string{
			"authorization": "Bearer token123",
			"cookie":        "session=xyz",
			"x-api-key":     "key456",
			"content-type":  "application/json",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "Bearer ***", sanitized["authorization"])
		assert.Equal(t, "***", sanitized["cookie"])
		assert.Equal(t, "***", sanitized["x-api-key"])
		assert.Equal(t, "application/json", sanitized["content-type"])
	})

	t.Run("masks secrets case-insensitively (mixed-case keys)", func(t *testing.T) {
		headers := map[string]string{
			"AuThOrIzAtIoN": "Bearer token123",
			"cOoKiE":        "session=xyz",
			"X-aPi-kEy":     "key456",
		}

		sanitized := client.SanitizeHeaders(headers)

		assert.Equal(t, "Bearer ***", sanitized["AuThOrIzAtIoN"])
		assert.Equal(t, "***", sanitized["cOoKiE"])
		assert.Equal(t, "***", sanitized["X-aPi-kEy"])
	})
}

func TestMaskHeaderValue(t *testing.T) {
	t.Run("masks Bearer token", func(t *testing.T) {
		result := maskHeaderValue("Authorization", "Bearer abc123")
		assert.Equal(t, "Bearer ***", result)
	})

	t.Run("masks Basic auth", func(t *testing.T) {
		result := maskHeaderValue("Authorization", "Basic xyz789")
		assert.Equal(t, "Basic ***", result)
	})

	t.Run("masks plain authorization", func(t *testing.T) {
		result := maskHeaderValue("Authorization", "token123")
		assert.Equal(t, "***", result)
	})

	t.Run("masks cookie", func(t *testing.T) {
		result := maskHeaderValue("Cookie", "session=abc; user=john")
		assert.Equal(t, "***", result)
	})

	t.Run("masks custom secret header", func(t *testing.T) {
		result := maskHeaderValue("X-Secret-Key", "my_secret")
		assert.Equal(t, "***", result)
	})

	t.Run("authorization masking is case-insensitive", func(t *testing.T) {
		result := maskHeaderValue("aUtHoRiZaTiOn", "Bearer abc123")
		assert.Equal(t, "Bearer ***", result)
	})
}

func TestCustomSecretHeaders(t *testing.T) {
	client := New(Config{
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

	sanitized := client.SanitizeHeaders(headers)

	assert.Equal(t, "***", sanitized["X-Custom-Secret"])
	assert.Equal(t, "***", sanitized["X-API-Token"])
	assert.Equal(t, "Bearer should_not_be_masked", sanitized["Authorization"])
	assert.Equal(t, "application/json", sanitized["Content-Type"])
}

func TestCustomSecretHeaders_CaseInsensitive(t *testing.T) {
	client := New(Config{
		BaseURL: "https://example.com",
		SecretHeaders: []string{
			"X-Custom-Secret",
		},
	})

	headers := map[string]string{
		"x-custom-secret": "secret123",
		"X-CUSTOM-SECRET": "secret456",
	}

	sanitized := client.SanitizeHeaders(headers)

	assert.Equal(t, "***", sanitized["x-custom-secret"])
	assert.Equal(t, "***", sanitized["X-CUSTOM-SECRET"])
}

func TestSanitizeHeaders_EdgeCases(t *testing.T) {
	client := New(Config{BaseURL: "https://example.com"})

	t.Run("does not mutate original map", func(t *testing.T) {
		original := map[string]string{
			"Authorization": "Bearer secret",
			"Public":        "value",
		}

		originalCopy := map[string]string{
			"Authorization": "Bearer secret",
			"Public":        "value",
		}

		sanitized := client.SanitizeHeaders(original)

		assert.Equal(t, "Bearer ***", sanitized["Authorization"])
		assert.Equal(t, originalCopy, original)
	})

	t.Run("handles nil input map", func(t *testing.T) {
		var headers map[string]string = nil
		sanitized := client.SanitizeHeaders(headers)

		assert.NotNil(t, sanitized)
		assert.Empty(t, sanitized)
	})

	t.Run("handles trimming in maskHeaderValue keys", func(t *testing.T) {
		headers := map[string]string{
			" Authorization ": "Bearer token",
		}

		sanitized := client.SanitizeHeaders(headers)

		_ = sanitized
	})
}

func TestMaskHeaderValue_AuthorizationEdgeCases(t *testing.T) {
	t.Run("empty value", func(t *testing.T) {
		assert.Equal(t, "***", maskHeaderValue("Authorization", ""))
	})

	t.Run("only spaces", func(t *testing.T) {
		assert.Equal(t, "***", maskHeaderValue("Authorization", "   "))
	})

	t.Run("scheme without token", func(t *testing.T) {
		assert.Equal(t, "***", maskHeaderValue("Authorization", "Bearer"))
	})

	t.Run("scheme with trailing space only", func(t *testing.T) {
		assert.Equal(t, "***", maskHeaderValue("Authorization", "Bearer "))
	})

	t.Run("multiple spaces between scheme and token", func(t *testing.T) {
		assert.Equal(t, "Bearer ***", maskHeaderValue("Authorization", "Bearer   token"))
	})
}
