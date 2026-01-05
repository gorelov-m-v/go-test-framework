package allure

import "strings"

const MaskValue = "***MASKED***"

type MaskingConfig struct {
	SensitiveHeaders []string
	SensitiveFields  []string
	MaskValue        string
}

func DefaultConfig() MaskingConfig {
	return MaskingConfig{
		SensitiveHeaders: []string{
			"authorization",
			"cookie",
			"set-cookie",
			"x-api-key",
			"api-key",
			"x-token",
			"token",
			"x-secret",
		},
		SensitiveFields: []string{"password", "secret", "token"},
		MaskValue:       "***MASKED***",
	}
}

func (c MaskingConfig) ShouldMaskField(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, field := range c.SensitiveFields {
		if strings.Contains(key, strings.ToLower(field)) {
			return true
		}
	}
	return false
}

func (c MaskingConfig) ShouldMaskHeader(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, header := range c.SensitiveHeaders {
		if strings.ToLower(header) == key {
			return true
		}
	}
	return false
}

func (c MaskingConfig) MaskHeader(key, value string) string {
	if !c.ShouldMaskHeader(key) {
		return value
	}

	key = strings.ToLower(strings.TrimSpace(key))
	if key == "authorization" {
		parts := strings.SplitN(strings.TrimSpace(value), " ", 2)
		if len(parts) == 2 && parts[0] != "" {
			return parts[0] + " " + c.MaskValue
		}
	}

	return c.MaskValue
}

func (c MaskingConfig) ShouldMaskValue(value string) bool {
	lower := strings.ToLower(value)
	for _, field := range c.SensitiveFields {
		if strings.Contains(lower, strings.ToLower(field)) {
			return true
		}
	}
	return false
}
