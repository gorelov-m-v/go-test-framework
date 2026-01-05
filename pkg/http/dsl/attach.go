package dsl

import (
	"go-test-framework/pkg/allure"
	"go-test-framework/pkg/http/client"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

var httpReporter = allure.NewDefaultReporter()

func attachRequest[TReq any](sCtx provider.StepCtx, httpClient *client.Client, req *client.Request[TReq]) {
	dto := allure.ToHTTPRequestDTO[TReq](req)
	httpReporter.AttachHTTPRequest(sCtx, httpClient, dto)
}

func attachResponse[TResp any](sCtx provider.StepCtx, httpClient *client.Client, resp *client.Response[TResp]) {
	dto := allure.ToHTTPResponseDTO[TResp](resp)
	httpReporter.AttachHTTPResponse(sCtx, httpClient, dto)
}
