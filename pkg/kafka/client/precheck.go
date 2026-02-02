package client

import (
	"github.com/gorelov-m-v/go-test-framework/internal/expect"
	"github.com/gorelov-m-v/go-test-framework/internal/polling"
)

func BytesPreCheckConfig() expect.PreCheckConfig[[]byte] {
	return expect.PreCheckConfig[[]byte]{
		IsNil:          func(b []byte) bool { return b == nil },
		EmptyBodyCheck: func(b []byte) bool { return len(b) == 0 },
	}
}

func BuildBytesPreCheck() func(error, []byte) (polling.CheckResult, bool) {
	return expect.BuildPreCheckWithBody(BytesPreCheckConfig())
}
