package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func decodeResponse[TResp any](resp *http.Response, duration time.Duration) (*Response[TResp], error) {
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response[TResp]{
			StatusCode:   resp.StatusCode,
			Headers:      resp.Header,
			RawBody:      rawBody,
			Duration:     duration,
			NetworkError: fmt.Sprintf("failed to read response body: %v", err),
		}, err
	}

	result := &Response[TResp]{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		RawBody:    rawBody,
		Duration:   duration,
	}

	if resp.StatusCode >= 400 {
		result.Error = parseErrorResponse(rawBody, resp.StatusCode)
		return result, nil
	}

	if len(rawBody) > 0 && isJSONContentType(resp.Header.Get("Content-Type")) {
		var body TResp
		if err := json.Unmarshal(rawBody, &body); err != nil {
			result.NetworkError = fmt.Sprintf("failed to decode response body: %v", err)
		} else {
			result.Body = body
		}
	}

	return result, nil
}

func parseErrorResponse(rawBody []byte, statusCode int) *ErrorResponse {
	errResp := &ErrorResponse{
		Body:       string(rawBody),
		StatusCode: statusCode,
	}

	if len(rawBody) == 0 {
		return errResp
	}

	var envelope struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}

	if err := json.Unmarshal(rawBody, &envelope); err == nil {
		errResp.Message = envelope.Message
		errResp.Errors = envelope.Errors
	}

	return errResp
}

func isJSONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "json")
}
