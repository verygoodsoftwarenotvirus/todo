package observability

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"go.opentelemetry.io/otel/trace"
)

// PrepareError standardizes our error handling by logging, tracing, and formatting an error consistently.
func PrepareError(err error, logger logging.Logger, span trace.Span, descriptionFmt string, descriptionArgs ...interface{}) error {
	desc := fmt.Sprintf(descriptionFmt, descriptionArgs...)

	logging.EnsureLogger(logger).Error(err, desc)
	tracing.AttachErrorToSpan(span, desc, err)

	return fmt.Errorf("%s: %w", desc, err)
}

// AcknowledgeError standardizes our error handling by logging and tracing consistently.
func AcknowledgeError(err error, logger logging.Logger, span trace.Span, descriptionFmt string, descriptionArgs ...interface{}) {
	if err != nil {
		desc := fmt.Sprintf(descriptionFmt, descriptionArgs...)
		logging.EnsureLogger(logger).Error(err, fmt.Sprintf(descriptionFmt, descriptionArgs...))
		tracing.AttachErrorToSpan(span, desc, err)
	}
}
