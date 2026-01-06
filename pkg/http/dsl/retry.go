package dsl

import (
	"context"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"go-test-framework/pkg/extension"
	"go-test-framework/pkg/http/client"
	"go-test-framework/pkg/retry"
)

type checkResult struct {
	ok        bool
	retryable bool
	reason    string
}

type expectation struct {
	name   string
	check  func(err error, resp *client.Response[any]) checkResult
	report func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, resp *client.Response[any], checkRes checkResult)
}

func newExpectation(
	name string,
	checkFn func(err error, resp *client.Response[any]) checkResult,
	reportFn func(stepCtx provider.StepCtx, mode extension.AssertionMode, err error, resp *client.Response[any], checkRes checkResult),
) *expectation {
	return &expectation{
		name:   name,
		check:  checkFn,
		report: reportFn,
	}
}

func (c *Call[TReq, TResp]) executeSingle() (*client.Response[TResp], error, extension.PollingSummary) {
	executor := func(ctx context.Context) (*client.Response[TResp], error) {
		resp, err := client.DoTyped[TReq, TResp](ctx, c.client, c.req)
		if err != nil && resp == nil {
			resp = &client.Response[TResp]{NetworkError: err.Error()}
		}

		if resp == nil {
			resp = &client.Response[TResp]{NetworkError: "nil response"}
			if err == nil {
				err = fmt.Errorf("unexpected nil response")
			}
		}

		return resp, err
	}

	resp, err, summary := retry.ExecuteSingle(c.ctx, executor)

	if err == nil && resp.NetworkError != "" {
		summary.Success = false
		summary.LastError = resp.NetworkError
	}

	return resp, err, summary
}

func (c *Call[TReq, TResp]) executeWithRetry(
	stepCtx provider.StepCtx,
	expectations []*expectation,
) (*client.Response[TResp], error, extension.PollingSummary) {
	asyncCfg := c.client.GetAsyncConfig()
	executor := func(ctx context.Context) (*client.Response[TResp], error) {
		resp, err := client.DoTyped[TReq, TResp](ctx, c.client, c.req)
		if err != nil && resp == nil {
			resp = &client.Response[TResp]{NetworkError: err.Error()}
		}

		if resp == nil {
			resp = &client.Response[TResp]{NetworkError: "nil response"}
			if err == nil {
				err = fmt.Errorf("unexpected nil response")
			}
		}

		return resp, err
	}

	checker := func(resp *client.Response[TResp], err error) []retry.CheckResult {
		respAny := &client.Response[any]{
			StatusCode:   resp.StatusCode,
			Headers:      resp.Headers,
			RawBody:      resp.RawBody,
			Error:        resp.Error,
			Duration:     resp.Duration,
			NetworkError: resp.NetworkError,
		}

		results := make([]retry.CheckResult, 0, len(expectations))
		for _, exp := range expectations {
			checkRes := exp.check(err, respAny)
			results = append(results, retry.CheckResult{
				Ok:        checkRes.ok,
				Retryable: checkRes.retryable,
				Reason:    checkRes.reason,
			})
		}

		return results
	}

	resp, err, summary := retry.ExecuteWithRetry(c.ctx, stepCtx, asyncCfg, executor, checker)

	if err == nil && resp != nil && resp.NetworkError != "" {
		if summary.Success {
			summary.Success = false
		}
		if summary.LastError == "" {
			summary.LastError = resp.NetworkError
		}
	}

	return resp, err, summary
}

func reportExpectations(
	stepCtx provider.StepCtx,
	mode extension.AssertionMode,
	expectations []*expectation,
	err error,
	resp *client.Response[any],
) {
	for _, exp := range expectations {
		checkRes := exp.check(err, resp)
		exp.report(stepCtx, mode, err, resp, checkRes)
	}
}

func validateJSONPath(path string) error {
	if path == "" {
		return fmt.Errorf("JSON path cannot be empty")
	}
	return nil
}
