package dsl

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
)

func attachFoundMessage(stepCtx provider.StepCtx, message interface{}) {
	if message == nil {
		return
	}

	jsonData, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		stepCtx.WithAttachments(
			allure.NewAttachment("Kafka Message Found", allure.Text, []byte(fmt.Sprintf("%+v", message))),
		)
		return
	}

	stepCtx.WithAttachments(
		allure.NewAttachment("Kafka Message Found", allure.JSON, jsonData),
	)
}

func attachSearchInfoByTopic(
	stepCtx provider.StepCtx,
	topicName string,
	filters map[string]string,
	timeout time.Duration,
	unique bool,
) {
	info := map[string]interface{}{
		"topic":   topicName,
		"filters": filters,
		"timeout": timeout.String(),
		"unique":  unique,
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return
	}

	stepCtx.WithAttachments(
		allure.NewAttachment("Kafka Search Info", allure.Text, jsonData),
	)
}

func attachNotFoundMessageByTopic(
	stepCtx provider.StepCtx,
	topicName string,
	filters map[string]string,
) {
	info := map[string]interface{}{
		"topic":   topicName,
		"filters": filters,
		"status":  "NOT_FOUND",
	}

	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return
	}

	stepCtx.WithAttachments(
		allure.NewAttachment("Kafka Message Not Found", allure.Text, jsonData),
	)
}
