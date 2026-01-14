package client

// GRPCSetter is the interface that Link structs must implement
// to receive a gRPC client from the DI builder.
type GRPCSetter interface {
	SetGRPC(c *Client)
}
