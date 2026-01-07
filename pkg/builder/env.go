package builder

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/spf13/viper"

	"go-test-framework/pkg/config"
	dbclient "go-test-framework/pkg/database/client"
	"go-test-framework/pkg/http/client"
)

var debugEnabled = os.Getenv("GO_TEST_FRAMEWORK_DEBUG") == "1"

func debugLog(format string, args ...any) {
	if debugEnabled {
		log.Printf("BuildEnv: "+format, args...)
	}
}

func BuildEnv(envPtr any) error {
	v, err := config.Viper()
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

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag config:\"%s\" but is not exported", structName, field.Name, configKey)
	}

	if !v.IsSet(configKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": config key '%s' not found", structName, field.Name, configKey, configKey)
	}

	var svcCfg config.ServiceConfig
	if err := v.UnmarshalKey(configKey, &svcCfg); err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag config:\"%s\": failed to unmarshal config: %w", structName, field.Name, configKey, err)
	}

	debugLog("injecting config '%s' into field '%s'", configKey, field.Name)

	httpClient := client.New(client.Config{
		BaseURL:        svcCfg.BaseURL,
		Timeout:        svcCfg.Timeout,
		DefaultHeaders: svcCfg.DefaultHeaders,
	})

	target := fieldValue.Addr().Interface()
	setter, ok := target.(client.HTTPSetter)
	if !ok {
		return fmt.Errorf("BuildEnv Error: Field '%s' has tag 'config' but does not implement 'httpclient.HTTPSetter'. Please use a Link struct", field.Name)
	}

	setter.SetHTTP(httpClient)
	debugLog("injected HTTP client into '%s' via SetHTTP", field.Name)
	return nil
}

func injectDBClient(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, dbConfigKey, structName string) error {
	debugLog("found tag 'db_config:%s' on field '%s' (type=%s)", dbConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag db_config:\"%s\" but is not exported", structName, field.Name, dbConfigKey)
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

	target := fieldValue.Addr().Interface()
	setter, ok := target.(dbclient.DBSetter)
	if !ok {
		return fmt.Errorf("BuildEnv Error: Field '%s' has tag 'db_config' but does not implement 'dbclient.DBSetter'. Please use a Link struct", field.Name)
	}

	setter.SetDB(dbClient)
	debugLog("injected DB client into '%s' via SetDB", field.Name)
	return nil
}

func injectAsyncConfig(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, asyncConfigKey, structName string) error {
	debugLog("found tag 'async_config:%s' on field '%s' (type=%s)", asyncConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag async_config:\"%s\" but is not exported", structName, field.Name, asyncConfigKey)
	}

	if fieldValue.Type() != reflect.TypeOf(config.AsyncConfig{}) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag async_config:\"%s\": field must be 'config.AsyncConfig', got '%s'", structName, field.Name, asyncConfigKey, fieldValue.Type())
	}

	var asyncCfg config.AsyncConfig
	if v.IsSet(asyncConfigKey) {
		if err := v.UnmarshalKey(asyncConfigKey, &asyncCfg); err != nil {
			return fmt.Errorf("BuildEnv(%s): field '%s' tag async_config:\"%s\": failed to unmarshal config: %w", structName, field.Name, asyncConfigKey, err)
		}
		debugLog("loaded async config from '%s'", asyncConfigKey)
	} else {
		asyncCfg = config.DefaultAsyncConfig()
		debugLog("using default async config for field '%s'", field.Name)
	}

	fieldValue.Set(reflect.ValueOf(asyncCfg))
	debugLog("injected async config into '%s'", field.Name)
	return nil
}
