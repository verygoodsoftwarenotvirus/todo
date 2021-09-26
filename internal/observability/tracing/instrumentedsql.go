package tracing

import (
	"context"

	"github.com/luna-duclos/instrumentedsql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
)

// NewInstrumentedSQLTracer wraps a Tracer for instrumentedsql.
func NewInstrumentedSQLTracer(name string) instrumentedsql.Tracer {
	return &instrumentedSQLTracerWrapper{tracer: NewTracer(name)}
}

var _ instrumentedsql.Tracer = (*instrumentedSQLTracerWrapper)(nil)

type instrumentedSQLTracerWrapper struct {
	tracer Tracer
}

// GetSpan wraps tracer.GetSpan.
func (t *instrumentedSQLTracerWrapper) GetSpan(ctx context.Context) instrumentedsql.Span {
	ctx, span := t.tracer.StartSpan(ctx)

	return &instrumentedSQLSpanWrapper{
		ctx:  ctx,
		span: span,
	}
}

// NewInstrumentedSQLLogger wraps a logging.Logger for instrumentedsql.
func NewInstrumentedSQLLogger(logger logging.Logger) instrumentedsql.Logger {
	return &instrumentedSQLLoggerWrapper{logger: logging.EnsureLogger(logger).WithName("sql")}
}

type instrumentedSQLLoggerWrapper struct {
	logger logging.Logger
}

func (w *instrumentedSQLLoggerWrapper) Log(_ context.Context, msg string, keyvals ...interface{}) {
	// this is noisy AF, log at your own peril
}
