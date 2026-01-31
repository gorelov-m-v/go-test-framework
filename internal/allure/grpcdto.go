package allure

import (
	"time"

	"google.golang.org/grpc/metadata"
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
