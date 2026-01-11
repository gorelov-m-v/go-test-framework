package allure

import (
	"encoding/json"
	"strings"

	"go-test-framework/pkg/http/client"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func (r *Reporter) AttachHTTPRequest(sCtx provider.StepCtx, httpClient *client.Client, req HTTPRequestDTO) {
	builder := NewReportBuilder()

	r.writeRequestBasicInfo(builder, httpClient, req.Method, req.Path, req.PathParams, req.QueryParams)
	r.writeParams(builder, req.PathParams, "Path Params")
	r.writeParams(builder, req.QueryParams, "Query Params")
	r.writeRequestHeaders(builder, httpClient, req.Headers)
	r.writeRequestBody(builder, req.Body, req.RawBody, req.Multipart)

	sCtx.WithNewAttachment("HTTP Request", allure.Text, builder.Bytes())
}

func (r *Reporter) AttachHTTPResponse(sCtx provider.StepCtx, httpClient *client.Client, resp HTTPResponseDTO) {
	builder := NewReportBuilder()

	if resp.NetworkError != "" {
		builder.WriteLine("Network Error: %s", resp.NetworkError)
		builder.WriteLine("Duration: %v", resp.Duration)
		sCtx.WithNewAttachment("HTTP Response", allure.Text, builder.Bytes())
		return
	}

	r.writeResponseStatus(builder, resp.StatusCode, resp.Duration)
	r.writeResponseHeaders(builder, httpClient, resp.Headers)
	r.writeResponseError(builder, resp.Error)
	r.writeResponseBody(builder, resp.RawBody)

	sCtx.WithNewAttachment("HTTP Response", allure.Text, builder.Bytes())
}

func (r *Reporter) writeRequestBasicInfo(builder *ReportBuilder, httpClient *client.Client, method, path string, pathParams, queryParams map[string]string) {
	builder.WriteLine("Method: %s", method)
	builder.WriteLine("Path: %s", path)

	if httpClient != nil {
		if eff, err := client.BuildEffectiveURL(httpClient.BaseURL, path, pathParams, queryParams); err == nil {
			builder.WriteLine("Effective URL: %s", eff)
		} else {
			builder.WriteLine("Effective URL: (failed to resolve: %v)", err)
		}
	}
}

func (r *Reporter) writeParams(builder *ReportBuilder, params map[string]string, title string) {
	if len(params) == 0 {
		return
	}
	builder.WriteSection(title)
	builder.WriteMap(params)
}

func (r *Reporter) writeRequestHeaders(builder *ReportBuilder, httpClient *client.Client, headers map[string]string) {
	if len(headers) == 0 {
		return
	}
	builder.WriteSection("Headers")
	for k, v := range headers {
		maskedValue := v
		if httpClient != nil && httpClient.ShouldMaskHeader(k) {
			maskedValue = maskHeaderValue(k, v)
		}
		builder.WriteKeyValue(k, maskedValue)
	}
}

func (r *Reporter) writeRequestBody(builder *ReportBuilder, body any, rawBody []byte, multipart *client.MultipartForm) {
	switch {
	case body != nil:
		builder.WriteSection("Body (json)")
		builder.WriteJSONOrError(body)

	case len(rawBody) > 0:
		builder.WriteSection("Body (raw)")
		builder.WriteTruncated(rawBody, 1000)

	case multipart != nil:
		builder.WriteSection("Body (multipart/form-data)")
		builder.WriteLine("Fields:")
		for k, v := range multipart.Fields {
			builder.WriteKeyValue(k, v)
		}
		if len(multipart.Files) > 0 {
			builder.WriteLine("Files:")
			for _, f := range multipart.Files {
				builder.WriteLine("  %s: %s (%d bytes)", f.FieldName, f.FileName, len(f.Content))
			}
		}

	default:
		builder.WriteSection("Body")
		builder.WriteLine("(empty)")
	}
}

func (r *Reporter) writeResponseStatus(builder *ReportBuilder, code int, duration interface{}) {
	builder.WriteLine("Status: %d", code)
	builder.WriteLine("Duration: %v", duration)
}

func (r *Reporter) writeResponseHeaders(builder *ReportBuilder, httpClient *client.Client, headers map[string][]string) {
	if len(headers) == 0 {
		return
	}
	builder.WriteSection("Headers")
	for k, values := range headers {
		for _, v := range values {
			maskedValue := v
			if httpClient != nil && httpClient.ShouldMaskHeader(k) {
				maskedValue = maskHeaderValue(k, v)
			}
			builder.WriteKeyValue(k, maskedValue)
		}
	}
}

func (r *Reporter) writeResponseError(builder *ReportBuilder, err *client.ErrorResponse) {
	if err == nil {
		return
	}
	builder.WriteSection("Error")
	if err.Message != "" {
		builder.WriteLine("  Message: %s", err.Message)
	}
	if len(err.Errors) > 0 {
		builder.WriteLine("  Errors:")
		for field, messages := range err.Errors {
			for _, msg := range messages {
				builder.WriteLine("    %s: %s", field, msg)
			}
		}
	}
	if err.Message == "" && len(err.Errors) == 0 {
		builder.WriteLine("  Body: %s", err.Body)
	}
}

func (r *Reporter) writeResponseBody(builder *ReportBuilder, rawBody []byte) {
	builder.WriteSection("Body")
	if len(rawBody) == 0 {
		builder.WriteLine("(empty)")
		return
	}

	var jsonData any
	if err := json.Unmarshal(rawBody, &jsonData); err == nil {
		if bodyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
			builder.buf.Write(bodyJSON)
			builder.buf.WriteString("\n")
			return
		}
	}

	builder.WriteTruncated(rawBody, 1000)
}

func maskHeaderValue(key, value string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "authorization" {
		parts := strings.SplitN(strings.TrimSpace(value), " ", 2)
		if len(parts) == 2 && parts[0] != "" {
			return parts[0] + " " + MaskValue
		}
	}
	return MaskValue
}
