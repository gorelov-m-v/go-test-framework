package client

import (
	"time"

	"google.golang.org/grpc/metadata"
)

// Request represents a gRPC request
type Request[TReq any] struct {
	Service  string
	Method   string
	Body     *TReq
	Metadata metadata.MD
}

// Response represents a gRPC response
type Response[TResp any] struct {
	Body     *TResp
	Metadata metadata.MD
	Duration time.Duration
	Error    error
	RawBody  []byte
}
