package client

type DBSetter interface {
	SetDB(c *Client)
}
