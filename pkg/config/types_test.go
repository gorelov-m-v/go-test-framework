package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAsyncConfig(t *testing.T) {
	cfg := DefaultAsyncConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Equal(t, 200*time.Millisecond, cfg.Interval)
	assert.Equal(t, 0.2, cfg.Jitter)

	assert.True(t, cfg.Backoff.Enabled)
	assert.Equal(t, 1.5, cfg.Backoff.Factor)
	assert.Equal(t, 1*time.Second, cfg.Backoff.MaxInterval)
}

func TestAsyncConfig_WithDefaults_ZeroTimeout(t *testing.T) {
	cfg := AsyncConfig{
		Timeout: 0,
	}

	result := cfg.WithDefaults()

	assert.Equal(t, DefaultAsyncConfig(), result)
}

func TestAsyncConfig_WithDefaults_NonZeroTimeout(t *testing.T) {
	cfg := AsyncConfig{
		Enabled:  false,
		Timeout:  5 * time.Second,
		Interval: 100 * time.Millisecond,
		Jitter:   0.1,
		Backoff: BackoffConfig{
			Enabled:     false,
			Factor:      2.0,
			MaxInterval: 2 * time.Second,
		},
	}

	result := cfg.WithDefaults()

	assert.Equal(t, cfg, result)
	assert.False(t, result.Enabled)
	assert.Equal(t, 5*time.Second, result.Timeout)
	assert.Equal(t, 100*time.Millisecond, result.Interval)
}

func TestAsyncConfig_WithDefaults_PreservesCustomValues(t *testing.T) {
	cfg := AsyncConfig{
		Enabled:  true,
		Timeout:  30 * time.Second,
		Interval: 500 * time.Millisecond,
		Jitter:   0.5,
		Backoff: BackoffConfig{
			Enabled:     true,
			Factor:      3.0,
			MaxInterval: 5 * time.Second,
		},
	}

	result := cfg.WithDefaults()

	assert.Equal(t, 30*time.Second, result.Timeout)
	assert.Equal(t, 500*time.Millisecond, result.Interval)
	assert.Equal(t, 0.5, result.Jitter)
	assert.Equal(t, 3.0, result.Backoff.Factor)
	assert.Equal(t, 5*time.Second, result.Backoff.MaxInterval)
}

func TestBackoffConfig_ZeroValues(t *testing.T) {
	cfg := BackoffConfig{}

	assert.False(t, cfg.Enabled)
	assert.Equal(t, 0.0, cfg.Factor)
	assert.Equal(t, time.Duration(0), cfg.MaxInterval)
}

func TestAsyncConfig_ZeroValues(t *testing.T) {
	cfg := AsyncConfig{}

	assert.False(t, cfg.Enabled)
	assert.Equal(t, time.Duration(0), cfg.Timeout)
	assert.Equal(t, time.Duration(0), cfg.Interval)
	assert.Equal(t, 0.0, cfg.Jitter)
}
