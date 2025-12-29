package config

import (
	"os"
	"testing"
	"time"

	"go-test-framework/pkg/httpclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock client wrapper for testing
type MockClient struct {
	HTTP *httpclient.Client
}

type TestEnv struct {
	Service1 *MockClient `config:"service1"`
	Service2 *MockClient `config:"service2"`
}

func TestBuildEnv_Success(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
  defaultHeaders:
    Authorization: Bearer token1

service2:
  baseURL: https://service2.com
  timeout: 20s
  defaultHeaders:
    Accept: application/json
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	env := &TestEnv{}
	err = BuildEnv(env)
	require.NoError(t, err)

	// Verify Service1
	require.NotNil(t, env.Service1)
	require.NotNil(t, env.Service1.HTTP)
	assert.Equal(t, "https://service1.com", env.Service1.HTTP.BaseURL)
	assert.Equal(t, 10*time.Second, env.Service1.HTTP.HTTPClient.Timeout)
	// Viper converts map keys to lowercase
	assert.Equal(t, "Bearer token1", env.Service1.HTTP.DefaultHeaders["authorization"])

	// Verify Service2
	require.NotNil(t, env.Service2)
	require.NotNil(t, env.Service2.HTTP)
	assert.Equal(t, "https://service2.com", env.Service2.HTTP.BaseURL)
	assert.Equal(t, 20*time.Second, env.Service2.HTTP.HTTPClient.Timeout)
	// Viper converts map keys to lowercase
	assert.Equal(t, "application/json", env.Service2.HTTP.DefaultHeaders["accept"])
}

func TestBuildEnv_NotAPointer(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	env := TestEnv{}
	err = BuildEnv(env) // Passing struct instead of pointer
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects a pointer to struct")
}

func TestBuildEnv_ConfigKeyNotFound(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	type EnvWithMissingKey struct {
		Service *MockClient `config:"nonExistentService"`
	}

	env := &EnvWithMissingKey{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config key 'nonExistentService'")
	assert.Contains(t, err.Error(), "not found")
}

func TestBuildEnv_FieldNotExported(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	type EnvWithUnexportedField struct {
		service *MockClient `config:"service1"` // lowercase = unexported
	}

	env := &EnvWithUnexportedField{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exported")
}

func TestBuildEnv_ClientMissingHTTPField(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	type BadClient struct {
		Client *httpclient.Client // Wrong field name
	}

	type EnvWithBadClient struct {
		Service *BadClient `config:"service1"`
	}

	env := &EnvWithBadClient{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have an exported field 'HTTP *httpclient.Client'")
}

func TestBuildEnv_FieldNotPointer(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	type EnvWithNonPointerField struct {
		Service MockClient `config:"service1"` // Not a pointer
	}

	env := &EnvWithNonPointerField{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be a pointer type")
}

func TestBuildEnv_SkipsFieldsWithoutConfigTag(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	type EnvWithMixedFields struct {
		Service1       *MockClient `config:"service1"`
		IgnoredService *MockClient // No config tag
	}

	env := &EnvWithMixedFields{}
	err = BuildEnv(env)
	require.NoError(t, err)

	// Service1 should be initialized
	assert.NotNil(t, env.Service1)
	assert.NotNil(t, env.Service1.HTTP)

	// IgnoredService should remain nil
	assert.Nil(t, env.IgnoredService)
}
