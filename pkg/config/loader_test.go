package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindConfigPaths_ReturnsDefaultWhenNoConfigsDir(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	paths := findConfigPaths()

	assert.Equal(t, []string{"./configs", "../configs"}, paths)
}

func TestFindConfigPaths_FindsConfigsDir(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	tmpDir := t.TempDir()
	configsDir := filepath.Join(tmpDir, "configs")
	err = os.Mkdir(configsDir, 0755)
	require.NoError(t, err)

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	paths := findConfigPaths()

	require.Len(t, paths, 1)
	assert.Equal(t, configsDir, paths[0])
}

func TestFindConfigPaths_FindsConfigsDirInParent(t *testing.T) {
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	tmpDir := t.TempDir()
	configsDir := filepath.Join(tmpDir, "configs")
	err = os.Mkdir(configsDir, 0755)
	require.NoError(t, err)

	subDir := filepath.Join(tmpDir, "sub", "deep")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	err = os.Chdir(subDir)
	require.NoError(t, err)

	paths := findConfigPaths()

	require.Len(t, paths, 1)
	assert.Equal(t, configsDir, paths[0])
}

func TestConfigureAllure_SkipsWhenEnvAlreadySet(t *testing.T) {
	originalEnv := os.Getenv("ALLURE_OUTPUT_PATH")
	defer os.Setenv("ALLURE_OUTPUT_PATH", originalEnv)

	os.Setenv("ALLURE_OUTPUT_PATH", "/existing/path")

	v := viper.New()
	v.Set("allure.outputPath", "/new/path")

	configureAllure(v)

	assert.Equal(t, "/existing/path", os.Getenv("ALLURE_OUTPUT_PATH"))
}

func TestConfigureAllure_SkipsWhenOutputPathEmpty(t *testing.T) {
	originalEnv := os.Getenv("ALLURE_OUTPUT_PATH")
	defer os.Setenv("ALLURE_OUTPUT_PATH", originalEnv)

	os.Unsetenv("ALLURE_OUTPUT_PATH")

	v := viper.New()

	configureAllure(v)

	assert.Empty(t, os.Getenv("ALLURE_OUTPUT_PATH"))
}

func TestConfigureAllure_SetsAbsolutePath(t *testing.T) {
	originalEnv := os.Getenv("ALLURE_OUTPUT_PATH")
	defer os.Setenv("ALLURE_OUTPUT_PATH", originalEnv)

	os.Unsetenv("ALLURE_OUTPUT_PATH")

	tmpDir := t.TempDir()
	absolutePath := filepath.Join(tmpDir, "absolute", "path")
	configFile := filepath.Join(tmpDir, "configs", "config.yaml")
	err := os.MkdirAll(filepath.Dir(configFile), 0755)
	require.NoError(t, err)
	configContent := "allure:\n  outputPath: " + absolutePath + "\n"
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	v := viper.New()
	v.SetConfigFile(configFile)
	err = v.ReadInConfig()
	require.NoError(t, err)

	configureAllure(v)

	assert.Equal(t, absolutePath, os.Getenv("ALLURE_OUTPUT_PATH"))
}

func TestConfigureAllure_ConvertsRelativePath(t *testing.T) {
	originalEnv := os.Getenv("ALLURE_OUTPUT_PATH")
	defer os.Setenv("ALLURE_OUTPUT_PATH", originalEnv)

	os.Unsetenv("ALLURE_OUTPUT_PATH")

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	configFile := filepath.Join(configDir, "config.yaml")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(configFile, []byte("allure:\n  outputPath: tests/allure-results\n"), 0644)
	require.NoError(t, err)

	v := viper.New()
	v.SetConfigFile(configFile)
	err = v.ReadInConfig()
	require.NoError(t, err)

	configureAllure(v)

	expected := filepath.Clean(filepath.Join(tmpDir, "tests"))
	assert.Equal(t, expected, os.Getenv("ALLURE_OUTPUT_PATH"))
}

func TestConfigureAllure_StripsAllureResultsSuffix(t *testing.T) {
	originalEnv := os.Getenv("ALLURE_OUTPUT_PATH")
	defer os.Setenv("ALLURE_OUTPUT_PATH", originalEnv)

	os.Unsetenv("ALLURE_OUTPUT_PATH")

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "configs")
	configFile := filepath.Join(configDir, "config.yaml")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(configFile, []byte("allure:\n  outputPath: output/allure-results\n"), 0644)
	require.NoError(t, err)

	v := viper.New()
	v.SetConfigFile(configFile)
	err = v.ReadInConfig()
	require.NoError(t, err)

	configureAllure(v)

	expected := filepath.Clean(filepath.Join(tmpDir, "output"))
	assert.Equal(t, expected, os.Getenv("ALLURE_OUTPUT_PATH"))
}

func TestServiceConfig_ZeroValues(t *testing.T) {
	cfg := ServiceConfig{}

	assert.Empty(t, cfg.BaseURL)
	assert.Zero(t, cfg.Timeout)
	assert.Nil(t, cfg.DefaultHeaders)
	assert.Empty(t, cfg.MaskHeaders)
	assert.Empty(t, cfg.ContractSpec)
	assert.Empty(t, cfg.ContractBasePath)
}

func TestServiceConfig_WithValues(t *testing.T) {
	cfg := ServiceConfig{
		BaseURL:          "https://api.example.com",
		Timeout:          30000000000,
		DefaultHeaders:   map[string]string{"Authorization": "Bearer token"},
		MaskHeaders:      "Authorization,Cookie",
		ContractSpec:     "openapi/spec.yaml",
		ContractBasePath: "/api/v1",
	}

	assert.Equal(t, "https://api.example.com", cfg.BaseURL)
	assert.Equal(t, "Authorization,Cookie", cfg.MaskHeaders)
	assert.Equal(t, "Bearer token", cfg.DefaultHeaders["Authorization"])
}

func TestAllureConfig_ZeroValues(t *testing.T) {
	cfg := AllureConfig{}

	assert.Empty(t, cfg.OutputPath)
}

func TestAllureConfig_WithValues(t *testing.T) {
	cfg := AllureConfig{
		OutputPath: "/path/to/allure",
	}

	assert.Equal(t, "/path/to/allure", cfg.OutputPath)
}
