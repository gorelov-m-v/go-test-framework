package allure

import (
	"fmt"
	"net/http"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

type PollingSummaryDTO struct {
	Attempts     int
	ElapsedTime  string
	Success      bool
	LastError    string
	FailedChecks []string
}

// ToPollingSummaryDTO converts polling.PollingSummary to DTO.
// Returns nil if no polling attempts were made.
func ToPollingSummaryDTO(ps polling.PollingSummary) *PollingSummaryDTO {
	if ps.Attempts == 0 {
		return nil
	}
	return &PollingSummaryDTO{
		Attempts:     ps.Attempts,
		ElapsedTime:  ps.ElapsedTime,
		Success:      ps.Success,
		LastError:    ps.LastError,
		FailedChecks: ps.FailedChecks,
	}
}

type HTTPReportDTO struct {
	Request  HTTPRequestDTO
	Response HTTPResponseDTO
	Polling  *PollingSummaryDTO
}

func (r *Reporter) AttachHTTPReport(sCtx provider.StepCtx, httpClient HTTPClientInfo, report HTTPReportDTO) {
	builder := NewReportBuilder()

	title := fmt.Sprintf("%s %s", report.Request.Method, report.Request.Path)
	if report.Response.StatusCode > 0 {
		title = fmt.Sprintf("%s %s → %d %s",
			report.Request.Method,
			report.Request.Path,
			report.Response.StatusCode,
			http.StatusText(report.Response.StatusCode))
	}
	builder.WriteHeader(title)

	r.writeHTTPRequestSection(builder, httpClient, report.Request)
	r.writeHTTPResponseSection(builder, httpClient, report.Response)

	if report.Polling != nil && report.Polling.Attempts > 0 {
		r.writePollingSection(builder, report.Polling)
	}

	sCtx.WithNewAttachment("HTTP Call", allure.Text, builder.Bytes())
}

type HTTPClientInfo interface {
	GetBaseURL() string
	ShouldMaskHeader(key string) bool
	BuildEffectiveURL(path string, pathParams, queryParams map[string]string) (string, error)
}

func (r *Reporter) writeHTTPRequestSection(builder *ReportBuilder, httpClient HTTPClientInfo, req HTTPRequestDTO) {
	builder.WriteSectionHeader("REQUEST")

	builder.WriteLine("Method: %s", req.Method)
	builder.WriteLine("Path: %s", req.Path)

	if httpClient != nil {
		if eff, err := httpClient.BuildEffectiveURL(req.Path, req.PathParams, req.QueryParams); err == nil {
			builder.WriteLine("URL: %s", eff)
		}
	}

	if len(req.PathParams) > 0 {
		builder.WriteSection("Path Params")
		builder.WriteMap(req.PathParams)
	}

	if len(req.QueryParams) > 0 {
		builder.WriteSection("Query Params")
		builder.WriteMap(req.QueryParams)
	}

	if len(req.Headers) > 0 {
		builder.WriteSection("Headers")
		for k, v := range req.Headers {
			maskedValue := v
			if httpClient != nil && httpClient.ShouldMaskHeader(k) {
				maskedValue = r.Config.MaskHeader(k, v)
			}
			builder.WriteKeyValue(k, maskedValue)
		}
	}

	r.writeRequestBody(builder, req.Body, req.RawBody, req.Multipart)
}

func (r *Reporter) writeHTTPResponseSection(builder *ReportBuilder, httpClient HTTPClientInfo, resp HTTPResponseDTO) {
	statusText := ""
	if resp.StatusCode > 0 {
		statusText = fmt.Sprintf(" [%d %s]", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	builder.WriteSectionHeader("RESPONSE" + statusText)

	if resp.NetworkError != "" {
		builder.WriteLine("Network Error: %s", resp.NetworkError)
		builder.WriteLine("Duration: %v", resp.Duration)
		return
	}

	builder.WriteLine("Status: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	builder.WriteLine("Duration: %v", resp.Duration)

	if len(resp.Headers) > 0 {
		builder.WriteSection("Headers")
		for k, values := range resp.Headers {
			for _, v := range values {
				maskedValue := v
				if httpClient != nil && httpClient.ShouldMaskHeader(k) {
					maskedValue = r.Config.MaskHeader(k, v)
				}
				builder.WriteKeyValue(k, maskedValue)
			}
		}
	}

	r.writeResponseError(builder, resp.Error)
	r.writeResponseBody(builder, resp.RawBody)
}

func (r *Reporter) writePollingSection(builder *ReportBuilder, polling *PollingSummaryDTO) {
	status := "success"
	if !polling.Success {
		status = "failed"
	}
	builder.WriteSectionHeader(fmt.Sprintf("POLLING (%s)", status))

	builder.WriteLine("Attempts: %d", polling.Attempts)
	builder.WriteLine("Elapsed: %s", polling.ElapsedTime)
	builder.WriteLine("Success: %t", polling.Success)

	if polling.LastError != "" {
		builder.WriteLine("Last Error: %s", polling.LastError)
	}

	if len(polling.FailedChecks) > 0 {
		builder.WriteSection("Failed Checks")
		for i, check := range polling.FailedChecks {
			builder.WriteLine("  [%d] %s", i+1, check)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// gRPC Report
// ═══════════════════════════════════════════════════════════════════════════

type GRPCReportDTO struct {
	Request  GRPCRequestDTO
	Response GRPCResponseDTO
	Polling  *PollingSummaryDTO
}

func (r *Reporter) AttachGRPCReport(sCtx provider.StepCtx, report GRPCReportDTO) {
	builder := NewReportBuilder()

	title := fmt.Sprintf("gRPC %s", report.Request.Method)
	if report.Response.Status != "" {
		title = fmt.Sprintf("gRPC %s → %s", report.Request.Method, report.Response.Status)
	}
	builder.WriteHeader(title)

	r.writeGRPCRequestSection(builder, report.Request)
	r.writeGRPCResponseSection(builder, report.Response)

	if report.Polling != nil && report.Polling.Attempts > 0 {
		r.writePollingSection(builder, report.Polling)
	}

	sCtx.WithNewAttachment("gRPC Call", allure.Text, builder.Bytes())
}

func (r *Reporter) writeGRPCRequestSection(builder *ReportBuilder, req GRPCRequestDTO) {
	builder.WriteSectionHeader("REQUEST")

	builder.WriteLine("Target: %s", req.Target)
	builder.WriteLine("Method: %s", req.Method)

	if len(req.Metadata) > 0 {
		builder.WriteSection("Metadata")
		for key, values := range req.Metadata {
			for _, value := range values {
				maskedValue := r.Config.MaskHeader(key, value)
				builder.WriteKeyValue(key, maskedValue)
			}
		}
	}

	r.writeBody(builder, req.Body)
}

func (r *Reporter) writeGRPCResponseSection(builder *ReportBuilder, resp GRPCResponseDTO) {
	statusText := ""
	if resp.Status != "" {
		statusText = fmt.Sprintf(" [%s]", resp.Status)
	}
	builder.WriteSectionHeader("RESPONSE" + statusText)

	builder.WriteLine("Status: %s (%d)", resp.Status, resp.StatusCode)
	builder.WriteLine("Duration: %v", resp.Duration)

	if resp.Error != nil {
		r.writeGRPCError(builder, resp.Error)
	}

	if len(resp.Metadata) > 0 {
		builder.WriteSection("Metadata")
		for key, values := range resp.Metadata {
			for _, value := range values {
				maskedValue := r.Config.MaskHeader(key, value)
				builder.WriteKeyValue(key, maskedValue)
			}
		}
	}

	r.writeBody(builder, resp.Body)
}

// ═══════════════════════════════════════════════════════════════════════════
// Redis Report
// ═══════════════════════════════════════════════════════════════════════════

type RedisReportDTO struct {
	Request RedisRequestDTO
	Result  RedisResultDTO
	Polling *PollingSummaryDTO
}

func (r *Reporter) AttachRedisReport(sCtx provider.StepCtx, report RedisReportDTO) {
	builder := NewReportBuilder()

	status := "Not Found"
	if report.Result.Exists {
		status = "Found"
	}
	title := fmt.Sprintf("Redis GET %s → %s", report.Request.Key, status)
	builder.WriteHeader(title)

	r.writeRedisRequestSection(builder, report.Request)
	r.writeRedisResultSection(builder, report.Result)

	if report.Polling != nil && report.Polling.Attempts > 0 {
		r.writePollingSection(builder, report.Polling)
	}

	sCtx.WithNewAttachment("Redis Query", allure.Text, builder.Bytes())
}

func (r *Reporter) writeRedisRequestSection(builder *ReportBuilder, req RedisRequestDTO) {
	builder.WriteSectionHeader("REQUEST")

	builder.WriteLine("Server: %s", req.Server)
	builder.WriteLine("Key: %s", req.Key)
}

func (r *Reporter) writeRedisResultSection(builder *ReportBuilder, result RedisResultDTO) {
	status := "Not Found"
	if result.Exists {
		status = "Found"
	}
	builder.WriteSectionHeader(fmt.Sprintf("RESULT [%s]", status))

	builder.WriteLine("Exists: %t", result.Exists)
	builder.WriteLine("Duration: %v", result.Duration)

	if result.Exists {
		r.writeRedisTTL(builder, result.TTL)
		r.writeRedisValue(builder, result.Value)
	}

	if result.Error != nil {
		builder.WriteSection("Error")
		builder.WriteKeyValue("Message", result.Error.Error())
	}
}

// ═══════════════════════════════════════════════════════════════════════════
// Kafka Report
// ═══════════════════════════════════════════════════════════════════════════

type KafkaReportDTO struct {
	Search  KafkaSearchDTO
	Result  KafkaResultDTO
	Polling *PollingSummaryDTO
}

func (r *Reporter) AttachKafkaReport(sCtx provider.StepCtx, report KafkaReportDTO) {
	builder := NewReportBuilder()

	status := "Not Found"
	if report.Result.Found {
		status = fmt.Sprintf("Found (%d)", report.Result.MatchCount)
	}
	title := fmt.Sprintf("Kafka %s → %s", report.Search.Topic, status)
	builder.WriteHeader(title)

	r.writeKafkaSearchSection(builder, report.Search)
	r.writeKafkaResultSection(builder, report.Result)

	if report.Polling != nil && report.Polling.Attempts > 0 {
		r.writePollingSection(builder, report.Polling)
	}

	sCtx.WithNewAttachment("Kafka Consume", allure.Text, builder.Bytes())
}

func (r *Reporter) writeKafkaSearchSection(builder *ReportBuilder, search KafkaSearchDTO) {
	builder.WriteSectionHeader("SEARCH")

	builder.WriteLine("Topic: %s", search.Topic)
	builder.WriteLine("Timeout: %v", search.Timeout)
	builder.WriteLine("Unique: %t", search.Unique)

	if len(search.Filters) > 0 {
		builder.WriteSection("Filters")
		builder.WriteMap(search.Filters)
	}
}

func (r *Reporter) writeKafkaResultSection(builder *ReportBuilder, result KafkaResultDTO) {
	status := "Not Found"
	if result.Found {
		status = "Found"
	}
	builder.WriteSectionHeader(fmt.Sprintf("RESULT [%s]", status))

	builder.WriteLine("Found: %t", result.Found)
	if result.Found {
		builder.WriteLine("Match Count: %d", result.MatchCount)
	}

	if result.Found && result.Message != nil {
		builder.WriteSection("Message")
		builder.WriteJSONOrError(result.Message)
	} else if result.Found && len(result.RawMessage) > 0 {
		builder.WriteSection("Message (raw)")
		builder.WriteTruncated(result.RawMessage, 2000)
	}
}
