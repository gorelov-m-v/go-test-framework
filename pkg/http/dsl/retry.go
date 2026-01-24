package dsl

import (
	"context"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/client"
)

func (c *Call[TReq, TResp]) executeSingle() (*client.Response[TResp], error, polling.PollingSummary) {
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
	expectations []*expect.Expectation[*client.Response[any]],
) (*client.Response[TResp], error, polling.PollingSummary) {
	asyncCfg := c.client.AsyncConfig

	if !asyncCfg.Enabled {
		return c.executeSingle()
	}

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

	checker := func(resp *client.Response[TResp], err error) []polling.CheckResult {
		respAny := &client.Response[any]{
			StatusCode:   resp.StatusCode,
			Headers:      resp.Headers,
			RawBody:      resp.RawBody,
			Error:        resp.Error,
			Duration:     resp.Duration,
			NetworkError: resp.NetworkError,
		}

		results := make([]polling.CheckResult, 0, len(expectations))
		for _, exp := range expectations {
			checkRes := exp.Check(err, respAny)
			results = append(results, polling.CheckResult{
				Ok:        checkRes.Ok,
				Retryable: checkRes.Retryable,
				Reason:    checkRes.Reason,
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
