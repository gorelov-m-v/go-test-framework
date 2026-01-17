package builder

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	kafkaclient "github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/kafka/types"
)

func injectKafkaClient(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, kafkaConfigKey, structName string) error {
	debugLog("found tag 'kafka_config:%s' on field '%s' (type=%s)", kafkaConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag kafka_config:\"%s\" but is not exported", structName, field.Name, kafkaConfigKey)
	}

	if !v.IsSet(kafkaConfigKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag kafka_config:\"%s\": config key '%s' not found", structName, field.Name, kafkaConfigKey, kafkaConfigKey)
	}

	var kafkaCfg types.Config
	if err := v.UnmarshalKey(kafkaConfigKey, &kafkaCfg); err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag kafka_config:\"%s\": failed to unmarshal config: %w", structName, field.Name, kafkaConfigKey, err)
	}

	var asyncCfg config.AsyncConfig
	asyncKey := "kafka_dsl.async"
	if v.IsSet(asyncKey) {
		if err := v.UnmarshalKey(asyncKey, &asyncCfg); err != nil {
			return fmt.Errorf("BuildEnv(%s): field '%s' tag kafka_config:\"%s\": failed to unmarshal async config from '%s': %w", structName, field.Name, kafkaConfigKey, asyncKey, err)
		}
		debugLog("loaded async config from '%s' for Kafka", asyncKey)
	} else {
		asyncCfg = config.DefaultAsyncConfig()
		debugLog("using default async config for Kafka field '%s'", field.Name)
	}

	debugLog("injecting config '%s' into field '%s'", kafkaConfigKey, field.Name)

	kafkaClient, err := kafkaclient.New(kafkaCfg, asyncCfg)
	if err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag kafka_config:\"%s\": failed to create kafka client: %w", structName, field.Name, kafkaConfigKey, err)
	}

	target := fieldValue.Addr().Interface()
	setter, ok := target.(kafkaclient.KafkaSetter)
	if !ok {
		return fmt.Errorf("BuildEnv Error: Field '%s' has tag 'kafka_config' but does not implement 'kafkaclient.KafkaSetter'. Please use a Link struct", field.Name)
	}

	setter.SetKafka(kafkaClient)
	debugLog("injected Kafka client into '%s' via SetKafka", field.Name)
	return nil
}
