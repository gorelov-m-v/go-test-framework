package config

import (
	"fmt"
	"log"
	"os"
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
	BaseURL        string            `mapstructure:"baseURL"`
	Timeout        time.Duration     `mapstructure:"timeout"`
	DefaultHeaders map[string]string `mapstructure:"defaultHeaders"`
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
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath(".")

		log.Printf("[Config] Loading configuration for env: '%s' (file: %s.yaml)", env, configName)

		if err := v.ReadInConfig(); err != nil {
			loadErr = fmt.Errorf("failed to read config file '%s.yaml': %w", configName, err)
			return
		}

		configInstance = v
	})

	return configInstance, loadErr
}
