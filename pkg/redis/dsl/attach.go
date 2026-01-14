package dsl

import (
	"encoding/json"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"

	"github.com/gorelov-m-v/go-test-framework/pkg/redis/client"
)

type redisRequestDTO struct {
	Server string `json:"server"`
	Key    string `json:"key"`
}

type redisResultDTO struct {
	Key      string `json:"key"`
	Exists   bool   `json:"exists"`
	Value    string `json:"value,omitempty"`
	TTL      string `json:"ttl,omitempty"`
	Duration string `json:"duration"`
	Error    string `json:"error,omitempty"`
}

func attachRequest(stepCtx provider.StepCtx, q *Query) {
	dto := redisRequestDTO{
		Server: q.client.Addr(),
		Key:    q.key,
	}

	jsonBytes, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		stepCtx.Logf("Failed to marshal Redis request: %v", err)
		return
	}

	stepCtx.WithAttachments(allure.NewAttachment("Redis Query", allure.JSON, jsonBytes))
}

func attachResult(stepCtx provider.StepCtx, result *client.Result) {
	if result == nil {
		return
	}

	dto := redisResultDTO{
		Key:      result.Key,
		Exists:   result.Exists,
		Duration: result.Duration.String(),
	}

	if result.Exists {
		dto.Value = result.Value
		if result.TTL >= 0 {
			dto.TTL = result.TTL.String()
		} else if result.TTL == -1 {
			dto.TTL = "no expiration"
		}
	}

	if result.Error != nil {
		dto.Error = result.Error.Error()
	}

	jsonBytes, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		stepCtx.Logf("Failed to marshal Redis result: %v", err)
		return
	}

	attachmentName := "Redis Result"
	if result.Exists {
		attachmentName = "Redis Result [Found]"
	} else {
		attachmentName = "Redis Result [Not Found]"
	}

	stepCtx.WithAttachments(allure.NewAttachment(attachmentName, allure.JSON, jsonBytes))
}
