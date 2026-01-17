package builder

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

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
