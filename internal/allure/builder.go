package allure

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type ReportBuilder struct {
	buf bytes.Buffer
}

func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{}
}

func (b *ReportBuilder) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *ReportBuilder) String() string {
	return b.buf.String()
}

func (b *ReportBuilder) WriteLine(format string, args ...any) {
	b.buf.WriteString(fmt.Sprintf(format, args...))
	b.buf.WriteString("\n")
}

func (b *ReportBuilder) WriteJSON(v any) error {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b.buf.Write(bytes)
	b.buf.WriteString("\n")
	return nil
}

func (b *ReportBuilder) WriteJSONOrError(v any) {
	if err := b.WriteJSON(v); err != nil {
		b.WriteLine("(failed to marshal: %v)", err)
	}
}

func (b *ReportBuilder) WriteKeyValue(key string, value any) {
	b.WriteLine("  %s: %v", key, value)
}

func (b *ReportBuilder) WriteSection(title string) {
	b.WriteLine("\n%s:", title)
}

func (b *ReportBuilder) WriteTruncated(data []byte, maxSize int) {
	if len(data) <= maxSize {
		b.buf.Write(data)
		b.buf.WriteString("\n")
	} else {
		b.WriteLine("(truncated, %d bytes total)", len(data))
		b.buf.Write(data[:maxSize])
		b.buf.WriteString("\n...\n")
	}
}

func (b *ReportBuilder) WriteMap(m map[string]string) {
	for k, v := range m {
		b.WriteKeyValue(k, v)
	}
}

func (b *ReportBuilder) WriteHeader(title string) {
	line := "═══════════════════════════════════════════════════════════════"
	b.buf.WriteString(line + "\n")
	b.buf.WriteString(title + "\n")
	b.buf.WriteString(line + "\n\n")
}

func (b *ReportBuilder) WriteSectionHeader(title string) {
	line := "───────────────────────────────────────────────────────────────"
	b.buf.WriteString("\n" + title + "\n")
	b.buf.WriteString(line + "\n")
}
