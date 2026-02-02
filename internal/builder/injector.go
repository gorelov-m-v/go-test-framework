package builder

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/viper"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

var errNotSetter = errors.New("target does not implement setter interface")

type ClientInjector[TClient any] struct {
	TagName        string
	ClientName     string
	SetterTypeName string
	CreateClient   func(v *viper.Viper, configKey string) (*TClient, error)
	SetOnTarget    func(target any, client *TClient) error
}

func (i *ClientInjector[TClient]) Inject(
	v *viper.Viper,
	fieldValue reflect.Value,
	field reflect.StructField,
	configKey, structName string,
) error {
	debugLog("found tag '%s:%s' on field '%s' (type=%s)", i.TagName, configKey, field.Name, field.Type)

	if !fieldValue.CanSet() {
		return fmt.Errorf("BuildEnv(%s): field '%s' has tag %s:\"%s\" but is not exported",
			structName, field.Name, i.TagName, configKey)
	}

	if !v.IsSet(configKey) {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag %s:\"%s\": config key '%s' not found",
			structName, field.Name, i.TagName, configKey, configKey)
	}

	client, err := i.CreateClient(v, configKey)
	if err != nil {
		return fmt.Errorf("BuildEnv(%s): field '%s' tag %s:\"%s\": %w",
			structName, field.Name, i.TagName, configKey, err)
	}

	target := fieldValue.Addr().Interface()
	if err := i.SetOnTarget(target, client); err != nil {
		return fmt.Errorf("BuildEnv Error: Field '%s' has tag '%s' but does not implement '%s'. Please use a Link struct",
			field.Name, i.TagName, i.SetterTypeName)
	}

	debugLog("injected %s client into '%s'", i.ClientName, field.Name)
	return nil
}

type ConfigClientInjector[TConfig, TClient any] struct {
	TagName        string
	ClientName     string
	AsyncKey       string
	SetterTypeName string
	NewClient      func(TConfig) (*TClient, error)
	SetAsync       func(*TConfig, config.AsyncConfig)
	SetOnTarget    func(target any, client *TClient) error
}

func (i *ConfigClientInjector[TConfig, TClient]) ToInjector() *ClientInjector[TClient] {
	return &ClientInjector[TClient]{
		TagName:        i.TagName,
		ClientName:     i.ClientName,
		SetterTypeName: i.SetterTypeName,
		CreateClient:   i.createClient,
		SetOnTarget:    i.SetOnTarget,
	}
}

func (i *ConfigClientInjector[TConfig, TClient]) createClient(v *viper.Viper, configKey string) (*TClient, error) {
	var cfg TConfig
	if err := v.UnmarshalKey(configKey, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if i.AsyncKey != "" && i.SetAsync != nil {
		asyncCfg := loadAsyncConfig(v, i.AsyncKey, i.ClientName)
		i.SetAsync(&cfg, asyncCfg)
	}

	debugLog("creating %s client from config '%s'", i.ClientName, configKey)

	client, err := i.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", i.ClientName, err)
	}

	return client, nil
}

func loadAsyncConfig(v *viper.Viper, asyncKey, clientName string) config.AsyncConfig {
	var asyncCfg config.AsyncConfig
	if v.IsSet(asyncKey) {
		if err := v.UnmarshalKey(asyncKey, &asyncCfg); err != nil {
			debugLog("failed to unmarshal async config from '%s': %v, using defaults", asyncKey, err)
			return config.DefaultAsyncConfig()
		}
		debugLog("loaded async config from '%s' for %s", asyncKey, clientName)
	} else {
		asyncCfg = config.DefaultAsyncConfig()
		debugLog("using default async config for %s", clientName)
	}
	return asyncCfg
}
