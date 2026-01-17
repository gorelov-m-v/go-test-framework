package client

type RedisSetter interface {
	SetRedis(c *Client)
}
