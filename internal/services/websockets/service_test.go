package websockets

import (
	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

func buildTestService() *service {
	return &service{
		logger:         logging.NewNoopLogger(),
		encoderDecoder: mockencoding.NewMockEncoderDecoder(),
		tracer:         tracing.NewTracer("test"),
	}
}
