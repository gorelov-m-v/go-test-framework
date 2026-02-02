package builder

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

const (
	tagHTTPConfig  = "config"
	tagDBConfig    = "db_config"
	tagAsyncConfig = "async_config"
	tagKafkaConfig = "kafka_config"
	tagGRPCConfig  = "grpc_config"
	tagRedisConfig = "redis_config"
)

const (
	asyncKeyHTTP  = "http_dsl.async"
	asyncKeyDB    = "db_dsl.async"
	asyncKeyKafka = "kafka_dsl.async"
	asyncKeyGRPC  = "grpc_dsl.async"
	asyncKeyRedis = "redis_dsl.async"
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

		if configKey := field.Tag.Get(tagHTTPConfig); configKey != "" {
			if err := httpInjector.Inject(v, fieldValue, field, configKey, structName); err != nil {
				return err
			}
			continue
		}

		if configKey := field.Tag.Get(tagDBConfig); configKey != "" {
			if err := dbInjector.Inject(v, fieldValue, field, configKey, structName); err != nil {
				return err
			}
			continue
		}

		if configKey := field.Tag.Get(tagAsyncConfig); configKey != "" {
			if err := injectAsyncConfig(v, fieldValue, field, configKey, structName); err != nil {
				return err
			}
			continue
		}

		if configKey := field.Tag.Get(tagKafkaConfig); configKey != "" {
			if err := kafkaInjector.Inject(v, fieldValue, field, configKey, structName); err != nil {
				return err
			}
			continue
		}

		if configKey := field.Tag.Get(tagGRPCConfig); configKey != "" {
			if err := grpcInjector.Inject(v, fieldValue, field, configKey, structName); err != nil {
				return err
			}
			continue
		}

		if configKey := field.Tag.Get(tagRedisConfig); configKey != "" {
			if err := redisInjector.Inject(v, fieldValue, field, configKey, structName); err != nil {
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
