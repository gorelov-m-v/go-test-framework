package dsl

import (
	"reflect"

	"go-test-framework/pkg/allure"
	"go-test-framework/pkg/http/client"

	"github.com/ozontech/allure-go/pkg/framework/provider"
)

var httpReporter = allure.NewDefaultReporter()

func attachRequest[TReq any](sCtx provider.StepCtx, httpClient *client.Client, req *client.Request[TReq]) {
	var body any
	if !isNil(req.Body) {
		body = req.Body
	}

	anyReq := &client.Request[any]{
		Method:      req.Method,
		Path:        req.Path,
		PathParams:  req.PathParams,
		QueryParams: req.QueryParams,
		Headers:     req.Headers,
		Body:        &body,
		RawBody:     req.RawBody,
		Multipart:   req.Multipart,
	}
	httpReporter.AttachHTTPRequest(sCtx, httpClient, anyReq)
}

func attachResponse[TResp any](sCtx provider.StepCtx, httpClient *client.Client, resp *client.Response[TResp]) {
	var anyResp *client.Response[any]
	if resp != nil {
		var body any
		if !isNil(resp.Body) {
			body = resp.Body
		}

		anyResp = &client.Response[any]{
			StatusCode:   resp.StatusCode,
			Headers:      resp.Headers,
			Body:         &body,
			RawBody:      resp.RawBody,
			Error:        resp.Error,
			Duration:     resp.Duration,
			NetworkError: resp.NetworkError,
		}
	}
	httpReporter.AttachHTTPResponse(sCtx, httpClient, anyResp)
}

func isNil(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	return v.Kind() == reflect.Ptr && v.IsNil()
}
