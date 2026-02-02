package allure

import (
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

type GRPCRequestDTO struct {
	Target   string
	Method   string
	Metadata metadata.MD
	Body     any
}

type GRPCResponseDTO struct {
	StatusCode int
	Status     string
	Metadata   metadata.MD
	Body       any
	Error      error
	Duration   time.Duration
}

func ToGRPCRequestDTO[T any](target, method string, body *T, md metadata.MD) GRPCRequestDTO {
	dto := GRPCRequestDTO{
		Target:   target,
		Method:   method,
		Metadata: md,
	}
	if body != nil {
		dto.Body = body
	}
	return dto
}

func ToGRPCResponseDTO[T any](resp *client.Response[T]) GRPCResponseDTO {
	dto := GRPCResponseDTO{}

	if resp == nil {
		return dto
	}

	dto.Duration = resp.Duration
	dto.Metadata = resp.Metadata
	dto.Error = resp.Error

	if resp.Error != nil {
		st, ok := status.FromError(resp.Error)
		if ok {
			dto.Status = st.Code().String()
			dto.StatusCode = int(st.Code())
		} else {
			dto.Status = "UNKNOWN"
			dto.StatusCode = -1
		}
	} else {
		dto.Status = "OK"
		dto.StatusCode = 0
	}

	if resp.Body != nil {
		dto.Body = resp.Body
	}

	return dto
}
