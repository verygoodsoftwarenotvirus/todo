package tracing

import (
	"context"

	"github.com/luna-duclos/instrumentedsql"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"
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

var _ instrumentedsql.Span = (*instrumentedSQLSpanWrapper)(nil)

type instrumentedSQLSpanWrapper struct {
	ctx  context.Context
	span trace.Span
}

func (w *instrumentedSQLSpanWrapper) NewChild(s string) instrumentedsql.Span {
	w.ctx, w.span = w.span.Tracer().Start(w.ctx, s)

	return w
}

func (w *instrumentedSQLSpanWrapper) SetLabel(k, v string) {
	w.span.SetAttributes(label.String(k, v))
}

func (w *instrumentedSQLSpanWrapper) SetError(err error) {
	w.span.RecordError(err)
}

func (w *instrumentedSQLSpanWrapper) Finish() {
	w.span.End()
}

// NewInstrumentedSQLLogger wraps a logging.Logger for instrumentedsql.
func NewInstrumentedSQLLogger(logger logging.Logger) instrumentedsql.Logger {
	return &instrumentedSQLLoggerWrapper{logger: logger}
}

type instrumentedSQLLoggerWrapper struct {
	logger logging.Logger
}

func (w *instrumentedSQLLoggerWrapper) Log(_ context.Context, msg string, keyvals ...interface{}) {
	// this is noisy AF, log at your own peril
}
