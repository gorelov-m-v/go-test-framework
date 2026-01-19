package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var (
	once           sync.Once
	configInstance *viper.Viper
	loadErr        error
)

type ServiceConfig struct {
	BaseURL          string            `mapstructure:"baseURL"`
	Timeout          time.Duration     `mapstructure:"timeout"`
	DefaultHeaders   map[string]string `mapstructure:"defaultHeaders"`
	MaskHeaders      string            `mapstructure:"maskHeaders"`
	ContractSpec     string            `mapstructure:"contractSpec"`
	ContractBasePath string            `mapstructure:"contractBasePath"`
}

type AllureConfig struct {
	OutputPath string `mapstructure:"outputPath"`
}

func Viper() (*viper.Viper, error) {
	once.Do(func() {
		v := viper.New()

		env := os.Getenv("ENV")
		if env == "" {
			env = "local"
		}

		configName := fmt.Sprintf("config.%s", env)

		v.SetConfigName(configName)
		v.SetConfigType("yaml")

		for _, path := range findConfigPaths() {
			v.AddConfigPath(path)
		}

		log.Printf("[Config] Loading configuration for env: '%s' (file: %s.yaml)", env, configName)

		if err := v.ReadInConfig(); err != nil {
			loadErr = fmt.Errorf("failed to read config file '%s.yaml': %w", configName, err)
			return
		}

		configInstance = v

		configureAllure(v)
	})

	return configInstance, loadErr
}

func findConfigPaths() []string {
	var paths []string

	cwd, err := os.Getwd()
	if err != nil {
		return []string{"./configs", "../configs"}
	}

	dir := cwd
	for i := 0; i < 10; i++ {
		configDir := filepath.Join(dir, "configs")
		if info, err := os.Stat(configDir); err == nil && info.IsDir() {
			paths = append(paths, configDir)
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if len(paths) == 0 {
		paths = []string{"./configs", "../configs"}
	}

	return paths
}

func configureAllure(v *viper.Viper) {
	if os.Getenv("ALLURE_OUTPUT_PATH") != "" {
		return
	}

	outputPath := v.GetString("allure.outputPath")
	if outputPath == "" {
		return
	}

	if !filepath.IsAbs(outputPath) {
		configDir := filepath.Dir(v.ConfigFileUsed())
		outputPath = filepath.Join(configDir, "..", outputPath)
	}

	outputPath = filepath.Clean(outputPath)
	if filepath.Base(outputPath) == "allure-results" {
		outputPath = filepath.Dir(outputPath)
	}

	os.Setenv("ALLURE_OUTPUT_PATH", outputPath)
}
