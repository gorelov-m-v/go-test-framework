package client

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

type Client struct {
	rdb         *redis.Client
	addr        string
	AsyncConfig config.AsyncConfig
}

type Config struct {
	Addr        string             `mapstructure:"addr" yaml:"addr" json:"addr"`
	Password    string             `mapstructure:"password" yaml:"password" json:"password"`
	DB          int                `mapstructure:"db" yaml:"db" json:"db"`
	AsyncConfig config.AsyncConfig `mapstructure:"asyncConfig" yaml:"asyncConfig" json:"asyncConfig"`
}

func New(cfg Config) (*Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("Redis address is required")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	asyncCfg := cfg.AsyncConfig
	if asyncCfg.Timeout == 0 {
		asyncCfg = config.DefaultAsyncConfig()
	}

	return &Client{
		rdb:         rdb,
		addr:        cfg.Addr,
		AsyncConfig: asyncCfg,
	}, nil
}

func (c *Client) Addr() string {
	return c.addr
}

func (c *Client) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) *Result {
	start := time.Now()

	val, err := c.rdb.Get(ctx, key).Result()
	duration := time.Since(start)

	result := &Result{
		Key:      key,
		Duration: duration,
	}

	if err == redis.Nil {
		result.Exists = false
		result.Error = nil
	} else if err != nil {
		result.Exists = false
		result.Error = err
	} else {
		result.Value = val
		result.Exists = true
	}

	return result
}

func (c *Client) Exists(ctx context.Context, key string) *Result {
	start := time.Now()

	count, err := c.rdb.Exists(ctx, key).Result()
	duration := time.Since(start)

	result := &Result{
		Key:      key,
		Duration: duration,
		Exists:   count > 0,
		Error:    err,
	}

	return result
}

func (c *Client) TTL(ctx context.Context, key string) *Result {
	start := time.Now()

	ttl, err := c.rdb.TTL(ctx, key).Result()
	duration := time.Since(start)

	result := &Result{
		Key:      key,
		Duration: duration,
		Error:    err,
	}

	if ttl == -2 {
		result.Exists = false
		result.TTL = 0
	} else if ttl == -1 {
		result.Exists = true
		result.TTL = -1
	} else {
		result.Exists = true
		result.TTL = ttl
	}

	return result
}

func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *Client) RDB() *redis.Client {
	return c.rdb
}
