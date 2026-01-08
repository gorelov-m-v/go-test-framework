package com.uplatform.wallet_tests.api.kafka.dto.core.gambling.v3.game.enums;

import com.fasterxml.jackson.annotation.JsonValue;

public enum GameV3EventType {

    CATEGORY("category"),
    BRAND("brand"),
    GAME("game");

    private final String value;

    GameV3EventType(String value) {
        this.value = value;
    }

    @JsonValue
    public String value() {
        return value;
    }
}
