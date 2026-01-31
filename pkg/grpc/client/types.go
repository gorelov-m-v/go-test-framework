package client

import (
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
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

func (r *Response[TResp]) GetError() error {
	if r == nil {
		return nil
	}
	return r.Error
}

func ResponsePreCheckConfig() expect.PreCheckConfig[*Response[any]] {
	return expect.PreCheckConfig[*Response[any]]{
		IsNil:          func(r *Response[any]) bool { return r == nil },
		EmptyBodyCheck: func(r *Response[any]) bool { return r.Body == nil },
	}
}

func BuildPreCheck() func(error, *Response[any]) (polling.CheckResult, bool) {
	return expect.BuildPreCheck(ResponsePreCheckConfig())
}

func BuildPreCheckWithBody() func(error, *Response[any]) (polling.CheckResult, bool) {
	return expect.BuildPreCheckWithBody(ResponsePreCheckConfig())
}

func (r *Response[TResp]) ToAny() *Response[any] {
	if r == nil {
		return nil
	}
	respAny := &Response[any]{
		Metadata: r.Metadata,
		Duration: r.Duration,
		Error:    r.Error,
		RawBody:  r.RawBody,
	}
	if r.Body != nil {
		var bodyAny any = r.Body
		respAny.Body = &bodyAny
	}
	return respAny
}
