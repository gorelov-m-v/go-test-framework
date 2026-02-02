package builder

import (
	"github.com/spf13/viper"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	dbclient "github.com/gorelov-m-v/go-test-framework/pkg/database/client"
	grpcclient "github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
	httpclient "github.com/gorelov-m-v/go-test-framework/pkg/http/client"
	kafkaclient "github.com/gorelov-m-v/go-test-framework/pkg/kafka/client"
	redisclient "github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

var httpInjector = &ClientInjector[httpclient.Client]{
	TagName:        tagHTTPConfig,
	ClientName:     "HTTP",
	SetterTypeName: "httpclient.HTTPSetter",
	CreateClient: func(v *viper.Viper, configKey string) (*httpclient.Client, error) {
		var svcCfg config.ServiceConfig
		if err := v.UnmarshalKey(configKey, &svcCfg); err != nil {
			return nil, err
		}
		return httpclient.New(httpclient.Config{
			BaseURL:          svcCfg.BaseURL,
			Timeout:          svcCfg.Timeout,
			DefaultHeaders:   svcCfg.DefaultHeaders,
			MaskHeaders:      svcCfg.MaskHeaders,
			ContractSpec:     svcCfg.ContractSpec,
			ContractBasePath: svcCfg.ContractBasePath,
		})
	},
	SetOnTarget: func(target any, client *httpclient.Client) error {
		if s, ok := target.(httpclient.HTTPSetter); ok {
			s.SetHTTP(client)
			return nil
		}
		return errNotSetter
	},
}

var dbInjector = (&ConfigClientInjector[dbclient.Config, dbclient.Client]{
	TagName:        tagDBConfig,
	ClientName:     "Database",
	SetterTypeName: "dbclient.DBSetter",
	NewClient:      dbclient.New,
	SetOnTarget: func(target any, client *dbclient.Client) error {
		if s, ok := target.(dbclient.DBSetter); ok {
			s.SetDB(client)
			return nil
		}
		return errNotSetter
	},
}).ToInjector()

var redisInjector = (&ConfigClientInjector[redisclient.Config, redisclient.Client]{
	TagName:        tagRedisConfig,
	ClientName:     "Redis",
	AsyncKey:       asyncKeyRedis,
	SetterTypeName: "redisclient.RedisSetter",
	NewClient:      redisclient.New,
	SetAsync:       func(c *redisclient.Config, a config.AsyncConfig) { c.AsyncConfig = a },
	SetOnTarget: func(target any, client *redisclient.Client) error {
		if s, ok := target.(redisclient.RedisSetter); ok {
			s.SetRedis(client)
			return nil
		}
		return errNotSetter
	},
}).ToInjector()

var grpcInjector = (&ConfigClientInjector[grpcclient.Config, grpcclient.Client]{
	TagName:        tagGRPCConfig,
	ClientName:     "gRPC",
	AsyncKey:       asyncKeyGRPC,
	SetterTypeName: "grpcclient.GRPCSetter",
	NewClient:      grpcclient.New,
	SetAsync:       func(c *grpcclient.Config, a config.AsyncConfig) { c.AsyncConfig = a },
	SetOnTarget: func(target any, client *grpcclient.Client) error {
		if s, ok := target.(grpcclient.GRPCSetter); ok {
			s.SetGRPC(client)
			return nil
		}
		return errNotSetter
	},
}).ToInjector()

var kafkaInjector = (&ConfigClientInjector[kafkaclient.Config, kafkaclient.Client]{
	TagName:        tagKafkaConfig,
	ClientName:     "Kafka",
	AsyncKey:       asyncKeyKafka,
	SetterTypeName: "kafkaclient.KafkaSetter",
	NewClient:      kafkaclient.New,
	SetAsync:       func(c *kafkaclient.Config, a config.AsyncConfig) { c.AsyncConfig = a },
	SetOnTarget: func(target any, client *kafkaclient.Client) error {
		if s, ok := target.(kafkaclient.KafkaSetter); ok {
			s.SetKafka(client)
			return nil
		}
		return errNotSetter
	},
}).ToInjector()
