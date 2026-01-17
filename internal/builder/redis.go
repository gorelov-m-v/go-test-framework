package builder

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	redisclient "github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

func injectRedisClient(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, redisConfigKey, structName string) error {
	debugLog("found tag 'redis_config:%s' on field '%s' (type=%s)", redisConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag redis_config:\"%s\" but is not exported", structName, field.Name, redisConfigKey)
	}

	if !v.IsSet(redisConfigKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag redis_config:\"%s\": config key '%s' not found", structName, field.Name, redisConfigKey, redisConfigKey)
	}

	var redisCfg redisclient.Config
	if err := v.UnmarshalKey(redisConfigKey, &redisCfg); err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag redis_config:\"%s\": failed to unmarshal config: %w", structName, field.Name, redisConfigKey, err)
	}

	var asyncCfg config.AsyncConfig
	asyncKey := "redis_dsl.async"
	if v.IsSet(asyncKey) {
		if err := v.UnmarshalKey(asyncKey, &asyncCfg); err != nil {
			return fmt.Errorf("BuildEnv(%s): field '%s' tag redis_config:\"%s\": failed to unmarshal async config from '%s': %w", structName, field.Name, redisConfigKey, asyncKey, err)
		}
		debugLog("loaded async config from '%s' for Redis", asyncKey)
	} else {
		asyncCfg = config.DefaultAsyncConfig()
		debugLog("using default async config for Redis field '%s'", field.Name)
	}

	redisCfg.AsyncConfig = asyncCfg

	debugLog("injecting config '%s' into field '%s'", redisConfigKey, field.Name)

	redisClient, err := redisclient.New(redisCfg)
	if err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag redis_config:\"%s\": failed to create Redis client: %w", structName, field.Name, redisConfigKey, err)
	}

	target := fieldValue.Addr().Interface()
	setter, ok := target.(redisclient.RedisSetter)
	if !ok {
		return fmt.Errorf("BuildEnv Error: Field '%s' has tag 'redis_config' but does not implement 'redisclient.RedisSetter'. Please use a Link struct", field.Name)
	}

	setter.SetRedis(redisClient)
	debugLog("injected Redis client into '%s' via SetRedis", field.Name)
	return nil
}
