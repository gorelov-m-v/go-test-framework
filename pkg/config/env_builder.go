package config

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/spf13/viper"

	dbclient "go-test-framework/pkg/database/client"
	"go-test-framework/pkg/http/client"
)

var debugEnabled = os.Getenv("GO_TEST_FRAMEWORK_DEBUG") == "1"

func debugLog(format string, args ...any) {
	if debugEnabled {
		log.Printf("BuildEnv: "+format, args...)
	}
}

type AsyncConfig struct {
	Enabled  bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Timeout  time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	Interval time.Duration `mapstructure:"interval" yaml:"interval" json:"interval"`
	Backoff  BackoffConfig `mapstructure:"backoff" yaml:"backoff" json:"backoff"`
	Jitter   float64       `mapstructure:"jitter" yaml:"jitter" json:"jitter"`
}

type BackoffConfig struct {
	Enabled     bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Factor      float64       `mapstructure:"factor" yaml:"factor" json:"factor"`
	MaxInterval time.Duration `mapstructure:"max_interval" yaml:"max_interval" json:"max_interval"`
}

func DefaultAsyncConfig() AsyncConfig {
	return AsyncConfig{
		Enabled:  true,
		Timeout:  10 * time.Second,
		Interval: 200 * time.Millisecond,
		Backoff: BackoffConfig{
			Enabled:     true,
			Factor:      1.5,
			MaxInterval: 1 * time.Second,
		},
		Jitter: 0.2,
	}
}

func BuildEnv(envPtr any) error {
	v, err := Viper()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	envValue, structName, err := validateAndUnwrapStruct(envPtr)
	if err != nil {
		return err
	}

	debugLog("scanning struct '%s' for configuration tags", structName)

	envType := envValue.Type()
	for i := 0; i < envType.NumField(); i++ {
		field := envType.Field(i)
		fieldValue := envValue.Field(i)

		if configKey := field.Tag.Get("config"); configKey != "" {
			if err := injectHTTPClient(v, fieldValue, field, configKey, structName); err != nil {
				return err
			}
			continue
		}

		if dbConfigKey := field.Tag.Get("db_config"); dbConfigKey != "" {
			if err := injectDBClient(v, fieldValue, field, dbConfigKey, structName); err != nil {
				return err
			}
			continue
		}

		if asyncConfigKey := field.Tag.Get("async_config"); asyncConfigKey != "" {
			if err := injectAsyncConfig(v, fieldValue, field, asyncConfigKey, structName); err != nil {
				return err
			}
			continue
		}
	}

	return nil
}

func validateAndUnwrapStruct(envPtr any) (reflect.Value, string, error) {
	envValue := reflect.ValueOf(envPtr)
	if envValue.Kind() != reflect.Ptr {
		return reflect.Value{}, "", fmt.Errorf("BuildEnv expects a pointer to struct, got %T", envPtr)
	}

	envValue = envValue.Elem()
	if envValue.Kind() != reflect.Struct {
		return reflect.Value{}, "", fmt.Errorf("BuildEnv expects a pointer to struct, got pointer to %s", envValue.Kind())
	}

	return envValue, envValue.Type().Name(), nil
}

func injectHTTPClient(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, configKey, structName string) error {
	debugLog("found tag 'config:%s' on field '%s' (type=%s)", configKey, field.Name, field.Type)

	if err := validateFieldForInjection(fieldValue, field.Name, configKey, structName, "config"); err != nil {
		return err
	}

	if !v.IsSet(configKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": config key '%s' not found", structName, field.Name, configKey, configKey)
	}

	var svcCfg ServiceConfig
	if err := v.UnmarshalKey(configKey, &svcCfg); err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": failed to unmarshal config: %w", structName, field.Name, configKey, err)
	}

	debugLog("injecting config '%s' into field '%s'", configKey, field.Name)

	httpClient := client.New(client.Config{
		BaseURL:        svcCfg.BaseURL,
		Timeout:        svcCfg.Timeout,
		DefaultHeaders: svcCfg.DefaultHeaders,
	})

	if fieldValue.Type() == reflect.TypeOf((*client.Client)(nil)) {
		fieldValue.Set(reflect.ValueOf(httpClient))
		debugLog("injected HTTP client into '%s'", field.Name)
		return nil
	}

	if err := injectHTTPClientIntoWrapper(fieldValue, field, configKey, structName, httpClient); err != nil {
		return err
	}

	debugLog("injected HTTP client into '%s'", field.Name)
	return nil
}

func injectHTTPClientIntoWrapper(fieldValue reflect.Value, field reflect.StructField, configKey, structName string, httpClient *client.Client) error {
	if fieldValue.Type().Elem().Kind() != reflect.Struct {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": wrapper must be pointer to struct, got '%s'", structName, field.Name, configKey, fieldValue.Type())
	}

	wrapperType := fieldValue.Type().Elem()
	wrapperInstance := reflect.New(wrapperType)

	httpField := wrapperInstance.Elem().FieldByName("HTTP")
	if !httpField.IsValid() {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": wrapper type '%s' must have exported field 'HTTP *client.Client'", structName, field.Name, configKey, wrapperType.Name())
	}

	if !httpField.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": wrapper type '%s' field 'HTTP' is not exported", structName, field.Name, configKey, wrapperType.Name())
	}

	if httpField.Type() != reflect.TypeOf((*client.Client)(nil)) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": wrapper field 'HTTP' must be '*client.Client', got '%s'", structName, field.Name, configKey, httpField.Type())
	}

	httpField.Set(reflect.ValueOf(httpClient))
	fieldValue.Set(wrapperInstance)
	return nil
}

func injectDBClient(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, dbConfigKey, structName string) error {
	debugLog("found tag 'db_config:%s' on field '%s' (type=%s)", dbConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag db_config:\"%s\" but is not exported", structName, field.Name, dbConfigKey)
	}

	if fieldValue.Type() != reflect.TypeOf((*dbclient.Client)(nil)) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag db_config:\"%s\": field must be '*dbclient.Client', got '%s'", structName, field.Name, dbConfigKey, fieldValue.Type())
	}

	if !v.IsSet(dbConfigKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag db_config:\"%s\": config key '%s' not found", structName, field.Name, dbConfigKey, dbConfigKey)
	}

	var dbCfg dbclient.Config
	if err := v.UnmarshalKey(dbConfigKey, &dbCfg); err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag db_config:\"%s\": failed to unmarshal config: %w", structName, field.Name, dbConfigKey, err)
	}

	debugLog("injecting config '%s' into field '%s'", dbConfigKey, field.Name)

	dbClient, err := dbclient.New(dbCfg)
	if err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag db_config:\"%s\": failed to create db client: %w", structName, field.Name, dbConfigKey, err)
	}

	fieldValue.Set(reflect.ValueOf(dbClient))
	debugLog("injected DB client into '%s'", field.Name)
	return nil
}

func injectAsyncConfig(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, asyncConfigKey, structName string) error {
	debugLog("found tag 'async_config:%s' on field '%s' (type=%s)", asyncConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag async_config:\"%s\" but is not exported", structName, field.Name, asyncConfigKey)
	}

	if fieldValue.Type() != reflect.TypeOf(AsyncConfig{}) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag async_config:\"%s\": field must be 'config.AsyncConfig', got '%s'", structName, field.Name, asyncConfigKey, fieldValue.Type())
	}

	var asyncCfg AsyncConfig
	if v.IsSet(asyncConfigKey) {
		if err := v.UnmarshalKey(asyncConfigKey, &asyncCfg); err != nil {
			return fmt.Errorf("BuildEnv(%s): field '%s' tag async_config:\"%s\": failed to unmarshal config: %w", structName, field.Name, asyncConfigKey, err)
		}
		debugLog("loaded async config from '%s'", asyncConfigKey)
	} else {
		asyncCfg = DefaultAsyncConfig()
		debugLog("using default async config for field '%s'", field.Name)
	}

	fieldValue.Set(reflect.ValueOf(asyncCfg))
	debugLog("injected async config into '%s'", field.Name)
	return nil
}

func validateFieldForInjection(fieldValue reflect.Value, fieldName, configKey, structName, tagName string) error {
	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag %s:\"%s\" but is not exported", structName, fieldName, tagName, configKey)
	}

	if fieldValue.Kind() != reflect.Ptr {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag %s:\"%s\": field must be a pointer, got %s", structName, fieldName, tagName, configKey, fieldValue.Type())
	}

	return nil
}
