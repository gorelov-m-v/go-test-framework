package com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game.payload.BrandPayload;
import com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game.payload.CategoryPayload;
import com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game.payload.GamePayload;

@JsonIgnoreProperties(ignoreUnknown = true)
public record GameV3Message(
        MessageEnvelope message,
        CategoryPayload category,
        BrandPayload brand,
        GamePayload game
) {
    @JsonIgnoreProperties(ignoreUnknown = true)
    public record MessageEnvelope(
            @JsonProperty("eventType")
            String eventType
    ) {}
}
