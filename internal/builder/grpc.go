package builder

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	grpcclient "github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

func injectGRPCClient(v *viper.Viper, fieldValue reflect.Value, field reflect.StructField, grpcConfigKey, structName string) error {
	debugLog("found tag 'grpc_config:%s' on field '%s' (type=%s)", grpcConfigKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag grpc_config:\"%s\" but is not exported", structName, field.Name, grpcConfigKey)
	}

	if !v.IsSet(grpcConfigKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag grpc_config:\"%s\": config key '%s' not found", structName, field.Name, grpcConfigKey, grpcConfigKey)
	}

	var grpcCfg grpcclient.Config
	if err := v.UnmarshalKey(grpcConfigKey, &grpcCfg); err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag grpc_config:\"%s\": failed to unmarshal config: %w", structName, field.Name, grpcConfigKey, err)
	}

	var asyncCfg config.AsyncConfig
	asyncKey := "grpc_dsl.async"
	if v.IsSet(asyncKey) {
		if err := v.UnmarshalKey(asyncKey, &asyncCfg); err != nil {
			return fmt.Errorf("BuildEnv(%s): field '%s' tag grpc_config:\"%s\": failed to unmarshal async config from '%s': %w", structName, field.Name, grpcConfigKey, asyncKey, err)
		}
		debugLog("loaded async config from '%s' for gRPC", asyncKey)
	} else {
		asyncCfg = config.DefaultAsyncConfig()
		debugLog("using default async config for gRPC field '%s'", field.Name)
	}

	grpcCfg.AsyncConfig = asyncCfg

	debugLog("injecting config '%s' into field '%s'", grpcConfigKey, field.Name)

	grpcClient, err := grpcclient.New(grpcCfg)
	if err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag grpc_config:\"%s\": failed to create gRPC client: %w", structName, field.Name, grpcConfigKey, err)
	}

	target := fieldValue.Addr().Interface()
	setter, ok := target.(grpcclient.GRPCSetter)
	if !ok {
		return fmt.Errorf("BuildEnv Error: Field '%s' has tag 'grpc_config' but does not implement 'grpcclient.GRPCSetter'. Please use a Link struct", field.Name)
	}

	setter.SetGRPC(grpcClient)
	debugLog("injected gRPC client into '%s' via SetGRPC", field.Name)
	return nil
}
