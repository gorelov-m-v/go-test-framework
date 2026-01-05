package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL        string
	HTTPClient     *http.Client
	DefaultHeaders map[string]string
	maskHeaders    map[string]bool
}

type Config struct {
	BaseURL        string
	Timeout        time.Duration
	DefaultHeaders map[string]string
	MaskHeaders    string `mapstructure:"maskHeaders"`
}

func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	var maskHeaders map[string]bool
	if cfg.MaskHeaders != "" {
		maskHeaders = make(map[string]bool)
		parts := strings.Split(cfg.MaskHeaders, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				maskHeaders[strings.ToLower(trimmed)] = true
			}
		}
	}

	return &Client{
		BaseURL: cfg.BaseURL,
		HTTPClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		DefaultHeaders: cfg.DefaultHeaders,
		maskHeaders:    maskHeaders,
	}
}

func (c *Client) ShouldMaskHeader(name string) bool {
	if c.maskHeaders == nil {
		return false
	}
	return c.maskHeaders[strings.ToLower(strings.TrimSpace(name))]
}

func (c *Client) Do(ctx context.Context, req *Request[any]) (*Response[any], error) {
	return DoTyped[any, any](ctx, c, req)
}

func DoTyped[TReq any, TResp any](ctx context.Context, c *Client, req *Request[TReq]) (*Response[TResp], error) {
	start := time.Now()

	httpReq, err := buildRequest(ctx, c, req)
	if err != nil {
		return &Response[TResp]{
			NetworkError: fmt.Sprintf("failed to build request: %v", err),
			Duration:     time.Since(start),
		}, err
	}

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return &Response[TResp]{
			NetworkError: fmt.Sprintf("request failed: %v", err),
			Duration:     time.Since(start),
		}, err
	}
	defer resp.Body.Close()

	return decodeResponse[TResp](resp, time.Since(start))
}
