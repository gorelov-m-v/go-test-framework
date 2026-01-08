package com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game.payload;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
public record BrandPayload(
        String uuid,
        String alias,
        @JsonProperty("localized_names")
        Map<String, String> localizedNames,
        @JsonProperty("project_id")
        String projectId,
        String status
) {}
