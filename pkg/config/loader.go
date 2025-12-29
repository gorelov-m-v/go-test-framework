package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var (
	once           sync.Once
	configInstance *viper.Viper
	loadErr        error
)

// resetSingleton is used for testing purposes only
func resetSingleton() {
	once = sync.Once{}
	configInstance = nil
	loadErr = nil
}

// ServiceConfig represents standard configuration for HTTP service client
type ServiceConfig struct {
	BaseURL        string            `mapstructure:"baseURL"`
	Timeout        time.Duration     `mapstructure:"timeout"`
	DefaultHeaders map[string]string `mapstructure:"defaultHeaders"`
}

// Viper returns the singleton viper instance with loaded configuration
// It searches for config.yaml in ./configs/ and project root
// Supports environment variable overrides (e.g., CAPSERVICE_BASEURL)
func Viper() (*viper.Viper, error) {
	once.Do(func() {
		v := viper.New()
		v.SetConfigName("config")
		v.SetConfigType("yaml")

		// Search paths
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")

		// Enable environment variable overrides
		v.AutomaticEnv()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		if err := v.ReadInConfig(); err != nil {
			loadErr = fmt.Errorf("failed to read config file: %w", err)
			return
		}

		configInstance = v
	})

	return configInstance, loadErr
}

// UnmarshalByKey loads a specific section from config into the provided struct
// Example: UnmarshalByKey("testData", &myData)
func UnmarshalByKey(key string, out any) error {
	v, err := Viper()
	if err != nil {
		return err
	}

	if !v.IsSet(key) {
		return fmt.Errorf("config key '%s' not found", key)
	}

	return v.UnmarshalKey(key, out)
}

// GetServiceConfig retrieves and unmarshals a service configuration by key
func GetServiceConfig(key string) (*ServiceConfig, error) {
	var cfg ServiceConfig
	if err := UnmarshalByKey(key, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
