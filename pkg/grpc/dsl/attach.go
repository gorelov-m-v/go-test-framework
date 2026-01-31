package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"google.golang.org/grpc/status"

	"github.com/gorelov-m-v/go-test-framework/internal/allure"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

var grpcReporter = allure.NewDefaultReporter()

func attachGRPCReport[TReq, TResp any](
	stepCtx provider.StepCtx,
	c *Call[TReq, TResp],
	resp *client.Response[TResp],
	pollingSummary polling.PollingSummary,
) {
	reqDTO := allure.GRPCRequestDTO{
		Target:   c.client.Target(),
		Method:   c.fullMethod,
		Metadata: c.metadata,
	}
	if c.body != nil {
		reqDTO.Body = c.body
	}

	respDTO := allure.GRPCResponseDTO{
		Duration: resp.Duration,
		Metadata: resp.Metadata,
		Error:    resp.Error,
	}

	if resp.Error != nil {
		st, ok := status.FromError(resp.Error)
		if ok {
			respDTO.Status = st.Code().String()
			respDTO.StatusCode = int(st.Code())
		} else {
			respDTO.Status = "UNKNOWN"
			respDTO.StatusCode = -1
		}
	} else {
		respDTO.Status = "OK"
		respDTO.StatusCode = 0
	}

	if resp.Body != nil {
		respDTO.Body = resp.Body
	}

	report := allure.GRPCReportDTO{
		Request:  reqDTO,
		Response: respDTO,
		Polling:  allure.ToPollingSummaryDTO(pollingSummary),
	}

	grpcReporter.AttachGRPCReport(stepCtx, report)
}
