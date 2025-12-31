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
	secretHeaders  map[string]bool
}

type Config struct {
	BaseURL        string
	Timeout        time.Duration
	DefaultHeaders map[string]string
	SecretHeaders  []string
}

func New(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	secretHeaders := make(map[string]bool)

	addSecret := func(h string) {
		h = strings.TrimSpace(h)
		if h == "" {
			return
		}
		secretHeaders[strings.ToLower(h)] = true
	}

	if len(cfg.SecretHeaders) == 0 {
		defaultSecrets := []string{
			"Authorization", "Cookie", "Set-Cookie",
			"X-Api-Key", "Api-Key", "X-Token", "Token",
		}
		for _, h := range defaultSecrets {
			addSecret(h)
		}
	} else {
		for _, h := range cfg.SecretHeaders {
			addSecret(h)
		}
	}

	return &Client{
		BaseURL: cfg.BaseURL,
		HTTPClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		DefaultHeaders: cfg.DefaultHeaders,
		secretHeaders:  secretHeaders,
	}
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

func (c *Client) IsSecretHeader(name string) bool {
	return c.secretHeaders[strings.ToLower(name)]
}
