package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

func BuildEffectiveURL(base string, pathTemplate string, pathParams map[string]string, queryParams map[string]string) (string, error) {
	u, err := buildResolvedURL(base, pathTemplate, pathParams, queryParams)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func buildRequest[TReq any](ctx context.Context, c *Client, req *Request[TReq]) (*http.Request, error) {
	if err := validateBuildInput(c, req); err != nil {
		return nil, err
	}

	u, err := buildResolvedURL(c.BaseURL, req.Path, req.PathParams, req.QueryParams)
	if err != nil {
		return nil, err
	}

	body, contentType, err := buildBody(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	applyHeaders(httpReq.Header, c.DefaultHeaders, req.Headers)
	setContentTypeIfMissing(httpReq.Header, contentType)

	return httpReq, nil
}

func validateBuildInput[TReq any](c *Client, req *Request[TReq]) error {
	if c == nil {
		return fmt.Errorf("httpclient is nil")
	}
	if req == nil {
		return fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("base URL is empty")
	}
	if strings.TrimSpace(req.Method) == "" {
		return fmt.Errorf("request method is empty")
	}
	if strings.TrimSpace(req.Path) == "" {
		return fmt.Errorf("request path is empty")
	}
	return nil
}

func buildResolvedURL(base string, pathTemplate string, pathParams map[string]string, queryParams map[string]string) (*url.URL, error) {
	baseURL, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("base URL must include scheme and host: %q", base)
	}

	if baseURL.Path != "" && !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	fullPath := applyPathParams(strings.TrimSpace(pathTemplate), pathParams)
	fullPath = strings.TrimLeft(fullPath, "/")

	relURL, err := url.Parse(fullPath)
	if err != nil {
		return nil, fmt.Errorf("invalid request path: %w", err)
	}
	if relURL.IsAbs() {
		return nil, fmt.Errorf("request path must be relative, got absolute URL: %q", fullPath)
	}

	resolvedURL := baseURL.ResolveReference(relURL)

	if len(queryParams) > 0 {
		q := resolvedURL.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		resolvedURL.RawQuery = q.Encode()
	}

	return resolvedURL, nil
}

func applyPathParams(pathTemplate string, pathParams map[string]string) string {
	if len(pathParams) == 0 {
		return pathTemplate
	}
	fullPath := pathTemplate
	for key, value := range pathParams {
		placeholder := "{" + key + "}"
		fullPath = strings.ReplaceAll(fullPath, placeholder, url.PathEscape(value))
	}
	return fullPath
}

func buildBody[TReq any](req *Request[TReq]) (io.Reader, string, error) {
	hasMultipart := req.Multipart != nil
	hasBody := req.Body != nil
	hasBodyMap := req.BodyMap != nil
	hasRawBody := len(req.RawBody) > 0

	if countTrue(hasMultipart, hasBody, hasBodyMap, hasRawBody) > 1 {
		return nil, "", fmt.Errorf("only one body type can be set: Body, BodyMap, RawBody, or Multipart")
	}

	switch {
	case hasMultipart:
		return buildMultipartBody(req.Multipart)
	case hasBodyMap:
		return buildJSONBody(req.BodyMap)
	case hasBody:
		return buildJSONBody(req.Body)
	case hasRawBody:
		return bytes.NewReader(req.RawBody), "", nil
	default:
		return nil, "", nil
	}
}

func countTrue(flags ...bool) int {
	count := 0
	for _, f := range flags {
		if f {
			count++
		}
	}
	return count
}

func buildJSONBody(body any) (io.Reader, string, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal JSON body: %w", err)
	}
	return bytes.NewReader(jsonBytes), "application/json", nil
}

func buildMultipartBody(m *MultipartForm) (io.Reader, string, error) {
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	for key, value := range m.Fields {
		if err := writer.WriteField(key, value); err != nil {
			_ = writer.Close()
			return nil, "", fmt.Errorf("failed to write multipart field %q: %w", key, err)
		}
	}

	for _, file := range m.Files {
		part, err := writer.CreateFormFile(file.FieldName, file.FileName)
		if err != nil {
			_ = writer.Close()
			return nil, "", fmt.Errorf("failed to create multipart file part for field %q: %w", file.FieldName, err)
		}
		if _, err := part.Write(file.Content); err != nil {
			_ = writer.Close()
			return nil, "", fmt.Errorf("failed to write multipart file content for field %q: %w", file.FieldName, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close multipart writer: %w", err)
	}

	return bodyBuf, writer.FormDataContentType(), nil
}

func applyHeaders(h http.Header, defaults map[string]string, specific map[string]string) {
	for key, value := range defaults {
		h.Set(key, value)
	}
	for key, value := range specific {
		h.Set(key, value)
	}
}

func setContentTypeIfMissing(h http.Header, contentType string) {
	if contentType == "" {
		return
	}
	if h.Get("Content-Type") != "" {
		return
	}
	h.Set("Content-Type", contentType)
}
