package client

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type Request[T any] struct {
	Method      string
	Path        string
	PathParams  map[string]string
	QueryParams map[string]string
	Headers     map[string]string
	Body        *T
	BodyMap     map[string]interface{}
	RawBody     []byte
	Multipart   *MultipartForm
}

type Response[V any] struct {
	StatusCode   int
	Headers      http.Header
	Body         V
	RawBody      []byte
	Error        *ErrorResponse
	Duration     time.Duration
	NetworkError string
}

func (r *Response[V]) GetNetworkError() string {
	if r == nil {
		return ""
	}
	return r.NetworkError
}

func (r *Response[V]) GetError() error {
	if r == nil {
		return nil
	}
	if r.NetworkError != "" {
		return errors.New(r.NetworkError)
	}
	return nil
}

func ResponsePreCheckConfig() expect.PreCheckConfig[*Response[any]] {
	return expect.PreCheckConfig[*Response[any]]{
		IsNil:           func(r *Response[any]) bool { return r == nil },
		GetNetworkError: func(r *Response[any]) string { return r.NetworkError },
		EmptyBodyCheck:  func(r *Response[any]) bool { return len(r.RawBody) == 0 },
	}
}

func BuildPreCheck() func(error, *Response[any]) (polling.CheckResult, bool) {
	return expect.BuildPreCheck(ResponsePreCheckConfig())
}

func BuildPreCheckWithBody() func(error, *Response[any]) (polling.CheckResult, bool) {
	return expect.BuildPreCheckWithBody(ResponsePreCheckConfig())
}

func (r *Response[V]) ToAny() *Response[any] {
	if r == nil {
		return nil
	}
	return &Response[any]{
		StatusCode:   r.StatusCode,
		Headers:      r.Headers,
		RawBody:      r.RawBody,
		Error:        r.Error,
		Duration:     r.Duration,
		NetworkError: r.NetworkError,
	}
}

type ErrorResponse struct {
	Body       string
	StatusCode int
	Message    string
	Errors     map[string][]string
}

type MultipartForm struct {
	Fields map[string]string
	Files  []MultipartFile
}

type MultipartFile struct {
	FieldName string
	FileName  string
	Content   []byte
}
