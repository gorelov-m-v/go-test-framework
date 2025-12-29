package config

import (
	"fmt"
	"reflect"

	"go-test-framework/pkg/httpclient"
)

// BuildEnv initializes a test environment struct by auto-wiring HTTP clients
// based on struct field tags config:"serviceKey"
//
// Requirements:
//   - envPtr must be a pointer to a struct
//   - Struct fields must be pointers to client wrapper types
//   - Client wrapper structs must have an exported field: HTTP *httpclient.Client
//   - Config YAML must contain sections matching the config tag values
//
// Example:
//
//	type TestEnv struct {
//	    CapClient *client.CapClient `config:"capService"`
//	}
//
//	env := &TestEnv{}
//	if err := config.BuildEnv(env); err != nil {
//	    log.Fatal(err)
//	}
func BuildEnv(envPtr any) error {
	v, err := Viper()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate input is a pointer to struct
	envValue := reflect.ValueOf(envPtr)
	if envValue.Kind() != reflect.Ptr {
		return fmt.Errorf("BuildEnv expects a pointer to struct, got %T", envPtr)
	}

	envValue = envValue.Elem()
	if envValue.Kind() != reflect.Struct {
		return fmt.Errorf("BuildEnv expects a pointer to struct, got pointer to %s", envValue.Kind())
	}

	envType := envValue.Type()

	// Iterate over struct fields
	for i := 0; i < envType.NumField(); i++ {
		field := envType.Field(i)
		fieldValue := envValue.Field(i)

		// Skip fields without config tag
		configKey := field.Tag.Get("config")
		if configKey == "" {
			continue
		}

		// Validate field is settable
		if !fieldValue.CanSet() {
			return fmt.Errorf("field '%s' is not exported (cannot set)", field.Name)
		}

		// Validate field is a pointer
		if fieldValue.Kind() != reflect.Ptr {
			return fmt.Errorf("field '%s' must be a pointer type, got %s", field.Name, fieldValue.Kind())
		}

		// Load service config from YAML
		if !v.IsSet(configKey) {
			return fmt.Errorf("config key '%s' (for field '%s') not found in config file", configKey, field.Name)
		}

		var svcCfg ServiceConfig
		if err := v.UnmarshalKey(configKey, &svcCfg); err != nil {
			return fmt.Errorf("failed to unmarshal config key '%s': %w", configKey, err)
		}

		// Create httpclient.Client with loaded config
		httpClient := httpclient.New(httpclient.Config{
			BaseURL:        svcCfg.BaseURL,
			Timeout:        svcCfg.Timeout,
			DefaultHeaders: svcCfg.DefaultHeaders,
		})

		// Create instance of the client wrapper type
		clientType := fieldValue.Type().Elem()
		clientInstance := reflect.New(clientType)

		// Find and inject HTTP field in the wrapper
		httpField := clientInstance.Elem().FieldByName("HTTP")
		if !httpField.IsValid() {
			return fmt.Errorf("client type '%s' must have an exported field 'HTTP *httpclient.Client'", clientType.Name())
		}

		if !httpField.CanSet() {
			return fmt.Errorf("field 'HTTP' in type '%s' is not exported", clientType.Name())
		}

		if httpField.Type() != reflect.TypeOf((*httpclient.Client)(nil)) {
			return fmt.Errorf("field 'HTTP' in type '%s' must be of type *httpclient.Client", clientType.Name())
		}

		// Inject the httpclient
		httpField.Set(reflect.ValueOf(httpClient))

		// Assign the initialized wrapper to the env struct field
		fieldValue.Set(clientInstance)
	}

	return nil
}
