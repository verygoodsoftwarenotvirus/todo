package observability

import (
	"fmt"
	"time"

	"go.opentelemetry.io/otel/trace"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
)

// PrepareError standardizes our error handling by logging, tracing, and formatting an error consistently.
func PrepareError(err error, logger logging.Logger, span tracing.Span, descriptionFmt string, descriptionArgs ...interface{}) error {
	if err == nil {
		return nil
	}

	desc := fmt.Sprintf(descriptionFmt, descriptionArgs...)
	logging.EnsureLogger(logger).Error(err, desc)

	if span != nil {
		tracing.AttachErrorToSpan(span, desc, err)
	}

	return fmt.Errorf("%s: %w", desc, err)
}

// AcknowledgeError standardizes our error handling by logging and tracing consistently.
func AcknowledgeError(err error, logger logging.Logger, span tracing.Span, descriptionFmt string, descriptionArgs ...interface{}) {
	desc := fmt.Sprintf(descriptionFmt, descriptionArgs...)
	logging.EnsureLogger(logger).Error(err, desc)
	if span != nil {
		tracing.AttachErrorToSpan(span, desc, err)
	}
}

// NoteEvent standardizes our logging and tracing notifications.
func NoteEvent(logger logging.Logger, span tracing.Span, descriptionFmt string, descriptionArgs ...interface{}) {
	desc := fmt.Sprintf(descriptionFmt, descriptionArgs...)
	logging.EnsureLogger(logger).Debug(desc)
	span.AddEvent(desc, trace.WithTimestamp(time.Now().UTC()))
}
