package dsl

import (
	"context"
	"fmt"

	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
	"github.com/gorelov-m-v/go-test-framework/internal/retry"
	"github.com/gorelov-m-v/go-test-framework/pkg/grpc/client"
)

func (c *Call[TReq, TResp]) executeSingle() (*client.Response[TResp], error, polling.PollingSummary) {
	executor := func(ctx context.Context) (*client.Response[TResp], error) {
		resp, err := client.Invoke[TReq, TResp](ctx, c.client, c.fullMethod, c.body, c.metadata)
		if err != nil && resp == nil {
			resp = &client.Response[TResp]{Error: err}
		}

		if resp == nil {
			resp = &client.Response[TResp]{Error: fmt.Errorf("nil response")}
			if err == nil {
				err = fmt.Errorf("unexpected nil response")
			}
		}

		return resp, err
	}

	resp, err, summary := retry.ExecuteSingle(c.ctx, executor)

	if err == nil && resp != nil && resp.Error != nil {
		summary.Success = false
		summary.LastError = resp.Error.Error()
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
		resp, err := client.Invoke[TReq, TResp](ctx, c.client, c.fullMethod, c.body, c.metadata)
		if err != nil && resp == nil {
			resp = &client.Response[TResp]{Error: err}
		}

		if resp == nil {
			resp = &client.Response[TResp]{Error: fmt.Errorf("nil response")}
			if err == nil {
				err = fmt.Errorf("unexpected nil response")
			}
		}

		return resp, err
	}

	checker := retry.BuildExpectationsCheckerWithConvert(expectations, func(resp *client.Response[TResp]) *client.Response[any] {
		respAny := &client.Response[any]{
			Body:     nil,
			Metadata: nil,
			Duration: 0,
			Error:    nil,
		}

		if resp != nil {
			if resp.Body != nil {
				var bodyAny any = resp.Body
				respAny.Body = &bodyAny
			}
			respAny.Metadata = resp.Metadata
			respAny.Duration = resp.Duration
			respAny.Error = resp.Error
			respAny.RawBody = resp.RawBody
		}

		return respAny
	})

	resp, err, summary := retry.ExecuteWithRetry(c.ctx, stepCtx, asyncCfg, executor, checker)

	retry.PostProcessSummary(resp, err, &summary)

	return resp, err, summary
}
