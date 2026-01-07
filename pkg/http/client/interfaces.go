package client

type HTTPSetter interface {
	SetHTTP(c *Client)
}
