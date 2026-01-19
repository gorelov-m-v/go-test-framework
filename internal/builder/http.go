package builder

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

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
		BaseURL:          svcCfg.BaseURL,
		Timeout:          svcCfg.Timeout,
		DefaultHeaders:   svcCfg.DefaultHeaders,
		MaskHeaders:      svcCfg.MaskHeaders,
		ContractSpec:     svcCfg.ContractSpec,
		ContractBasePath: svcCfg.ContractBasePath,
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
