package config

import (
	"os"
	"testing"
	"time"

	"go-test-framework/pkg/http/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	HTTP *client.Client
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

	resetSingleton()

	env := &TestEnv{}
	err = BuildEnv(env)
	require.NoError(t, err)

	require.NotNil(t, env.Service1)
	require.NotNil(t, env.Service1.HTTP)
	assert.Equal(t, "https://service1.com", env.Service1.HTTP.BaseURL)
	assert.Equal(t, 10*time.Second, env.Service1.HTTP.HTTPClient.Timeout)
	assert.Equal(t, "Bearer token1", env.Service1.HTTP.DefaultHeaders["authorization"])

	require.NotNil(t, env.Service2)
	require.NotNil(t, env.Service2.HTTP)
	assert.Equal(t, "https://service2.com", env.Service2.HTTP.BaseURL)
	assert.Equal(t, 20*time.Second, env.Service2.HTTP.HTTPClient.Timeout)
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

	resetSingleton()

	env := TestEnv{}
	err = BuildEnv(env)
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

	resetSingleton()

	type EnvWithUnexportedField struct {
		service *MockClient `config:"service1"`
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

	resetSingleton()

	type BadClient struct {
		Client *client.Client
	}

	type EnvWithBadClient struct {
		Service *BadClient `config:"service1"`
	}

	env := &EnvWithBadClient{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrapper type 'BadClient' must have exported field 'HTTP *client.Client'")
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

	resetSingleton()

	type EnvWithNonPointerField struct {
		Service MockClient `config:"service1"`
	}

	env := &EnvWithNonPointerField{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field must be a pointer")
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

	resetSingleton()

	type EnvWithMixedFields struct {
		Service1       *MockClient `config:"service1"`
		IgnoredService *MockClient
	}

	env := &EnvWithMixedFields{}
	err = BuildEnv(env)
	require.NoError(t, err)

	assert.NotNil(t, env.Service1)
	assert.NotNil(t, env.Service1.HTTP)
	assert.Nil(t, env.IgnoredService)
}

func TestBuildEnv_PointerToNonStruct(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	resetSingleton()

	var notAStruct int
	err = BuildEnv(&notAStruct)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects a pointer to struct")
	assert.Contains(t, err.Error(), "pointer to int")
}

func TestBuildEnv_DirectClientInjection(t *testing.T) {
	configContent := `
service1:
  baseURL: https://direct-service.com
  timeout: 15s
  defaultHeaders:
    X-API-Key: secret123
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	resetSingleton()

	type EnvWithDirectClient struct {
		DirectClient *client.Client `config:"service1"`
	}

	env := &EnvWithDirectClient{}
	err = BuildEnv(env)
	require.NoError(t, err)

	require.NotNil(t, env.DirectClient)
	assert.Equal(t, "https://direct-service.com", env.DirectClient.BaseURL)
	assert.Equal(t, 15*time.Second, env.DirectClient.HTTPClient.Timeout)
	assert.Equal(t, "secret123", env.DirectClient.DefaultHeaders["x-api-key"])
}

func TestBuildEnv_WrapperNotPointerToStruct(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	resetSingleton()

	type EnvWithPointerToInt struct {
		Service *int `config:"service1"`
	}

	env := &EnvWithPointerToInt{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrapper must be pointer to struct")
}

func TestBuildEnv_WrapperHTTPFieldWrongType(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	resetSingleton()

	type BadClientWithWrongHTTPType struct {
		HTTP string
	}

	type EnvWithBadHTTPType struct {
		Service *BadClientWithWrongHTTPType `config:"service1"`
	}

	env := &EnvWithBadHTTPType{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrapper field 'HTTP' must be '*client.Client'")
}

func TestBuildEnv_WrapperHTTPFieldNotExported(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	resetSingleton()

	type BadClientWithUnexportedHTTP struct {
		http *client.Client
	}

	type EnvWithUnexportedHTTP struct {
		Service *BadClientWithUnexportedHTTP `config:"service1"`
	}

	env := &EnvWithUnexportedHTTP{}
	err = BuildEnv(env)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wrapper type 'BadClientWithUnexportedHTTP' must have exported field 'HTTP *client.Client'")
}

func TestBuildEnv_ErrorMessagesContainContext(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	resetSingleton()

	type ContextTestEnv struct {
		Service MockClient `config:"service1"`
	}

	env := &ContextTestEnv{}
	err = BuildEnv(env)
	require.Error(t, err)

	// Error should contain: struct name, field name, tag info
	assert.Contains(t, err.Error(), "ContextTestEnv")
	assert.Contains(t, err.Error(), "Service")
	assert.Contains(t, err.Error(), "config:\"service1\"")
}

func TestBuildEnv_DebugLogging(t *testing.T) {
	configContent := `
service1:
  baseURL: https://service1.com
  timeout: 10s
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	resetSingleton()

	// Enable debug mode
	originalDebug := debugEnabled
	debugEnabled = true
	defer func() { debugEnabled = originalDebug }()

	// Note: This test verifies debug mode is enabled
	// In actual usage, set GO_TEST_FRAMEWORK_DEBUG=1 to see debug logs
	type SimpleEnv struct {
		Service1 *MockClient `config:"service1"`
	}

	simpleEnv := &SimpleEnv{}
	err = BuildEnv(simpleEnv)
	require.NoError(t, err)

	// Verify injection worked (debug logs would appear in output if run with GO_TEST_FRAMEWORK_DEBUG=1)
	require.NotNil(t, simpleEnv.Service1)
	require.NotNil(t, simpleEnv.Service1.HTTP)
}
