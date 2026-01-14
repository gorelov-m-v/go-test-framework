package client

// RedisSetter is the interface that Link structs must implement
// to receive a Redis client from the DI builder.
type RedisSetter interface {
	SetRedis(c *Client)
}
