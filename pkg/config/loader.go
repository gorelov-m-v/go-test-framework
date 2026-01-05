package config

import (
	"fmt"
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
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")

		if err := v.ReadInConfig(); err != nil {
			loadErr = fmt.Errorf("failed to read config file: %w", err)
			return
		}

		configInstance = v
	})

	return configInstance, loadErr
}

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
