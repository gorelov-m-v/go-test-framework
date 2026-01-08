package com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game.payload;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
public record CategoryPayload(
        String uuid,
        String name,
        @JsonProperty("localized_names")
        Map<String, String> localizedNames,
        String type,
        String status,
        @JsonProperty("parent_uuid")
        String parentUuid
) {}
