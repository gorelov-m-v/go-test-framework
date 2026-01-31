package allure

import (
	"time"
)

func (r *Reporter) writeRedisTTL(builder *ReportBuilder, ttl time.Duration) {
	switch {
	case ttl >= 0:
		builder.WriteLine("TTL: %v", ttl)
	case ttl == -1:
		builder.WriteLine("TTL: no expiration")
	case ttl == -2:
		builder.WriteLine("TTL: key does not exist")
	}
}

func (r *Reporter) writeRedisValue(builder *ReportBuilder, value string) {
	if value == "" {
		return
	}

	builder.WriteSection("Value")

	if len(value) > 1000 {
		builder.WriteTruncated([]byte(value), 1000)
	} else {
		builder.WriteLine("%s", value)
	}
}
