package allure

import (
	"google.golang.org/grpc/status"
)

func (r *Reporter) writeGRPCError(builder *ReportBuilder, err error) {
	builder.WriteSection("Error")

	st, ok := status.FromError(err)
	if ok {
		builder.WriteKeyValue("Code", st.Code().String())
		builder.WriteKeyValue("Message", st.Message())
	} else {
		builder.WriteKeyValue("Message", err.Error())
	}
}

func (r *Reporter) writeBody(builder *ReportBuilder, body any) {
	if body == nil {
		return
	}

	builder.WriteSection("Body")
	builder.WriteJSONOrError(body)
}
