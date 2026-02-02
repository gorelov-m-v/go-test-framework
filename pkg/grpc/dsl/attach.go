package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

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
	report := allure.GRPCReportDTO{
		Request:  allure.ToGRPCRequestDTO(c.client.Target(), c.fullMethod, c.body, c.metadata),
		Response: allure.ToGRPCResponseDTO(resp),
		Polling:  allure.ToPollingSummaryDTO(pollingSummary),
	}

	grpcReporter.AttachGRPCReport(stepCtx, report)
}
