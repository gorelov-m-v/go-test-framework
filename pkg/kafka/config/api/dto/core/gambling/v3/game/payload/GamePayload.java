package com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game.payload;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true)
public record GamePayload(
        String uuid,
        @JsonProperty("brand_uuids")
        List<String> brandUuids
) {}
