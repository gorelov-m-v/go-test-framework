package dsl

import (
	"encoding/json"
	"fmt"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"google.golang.org/grpc/status"

	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

type grpcRequestDTO struct {
	Target     string            `json:"target"`
	Method     string            `json:"method"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Body       any               `json:"body,omitempty"`
}

type grpcResponseDTO struct {
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	Duration   string            `json:"duration"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Body       any               `json:"body,omitempty"`
	Error      string            `json:"error,omitempty"`
}

func attachRequest[TReq any, TResp any](stepCtx provider.StepCtx, c *Call[TReq, TResp]) {
	dto := grpcRequestDTO{
		Target: c.client.Target(),
		Method: c.fullMethod,
	}

	if len(c.metadata) > 0 {
		dto.Metadata = make(map[string]string)
		for k, vals := range c.metadata {
			if len(vals) > 0 {
				dto.Metadata[k] = vals[0]
			}
		}
	}

	if c.body != nil {
		dto.Body = c.body
	}

	jsonBytes, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		stepCtx.Logf("Failed to marshal gRPC request: %v", err)
		return
	}

	stepCtx.WithAttachments(allure.NewAttachment("gRPC Request", allure.JSON, jsonBytes))
}

func attachResponse[TResp any](stepCtx provider.StepCtx, resp *client.Response[TResp]) {
	if resp == nil {
		return
	}

	dto := grpcResponseDTO{
		Duration: resp.Duration.String(),
	}

	// Extract status from error
	if resp.Error != nil {
		st, ok := status.FromError(resp.Error)
		if ok {
			dto.Status = st.Code().String()
			dto.StatusCode = int(st.Code())
			dto.Error = st.Message()
		} else {
			dto.Status = "UNKNOWN"
			dto.StatusCode = -1
			dto.Error = resp.Error.Error()
		}
	} else {
		dto.Status = "OK"
		dto.StatusCode = 0
	}

	if len(resp.Metadata) > 0 {
		dto.Metadata = make(map[string]string)
		for k, vals := range resp.Metadata {
			if len(vals) > 0 {
				dto.Metadata[k] = vals[0]
			}
		}
	}

	if resp.Body != nil {
		dto.Body = resp.Body
	}

	jsonBytes, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		stepCtx.Logf("Failed to marshal gRPC response: %v", err)
		return
	}

	// Determine attachment name based on status
	attachmentName := fmt.Sprintf("gRPC Response [%s]", dto.Status)
	stepCtx.WithAttachments(allure.NewAttachment(attachmentName, allure.JSON, jsonBytes))
}
