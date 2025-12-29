package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViper_LoadsConfigSuccessfully(t *testing.T) {
	// Setup: Create a temporary config file
	configContent := `
capService:
  baseURL: https://api.example.com
  timeout: 30s
  defaultHeaders:
    Accept: application/json
    Content-Type: application/json

testData:
  username: testuser
  password: testpass
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton for testing
	resetSingleton()

	v, err := Viper()
	require.NoError(t, err)
	assert.NotNil(t, v)

	// Verify config values are loaded
	assert.Equal(t, "https://api.example.com", v.GetString("capService.baseURL"))
	assert.Equal(t, 30*time.Second, v.GetDuration("capService.timeout"))
}

func TestUnmarshalByKey_Success(t *testing.T) {
	configContent := `
testData:
  username: admin
  password: secret
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	type TestData struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}

	var data TestData
	err = UnmarshalByKey("testData", &data)
	require.NoError(t, err)
	assert.Equal(t, "admin", data.Username)
	assert.Equal(t, "secret", data.Password)
}

func TestUnmarshalByKey_KeyNotFound(t *testing.T) {
	configContent := `
capService:
  baseURL: https://api.example.com
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	var data map[string]any
	err = UnmarshalByKey("nonExistentKey", &data)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config key 'nonExistentKey' not found")
}

func TestGetServiceConfig_Success(t *testing.T) {
	configContent := `
myService:
  baseURL: https://my-service.com
  timeout: 15s
  defaultHeaders:
    Authorization: Bearer token123
`
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove("config.yaml")

	// Reset singleton
	resetSingleton()

	cfg, err := GetServiceConfig("myService")
	require.NoError(t, err)
	assert.Equal(t, "https://my-service.com", cfg.BaseURL)
	assert.Equal(t, 15*time.Second, cfg.Timeout)
	// Viper converts map keys to lowercase
	assert.Equal(t, "Bearer token123", cfg.DefaultHeaders["authorization"])
}

func TestViper_ConfigFileNotFound(t *testing.T) {
	// Reset singleton
	resetSingleton()

	// Make sure no config.yaml exists
	os.Remove("config.yaml")
	os.RemoveAll("configs")

	v, err := Viper()
	assert.Error(t, err)
	assert.Nil(t, v)
	assert.Contains(t, err.Error(), "failed to read config file")
}
