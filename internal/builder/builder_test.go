package builder

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	dbclient "github.com/gorelov-m-v/go-test-framework/pkg/database/client"
	grpcclient "github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
	kafkaclient "github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	redisclient "github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

func TestValidateAndUnwrapStruct_ValidPointerToStruct(t *testing.T) {
	type TestEnv struct {
		Field string
	}
	env := &TestEnv{}

	value, name, err := validateAndUnwrapStruct(env)

	require.NoError(t, err)
	assert.Equal(t, "TestEnv", name)
	assert.Equal(t, reflect.Struct, value.Kind())
}

func TestValidateAndUnwrapStruct_NonPointer(t *testing.T) {
	type TestEnv struct {
		Field string
	}
	env := TestEnv{}

	_, _, err := validateAndUnwrapStruct(env)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects a pointer to struct")
	assert.Contains(t, err.Error(), "got builder.TestEnv")
}

func TestValidateAndUnwrapStruct_NilPointer(t *testing.T) {
	var env *struct{ Field string }

	_, _, err := validateAndUnwrapStruct(env)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects a pointer to struct")
}

func TestValidateAndUnwrapStruct_PointerToNonStruct(t *testing.T) {
	str := "not a struct"

	_, _, err := validateAndUnwrapStruct(&str)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects a pointer to struct")
	assert.Contains(t, err.Error(), "got pointer to string")
}

func TestValidateAndUnwrapStruct_PointerToInt(t *testing.T) {
	num := 42

	_, _, err := validateAndUnwrapStruct(&num)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "got pointer to int")
}

func TestValidateAndUnwrapStruct_PointerToSlice(t *testing.T) {
	slice := []string{"a", "b"}

	_, _, err := validateAndUnwrapStruct(&slice)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "got pointer to slice")
}

func TestValidateAndUnwrapStruct_PointerToMap(t *testing.T) {
	m := map[string]string{"key": "value"}

	_, _, err := validateAndUnwrapStruct(&m)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "got pointer to map")
}

func TestValidateAndUnwrapStruct_EmptyStruct(t *testing.T) {
	type Empty struct{}
	env := &Empty{}

	value, name, err := validateAndUnwrapStruct(env)

	require.NoError(t, err)
	assert.Equal(t, "Empty", name)
	assert.Equal(t, 0, value.NumField())
}

func TestValidateAndUnwrapStruct_AnonymousStruct(t *testing.T) {
	env := &struct {
		Field1 string
		Field2 int
	}{}

	value, name, err := validateAndUnwrapStruct(env)

	require.NoError(t, err)
	assert.Equal(t, "", name)
	assert.Equal(t, 2, value.NumField())
}

func TestValidateAndUnwrapStruct_NestedStruct(t *testing.T) {
	type Inner struct {
		Value string
	}
	type Outer struct {
		Inner Inner
	}
	env := &Outer{}

	value, name, err := validateAndUnwrapStruct(env)

	require.NoError(t, err)
	assert.Equal(t, "Outer", name)
	assert.Equal(t, 1, value.NumField())
}

type mockHTTPLink struct {
	client *client.Client
}

func (m *mockHTTPLink) SetHTTP(c *client.Client) {
	m.client = c
}

func newTestViper(cfg map[string]interface{}) *viper.Viper {
	v := viper.New()
	for key, value := range cfg {
		v.Set(key, value)
	}
	return v
}

func TestInjectHTTPClient_Success(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"http.testService.baseURL": "https://api.example.com",
		"http.testService.timeout": "30s",
	})

	type TestEnv struct {
		Service mockHTTPLink `config:"http.testService"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectHTTPClient(v, fieldValue, field, "http.testService", "TestEnv")

	require.NoError(t, err)
	assert.NotNil(t, env.Service.client)
}

func TestInjectHTTPClient_ConfigKeyNotFound(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		Service mockHTTPLink `config:"http.missing"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectHTTPClient(v, fieldValue, field, "http.missing", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config key 'http.missing' not found")
}

func TestInjectHTTPClient_UnexportedField(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"http.testService.baseURL": "https://api.example.com",
	})

	type TestEnv struct {
		service mockHTTPLink `config:"http.testService"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectHTTPClient(v, fieldValue, field, "http.testService", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exported")
}

type notHTTPSetter struct {
	Value string
}

func TestInjectHTTPClient_NotImplementsHTTPSetter(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"http.testService.baseURL": "https://api.example.com",
	})

	type TestEnv struct {
		Service notHTTPSetter `config:"http.testService"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectHTTPClient(v, fieldValue, field, "http.testService", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement")
	assert.Contains(t, err.Error(), "HTTPSetter")
}

func TestInjectAsyncConfig_Success(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"custom_async.enabled":  true,
		"custom_async.timeout":  "5s",
		"custom_async.interval": "100ms",
	})

	type TestEnv struct {
		Async config.AsyncConfig `async_config:"custom_async"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectAsyncConfig(v, fieldValue, field, "custom_async", "TestEnv")

	require.NoError(t, err)
	assert.True(t, env.Async.Enabled)
}

func TestInjectAsyncConfig_DefaultWhenNotSet(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		Async config.AsyncConfig `async_config:"missing_async"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectAsyncConfig(v, fieldValue, field, "missing_async", "TestEnv")

	require.NoError(t, err)
	defaultCfg := config.DefaultAsyncConfig()
	assert.Equal(t, defaultCfg.Enabled, env.Async.Enabled)
	assert.Equal(t, defaultCfg.Timeout, env.Async.Timeout)
	assert.Equal(t, defaultCfg.Interval, env.Async.Interval)
}

func TestInjectAsyncConfig_WrongFieldType(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		Async string `async_config:"custom_async"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectAsyncConfig(v, fieldValue, field, "custom_async", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be 'config.AsyncConfig'")
	assert.Contains(t, err.Error(), "got 'string'")
}

func TestInjectAsyncConfig_UnexportedField(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		async config.AsyncConfig `async_config:"custom_async"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectAsyncConfig(v, fieldValue, field, "custom_async", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exported")
}

func TestDebugLog_DoesNotPanicWhenDisabled(t *testing.T) {
	assert.NotPanics(t, func() {
		debugLog("test message: %s", "value")
	})
}

func TestTagConstants(t *testing.T) {
	assert.Equal(t, "config", tagHTTPConfig)
	assert.Equal(t, "db_config", tagDBConfig)
	assert.Equal(t, "async_config", tagAsyncConfig)
	assert.Equal(t, "kafka_config", tagKafkaConfig)
	assert.Equal(t, "grpc_config", tagGRPCConfig)
	assert.Equal(t, "redis_config", tagRedisConfig)
}

func TestAsyncKeyConstants(t *testing.T) {
	assert.Equal(t, "http_dsl.async", asyncKeyHTTP)
	assert.Equal(t, "db_dsl.async", asyncKeyDB)
	assert.Equal(t, "kafka_dsl.async", asyncKeyKafka)
	assert.Equal(t, "grpc_dsl.async", asyncKeyGRPC)
	assert.Equal(t, "redis_dsl.async", asyncKeyRedis)
}

func TestFieldCanSet_ExportedVsUnexported(t *testing.T) {
	type Mixed struct {
		Exported   string
		unexported string
	}
	env := &Mixed{}

	envValue := reflect.ValueOf(env).Elem()

	assert.True(t, envValue.Field(0).CanSet(), "Exported field should be settable")
	assert.False(t, envValue.Field(1).CanSet(), "unexported field should not be settable")
}

func TestMultipleHTTPClients(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"http.service1.baseURL": "https://api1.example.com",
		"http.service2.baseURL": "https://api2.example.com",
	})

	type TestEnv struct {
		Service1 mockHTTPLink `config:"http.service1"`
		Service2 mockHTTPLink `config:"http.service2"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	envType := envValue.Type()

	for i := 0; i < envType.NumField(); i++ {
		field := envType.Field(i)
		fieldValue := envValue.Field(i)
		configKey := field.Tag.Get("config")

		err := injectHTTPClient(v, fieldValue, field, configKey, "TestEnv")
		require.NoError(t, err, "Failed to inject field %s", field.Name)
	}

	assert.NotNil(t, env.Service1.client)
	assert.NotNil(t, env.Service2.client)
	assert.NotSame(t, env.Service1.client, env.Service2.client)
}

type mockDBLink struct {
	client *dbclient.Client
}

func (m *mockDBLink) SetDB(c *dbclient.Client) {
	m.client = c
}

type notDBSetter struct {
	Value string
}

func TestInjectDBClient_ConfigKeyNotFound(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		DB mockDBLink `db_config:"database.missing"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectDBClient(v, fieldValue, field, "database.missing", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config key 'database.missing' not found")
}

func TestInjectDBClient_UnexportedField(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"database.test.driver": "postgres",
		"database.test.dsn":    "postgres://localhost/test",
	})

	type TestEnv struct {
		db mockDBLink `db_config:"database.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectDBClient(v, fieldValue, field, "database.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exported")
}

func TestInjectDBClient_InvalidConfig(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"database.test.driver": "postgres",
		"database.test.dsn":    "postgres://localhost/test",
	})

	type TestEnv struct {
		DB mockDBLink `db_config:"database.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectDBClient(v, fieldValue, field, "database.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create db client")
}

type mockKafkaLink struct {
	client *kafkaclient.Client
}

func (m *mockKafkaLink) SetKafka(c *kafkaclient.Client) {
	m.client = c
}

type notKafkaSetter struct {
	Value string
}

func TestInjectKafkaClient_ConfigKeyNotFound(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		Kafka mockKafkaLink `kafka_config:"kafka.missing"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectKafkaClient(v, fieldValue, field, "kafka.missing", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config key 'kafka.missing' not found")
}

func TestInjectKafkaClient_UnexportedField(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"kafka.test.bootstrapServers": []string{"localhost:9092"},
	})

	type TestEnv struct {
		kafka mockKafkaLink `kafka_config:"kafka.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectKafkaClient(v, fieldValue, field, "kafka.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exported")
}

func TestInjectKafkaClient_MissingTopics(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"kafka.test.bootstrapServers": []string{"localhost:9092"},
		"kafka.test.groupId":          "test-group",
	})

	type TestEnv struct {
		Kafka mockKafkaLink `kafka_config:"kafka.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectKafkaClient(v, fieldValue, field, "kafka.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create kafka client")
}

type mockGRPCLink struct {
	client *grpcclient.Client
}

func (m *mockGRPCLink) SetGRPC(c *grpcclient.Client) {
	m.client = c
}

type notGRPCSetter struct {
	Value string
}

func TestInjectGRPCClient_ConfigKeyNotFound(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		GRPC mockGRPCLink `grpc_config:"grpc.missing"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectGRPCClient(v, fieldValue, field, "grpc.missing", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config key 'grpc.missing' not found")
}

func TestInjectGRPCClient_UnexportedField(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"grpc.test.target":   "localhost:9090",
		"grpc.test.insecure": true,
	})

	type TestEnv struct {
		grpc mockGRPCLink `grpc_config:"grpc.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectGRPCClient(v, fieldValue, field, "grpc.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exported")
}

func TestInjectGRPCClient_NotImplementsGRPCSetter(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"grpc.test.target":   "localhost:9090",
		"grpc.test.insecure": true,
	})

	type TestEnv struct {
		GRPC notGRPCSetter `grpc_config:"grpc.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectGRPCClient(v, fieldValue, field, "grpc.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not implement")
	assert.Contains(t, err.Error(), "GRPCSetter")
}

func TestInjectGRPCClient_Success(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"grpc.test.target":   "localhost:9090",
		"grpc.test.insecure": true,
	})

	type TestEnv struct {
		GRPC mockGRPCLink `grpc_config:"grpc.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectGRPCClient(v, fieldValue, field, "grpc.test", "TestEnv")

	require.NoError(t, err)
	assert.NotNil(t, env.GRPC.client)
}

func TestInjectGRPCClient_WithAsyncConfig(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"grpc.test.target":        "localhost:9090",
		"grpc.test.insecure":      true,
		"grpc_dsl.async.enabled":  true,
		"grpc_dsl.async.timeout":  "5s",
		"grpc_dsl.async.interval": "100ms",
	})

	type TestEnv struct {
		GRPC mockGRPCLink `grpc_config:"grpc.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectGRPCClient(v, fieldValue, field, "grpc.test", "TestEnv")

	require.NoError(t, err)
	assert.NotNil(t, env.GRPC.client)
}

type mockRedisLink struct {
	client *redisclient.Client
}

func (m *mockRedisLink) SetRedis(c *redisclient.Client) {
	m.client = c
}

type notRedisSetter struct {
	Value string
}

func TestInjectRedisClient_ConfigKeyNotFound(t *testing.T) {
	v := newTestViper(map[string]interface{}{})

	type TestEnv struct {
		Redis mockRedisLink `redis_config:"redis.missing"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectRedisClient(v, fieldValue, field, "redis.missing", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config key 'redis.missing' not found")
}

func TestInjectRedisClient_UnexportedField(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"redis.test.addr": "localhost:6379",
	})

	type TestEnv struct {
		redis mockRedisLink `redis_config:"redis.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectRedisClient(v, fieldValue, field, "redis.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not exported")
}

func TestInjectRedisClient_ConnectionError(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"redis.test.addr": "localhost:6379",
	})

	type TestEnv struct {
		Redis mockRedisLink `redis_config:"redis.test"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	field := envValue.Type().Field(0)
	fieldValue := envValue.Field(0)

	err := injectRedisClient(v, fieldValue, field, "redis.test", "TestEnv")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Redis client")
}

func TestMixedEnvStruct_HTTPAndGRPC(t *testing.T) {
	v := newTestViper(map[string]interface{}{
		"http.api.baseURL":      "https://api.example.com",
		"grpc.service.target":   "localhost:9090",
		"grpc.service.insecure": true,
	})

	type TestEnv struct {
		API     mockHTTPLink       `config:"http.api"`
		Service mockGRPCLink       `grpc_config:"grpc.service"`
		Async   config.AsyncConfig `async_config:"custom_async"`
	}
	env := &TestEnv{}

	envValue := reflect.ValueOf(env).Elem()
	envType := envValue.Type()

	for i := 0; i < envType.NumField(); i++ {
		field := envType.Field(i)
		fieldValue := envValue.Field(i)

		if configKey := field.Tag.Get(tagHTTPConfig); configKey != "" {
			err := injectHTTPClient(v, fieldValue, field, configKey, "TestEnv")
			require.NoError(t, err)
			continue
		}

		if grpcKey := field.Tag.Get(tagGRPCConfig); grpcKey != "" {
			err := injectGRPCClient(v, fieldValue, field, grpcKey, "TestEnv")
			require.NoError(t, err)
			continue
		}

		if asyncKey := field.Tag.Get(tagAsyncConfig); asyncKey != "" {
			err := injectAsyncConfig(v, fieldValue, field, asyncKey, "TestEnv")
			require.NoError(t, err)
			continue
		}
	}

	assert.NotNil(t, env.API.client)
	assert.NotNil(t, env.Service.client)
	assert.True(t, env.Async.Enabled)
}
