package allure

import (
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

type HTTPRequestDTO struct {
	Method      string
	Path        string
	PathParams  map[string]string
	QueryParams map[string]string
	Headers     map[string]string

	Body      any
	RawBody   []byte
	Multipart *client.MultipartForm
}

type HTTPResponseDTO struct {
	StatusCode   int
	Headers      map[string][]string
	Body         any
	RawBody      []byte
	Error        *client.ErrorResponse
	Duration     time.Duration
	NetworkError string
}

func ToHTTPRequestDTO[T any](req *client.Request[T]) HTTPRequestDTO {
	dto := HTTPRequestDTO{}

	if req == nil {
		return dto
	}

	dto.Method = req.Method
	dto.Path = req.Path
	dto.PathParams = req.PathParams
	dto.QueryParams = req.QueryParams
	dto.Headers = req.Headers
	dto.RawBody = req.RawBody
	dto.Multipart = req.Multipart

	if req.Body != nil {
		dto.Body = *req.Body
	} else if req.BodyMap != nil {
		dto.Body = req.BodyMap
	}

	return dto
}

func ToHTTPResponseDTO[T any](resp *client.Response[T]) HTTPResponseDTO {
	dto := HTTPResponseDTO{}

	if resp == nil {
		return dto
	}

	dto.StatusCode = resp.StatusCode
	dto.Headers = resp.Headers
	dto.RawBody = resp.RawBody
	dto.Error = resp.Error
	dto.Duration = resp.Duration
	dto.NetworkError = resp.NetworkError

	dto.Body = resp.Body

	return dto
}
