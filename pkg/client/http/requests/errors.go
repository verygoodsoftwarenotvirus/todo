package requests

import (
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"

	"go.opentelemetry.io/otel/trace"
)

var (
	// ErrNoURLProvided is a handy error to return when we expect a *url.URL and don't receive one.
	ErrNoURLProvided = errors.New("no URL provided")

	// ErrNilInputProvided indicates nil input was provided in an unacceptable context.
	ErrNilInputProvided = errors.New("nil input provided")

	// ErrInvalidIDProvided indicates nil input was provided in an unacceptable context.
	ErrInvalidIDProvided = errors.New("required ID provided is zero")

	// ErrEmptyUsernameProvided indicates the user provided an empty username for search.
	ErrEmptyUsernameProvided = errors.New("empty username provided")

	// ErrCookieRequired indicates a cookie is required.
	ErrCookieRequired = errors.New("cookie required for request")
)

// prepareError standardizes our error handling by logging, tracing, and formatting an error consistently.
func prepareError(err error, logger logging.Logger, span trace.Span, description string) error {
	logging.EnsureLogger(logger).Error(err, description)

	if err != nil {
		tracing.AttachErrorToSpan(span, err)
	}

	return fmt.Errorf("%s: %w", description, err)
}
