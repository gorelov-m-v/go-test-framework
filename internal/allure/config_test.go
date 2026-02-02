package allure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Empty(t, cfg.SensitiveHeaders)
	assert.Empty(t, cfg.SensitiveFields)
	assert.Equal(t, "***MASKED***", cfg.MaskValue)
}

func TestShouldMaskField(t *testing.T) {
	cfg := MaskingConfig{
		SensitiveFields: []string{"password", "api_token"},
	}

	t.Run("matches sensitive field", func(t *testing.T) {
		assert.True(t, cfg.ShouldMaskField("password"))
		assert.True(t, cfg.ShouldMaskField("PASSWORD"))
		assert.True(t, cfg.ShouldMaskField("  Password  "))
		assert.True(t, cfg.ShouldMaskField("api_token"))
	})

	t.Run("does not match", func(t *testing.T) {
		assert.False(t, cfg.ShouldMaskField("username"))
		assert.False(t, cfg.ShouldMaskField("email"))
	})

	t.Run("empty config", func(t *testing.T) {
		emptyCfg := MaskingConfig{}
		assert.False(t, emptyCfg.ShouldMaskField("password"))
	})
}

func TestShouldMaskHeader(t *testing.T) {
	cfg := MaskingConfig{
		SensitiveHeaders: []string{"Authorization", "X-Api-Key"},
	}

	t.Run("matches sensitive header", func(t *testing.T) {
		assert.True(t, cfg.ShouldMaskHeader("Authorization"))
		assert.True(t, cfg.ShouldMaskHeader("authorization"))
		assert.True(t, cfg.ShouldMaskHeader("  AUTHORIZATION  "))
		assert.True(t, cfg.ShouldMaskHeader("X-Api-Key"))
	})

	t.Run("does not match", func(t *testing.T) {
		assert.False(t, cfg.ShouldMaskHeader("Content-Type"))
		assert.False(t, cfg.ShouldMaskHeader("Accept"))
	})

	t.Run("empty config", func(t *testing.T) {
		emptyCfg := MaskingConfig{}
		assert.False(t, emptyCfg.ShouldMaskHeader("Authorization"))
	})
}

func TestMaskHeader(t *testing.T) {
	cfg := MaskingConfig{
		SensitiveHeaders: []string{"Authorization", "X-Api-Key"},
		MaskValue:        "***MASKED***",
	}

	t.Run("non-sensitive header unchanged", func(t *testing.T) {
		result := cfg.MaskHeader("Content-Type", "application/json")
		assert.Equal(t, "application/json", result)
	})

	t.Run("authorization header preserves scheme", func(t *testing.T) {
		result := cfg.MaskHeader("Authorization", "Bearer token123")
		assert.Equal(t, "Bearer ***MASKED***", result)

		result = cfg.MaskHeader("Authorization", "Basic abc123")
		assert.Equal(t, "Basic ***MASKED***", result)
	})

	t.Run("authorization without scheme", func(t *testing.T) {
		result := cfg.MaskHeader("Authorization", "token123")
		assert.Equal(t, "***MASKED***", result)
	})

	t.Run("other sensitive header fully masked", func(t *testing.T) {
		result := cfg.MaskHeader("X-Api-Key", "secret-key-123")
		assert.Equal(t, "***MASKED***", result)
	})
}

func TestShouldMaskValue(t *testing.T) {
	cfg := MaskingConfig{
		SensitiveFields: []string{"secret", "token"},
	}

	t.Run("matches", func(t *testing.T) {
		assert.True(t, cfg.ShouldMaskValue("secret"))
		assert.True(t, cfg.ShouldMaskValue("SECRET"))
		assert.True(t, cfg.ShouldMaskValue("token"))
	})

	t.Run("does not match", func(t *testing.T) {
		assert.False(t, cfg.ShouldMaskValue("username"))
	})

	t.Run("empty config", func(t *testing.T) {
		emptyCfg := MaskingConfig{}
		assert.False(t, emptyCfg.ShouldMaskValue("secret"))
	})
}
