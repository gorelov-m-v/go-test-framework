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

type Client struct {
	conn        *grpc.ClientConn
	target      string
	AsyncConfig config.AsyncConfig
}

type Config struct {
	Target      string             `mapstructure:"target" yaml:"target" json:"target"`
	Timeout     time.Duration      `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	Insecure    bool               `mapstructure:"insecure" yaml:"insecure" json:"insecure"`
	AsyncConfig config.AsyncConfig `mapstructure:"asyncConfig" yaml:"asyncConfig" json:"asyncConfig"`
}

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
		asyncCfg = config.DefaultAsyncConfig()
	}

	return &Client{
		conn:        conn,
		target:      cfg.Target,
		AsyncConfig: asyncCfg,
	}, nil
}

func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *Client) Target() string {
	return c.target
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

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

	responseMD := metadata.Join(headerMD, trailerMD)
	response.Metadata = responseMD

	if resp != nil {
		if protoMsg, ok := any(resp).(proto.Message); ok {
			response.RawBody, _ = proto.Marshal(protoMsg)
		} else {
			response.RawBody, _ = json.Marshal(resp)
		}
	}

	return response, err
}
