package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
)

// Client is the gRPC client wrapper
type Client struct {
	conn        *grpc.ClientConn
	target      string
	AsyncConfig config.AsyncConfig
}

// Config holds the gRPC client configuration
type Config struct {
	Target      string             `mapstructure:"target" yaml:"target" json:"target"`
	Timeout     time.Duration      `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	Insecure    bool               `mapstructure:"insecure" yaml:"insecure" json:"insecure"`
	AsyncConfig config.AsyncConfig `mapstructure:"asyncConfig" yaml:"asyncConfig" json:"asyncConfig"`
}

// New creates a new gRPC client
func New(cfg Config) (*Client, error) {
	if cfg.Target == "" {
		return nil, fmt.Errorf("gRPC target address is required")
	}

	opts := []grpc.DialOption{}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.Dial(cfg.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	asyncCfg := cfg.AsyncConfig
	if asyncCfg.Timeout == 0 {
		asyncCfg = defaultAsyncConfig()
	}

	return &Client{
		conn:        conn,
		target:      cfg.Target,
		AsyncConfig: asyncCfg,
	}, nil
}

func defaultAsyncConfig() config.AsyncConfig {
	return config.AsyncConfig{
		Enabled:  true,
		Timeout:  10 * time.Second,
		Interval: 200 * time.Millisecond,
		Backoff: config.BackoffConfig{
			Enabled:     true,
			Factor:      1.5,
			MaxInterval: 1 * time.Second,
		},
		Jitter: 0.2,
	}
}

// Conn returns the underlying gRPC connection
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Target returns the target address
func (c *Client) Target() string {
	return c.target
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Invoke calls a gRPC method using the generic invoker
func Invoke[TReq any, TResp any](
	ctx context.Context,
	c *Client,
	fullMethod string,
	req *TReq,
	md metadata.MD,
) (*Response[TResp], error) {
	start := time.Now()

	if ctx == nil {
		ctx = context.Background()
	}

	if md != nil {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	var headerMD, trailerMD metadata.MD
	resp := new(TResp)

	err := c.conn.Invoke(
		ctx,
		fullMethod,
		req,
		resp,
		grpc.Header(&headerMD),
		grpc.Trailer(&trailerMD),
	)

	duration := time.Since(start)

	response := &Response[TResp]{
		Body:     resp,
		Duration: duration,
		Error:    err,
	}

	// Merge header and trailer metadata
	responseMD := metadata.Join(headerMD, trailerMD)
	response.Metadata = responseMD

	// Serialize response body to raw bytes for inspection
	if resp != nil {
		if protoMsg, ok := any(resp).(proto.Message); ok {
			response.RawBody, _ = proto.Marshal(protoMsg)
		} else {
			response.RawBody, _ = json.Marshal(resp)
		}
	}

	return response, err
}
