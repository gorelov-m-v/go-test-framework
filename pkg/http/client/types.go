package client

import (
	"net/http"
	"time"
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
