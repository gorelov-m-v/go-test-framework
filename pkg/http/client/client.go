package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	"github.com/gorelov-m-v/go-test-framework/pkg/contract"
)

type Client struct {
	BaseURL           string
	HTTPClient        *http.Client
	DefaultHeaders    map[string]string
	AsyncConfig       config.AsyncConfig
	ContractValidator *contract.Validator
	ContractBasePath  string
	maskHeaders       map[string]bool
}

type Config struct {
	BaseURL          string
	Timeout          time.Duration
	DefaultHeaders   map[string]string
	MaskHeaders      string             `mapstructure:"maskHeaders"`
	ContractSpec     string             `mapstructure:"contractSpec"`
	ContractBasePath string             `mapstructure:"contractBasePath"`
	AsyncConfig      config.AsyncConfig `mapstructure:"asyncConfig"`
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

	asyncCfg := cfg.AsyncConfig
	if asyncCfg.Timeout == 0 {
		asyncCfg = config.DefaultAsyncConfig()
	}

	var contractValidator *contract.Validator
	if cfg.ContractSpec != "" {
		var err error
		contractValidator, err = contract.NewValidator(cfg.ContractSpec)
		if err != nil {
			panic(fmt.Sprintf("failed to load contract spec '%s': %v", cfg.ContractSpec, err))
		}
	}

	return &Client{
		BaseURL: cfg.BaseURL,
		HTTPClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		DefaultHeaders:    cfg.DefaultHeaders,
		AsyncConfig:       asyncCfg,
		ContractValidator: contractValidator,
		ContractBasePath:  cfg.ContractBasePath,
		maskHeaders:       maskHeaders,
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
