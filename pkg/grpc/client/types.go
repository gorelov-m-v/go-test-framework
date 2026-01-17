package client

import (
	"time"

	"google.golang.org/grpc/metadata"
)

type Request[TReq any] struct {
	Service  string
	Method   string
	Body     *TReq
	Metadata metadata.MD
}

type Response[TResp any] struct {
	Body     *TResp
	Metadata metadata.MD
	Duration time.Duration
	Error    error
	RawBody  []byte
}
