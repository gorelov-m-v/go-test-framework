package dsl

import (
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/allure"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

var httpReporter = allure.NewDefaultReporter()

func attachHTTPReport[TReq, TResp any](
	stepCtx provider.StepCtx,
	httpClient *client.Client,
	req *client.Request[TReq],
	resp *client.Response[TResp],
	pollingSummary polling.PollingSummary,
) {
	report := allure.HTTPReportDTO{
		Request:  allure.ToHTTPRequestDTO(req),
		Response: allure.ToHTTPResponseDTO(resp),
		Polling:  allure.ToPollingSummaryDTO(pollingSummary),
	}

	httpReporter.AttachHTTPReport(stepCtx, httpClient, report)
}
