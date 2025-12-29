package httpclient

import (
	"strings"
)

func (c *Client) SanitizeHeaders(headers map[string]string) map[string]string {
	sanitized := make(map[string]string, len(headers))

	for key, value := range headers {
		if c.IsSecretHeader(key) {
			sanitized[key] = maskHeaderValue(key, value)
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}

func maskHeaderValue(key, value string) string {
	k := strings.ToLower(strings.TrimSpace(key))

	if k == "authorization" {
		parts := strings.SplitN(strings.TrimSpace(value), " ", 2)
		if len(parts) == 2 && parts[0] != "" {
			return parts[0] + " ***"
		}
		return "***"
	}

	return "***"
}
