package httpdsl

import (
	"bytes"
	"encoding/json"
	"fmt"

	"go-test-framework/pkg/httpclient"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func attachRequest[TReq any](sCtx provider.StepCtx, client *httpclient.Client, req *httpclient.Request[TReq]) {
	rb := newReportBuilder(client)

	rb.writeRequestBasicInfo(req.Method, req.Path, req.PathParams, req.QueryParams)
	rb.writeParams(req.PathParams, "Path Params")
	rb.writeParams(req.QueryParams, "Query Params")
	rb.writeRequestHeaders(req.Headers)

	var body any = req.Body
	if req.Body == nil {
		body = nil
	}
	rb.writeRequestBody(body, req.RawBody, req.Multipart)

	sCtx.WithNewAttachment("HTTP Request", allure.Text, rb.Bytes())
}

func attachResponse[TResp any](sCtx provider.StepCtx, _ *httpclient.Client, resp *httpclient.Response[TResp]) {
	rb := newReportBuilder(nil)

	if resp == nil {
		rb.writeLine("Response: <nil>")
		sCtx.WithNewAttachment("HTTP Response", allure.Text, rb.Bytes())
		return
	}

	if resp.NetworkError != "" {
		rb.writeLine("Network Error: %s", resp.NetworkError)
		rb.writeLine("Duration: %v", resp.Duration)
		sCtx.WithNewAttachment("HTTP Response", allure.Text, rb.Bytes())
		return
	}

	rb.writeResponseStatus(resp.StatusCode, resp.Duration)
	rb.writeResponseHeaders(resp.Headers)
	rb.writeResponseError(resp.Error)
	rb.writeResponseBody(resp.RawBody)

	sCtx.WithNewAttachment("HTTP Response", allure.Text, rb.Bytes())
}

type reportBuilder struct {
	buf    bytes.Buffer
	client *httpclient.Client
}

func newReportBuilder(c *httpclient.Client) *reportBuilder {
	return &reportBuilder{client: c}
}

func (b *reportBuilder) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *reportBuilder) writeLine(format string, args ...any) {
	b.buf.WriteString(fmt.Sprintf(format, args...))
	b.buf.WriteString("\n")
}

func (b *reportBuilder) writeRequestBasicInfo(method, path string, pathParams, queryParams map[string]string) {
	b.writeLine("Method: %s", method)
	b.writeLine("Path: %s", path)

	if b.client != nil {
		if eff, err := httpclient.BuildEffectiveURL(b.client.BaseURL, path, pathParams, queryParams); err == nil {
			b.writeLine("Effective URL: %s", eff)
		} else {
			b.writeLine("Effective URL: (failed to resolve: %v)", err)
		}
	}
}

func (b *reportBuilder) writeParams(params map[string]string, title string) {
	if len(params) == 0 {
		return
	}
	b.writeLine("\n%s:", title)
	for k, v := range params {
		b.writeLine("  %s: %s", k, v)
	}
}

func (b *reportBuilder) writeRequestHeaders(headers map[string]string) {
	if len(headers) == 0 {
		return
	}
	b.writeLine("\nHeaders:")
	sanitized := headers
	if b.client != nil {
		sanitized = b.client.SanitizeHeaders(headers)
	}
	for k, v := range sanitized {
		b.writeLine("  %s: %s", k, v)
	}
}

func (b *reportBuilder) writeRequestBody(body any, rawBody []byte, multipart *httpclient.MultipartForm) {
	switch {
	case body != nil:
		b.writeLine("\nBody (json):")
		bytes, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			b.writeLine("(failed to marshal: %v)", err)
		} else {
			b.buf.Write(bytes)
			b.buf.WriteString("\n")
		}

	case len(rawBody) > 0:
		b.writeLine("\nBody (raw):")
		b.writeTruncated(rawBody)

	case multipart != nil:
		b.writeLine("\nBody (multipart/form-data):")
		b.writeLine("Fields:")
		for k, v := range multipart.Fields {
			b.writeLine("  %s: %s", k, v)
		}
		if len(multipart.Files) > 0 {
			b.writeLine("Files:")
			for _, f := range multipart.Files {
				b.writeLine("  %s: %s (%d bytes)", f.FieldName, f.FileName, len(f.Content))
			}
		}

	default:
		b.writeLine("\nBody: (empty)")
	}
}

func (b *reportBuilder) writeResponseStatus(code int, duration interface{}) {
	b.writeLine("Status: %d", code)
	b.writeLine("Duration: %v", duration)
}

func (b *reportBuilder) writeResponseHeaders(headers map[string][]string) {
	if len(headers) == 0 {
		return
	}
	b.writeLine("\nHeaders:")
	for k, values := range headers {
		for _, v := range values {
			b.writeLine("  %s: %s", k, v)
		}
	}
}

func (b *reportBuilder) writeResponseError(err *httpclient.ErrorResponse) {
	if err == nil {
		return
	}
	b.writeLine("\nError:")
	if err.Message != "" {
		b.writeLine("  Message: %s", err.Message)
	}
	if len(err.Errors) > 0 {
		b.writeLine("  Errors:")
		for field, messages := range err.Errors {
			for _, msg := range messages {
				b.writeLine("    %s: %s", field, msg)
			}
		}
	}
	if err.Message == "" && len(err.Errors) == 0 {
		b.writeLine("  Body: %s", err.Body)
	}
}

func (b *reportBuilder) writeResponseBody(rawBody []byte) {
	b.writeLine("\nBody:")
	if len(rawBody) == 0 {
		b.writeLine("(empty)")
		return
	}

	var jsonData any
	if err := json.Unmarshal(rawBody, &jsonData); err == nil {
		bodyJSON, err := json.MarshalIndent(jsonData, "", "  ")
		if err == nil {
			b.buf.Write(bodyJSON)
			b.buf.WriteString("\n")
			return
		}
	}

	b.writeTruncated(rawBody)
}

func (b *reportBuilder) writeTruncated(data []byte) {
	if len(data) <= 1000 {
		b.buf.Write(data)
		b.buf.WriteString("\n")
	} else {
		b.writeLine("(raw body, %d bytes)", len(data))
		b.buf.Write(data[:1000])
		b.buf.WriteString("\n...\n")
	}
}
