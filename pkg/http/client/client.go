package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorelov-m-v/go-test-framework/pkg/config"
	"github.com/gorelov-m-v/go-test-framework/pkg/http/contract"
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
	BaseURL          string                `mapstructure:"baseURL" yaml:"baseURL" json:"baseURL"`
	Timeout          time.Duration         `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	DefaultHeaders   map[string]string     `mapstructure:"defaultHeaders" yaml:"defaultHeaders" json:"defaultHeaders"`
	MaskHeaders      string                `mapstructure:"maskHeaders" yaml:"maskHeaders" json:"maskHeaders"`
	ContractSpec     string                `mapstructure:"contractSpec" yaml:"contractSpec" json:"contractSpec"`
	ContractBasePath string                `mapstructure:"contractBasePath" yaml:"contractBasePath" json:"contractBasePath"`
	AsyncConfig      config.AsyncConfig    `mapstructure:"async" yaml:"async" json:"async"`
}

func New(cfg Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	maskHeaders := parseMaskHeaders(cfg.MaskHeaders)

	asyncCfg := cfg.AsyncConfig.WithDefaults()

	var contractValidator *contract.Validator
	if cfg.ContractSpec != "" {
		var err error
		contractValidator, err = contract.NewValidator(cfg.ContractSpec)
		if err != nil {
			return nil, fmt.Errorf("failed to load contract spec '%s': %w", cfg.ContractSpec, err)
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
	}, nil
}

func parseMaskHeaders(maskHeaders string) map[string]bool {
	if maskHeaders == "" {
		return nil
	}
	result := make(map[string]bool)
	for _, part := range strings.Split(maskHeaders, ",") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result[strings.ToLower(trimmed)] = true
		}
	}
	return result
}

func (c *Client) ShouldMaskHeader(name string) bool {
	if c.maskHeaders == nil {
		return false
	}
	return c.maskHeaders[strings.ToLower(strings.TrimSpace(name))]
}

func (c *Client) GetBaseURL() string {
	return c.BaseURL
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) BuildEffectiveURL(path string, pathParams, queryParams map[string]string) (string, error) {
	return BuildEffectiveURL(c.BaseURL, path, pathParams, queryParams)
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
