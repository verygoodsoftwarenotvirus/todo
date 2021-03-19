package httpclient

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/trace"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response.
	ErrNotFound = errors.New("404: not found")

	// ErrInvalidRequestInput is a handy error to return when we receive a 400 response.
	ErrInvalidRequestInput = errors.New("400: bad request")

	// ErrNoURLProvided is a handy error to return when we expect a *url.URL and don't receive one.
	ErrNoURLProvided = errors.New("no URL provided")

	// ErrBanned is a handy error to return when we receive a 401 response.
	ErrBanned = errors.New("403: banned")

	// ErrUnauthorized is a handy error to return when we receive a 401 response.
	ErrUnauthorized = errors.New("401: not authorized")

	// ErrInvalidTOTPToken is an error for when our TOTP validation request goes awry.
	ErrInvalidTOTPToken = errors.New("invalid TOTP token")

	// ErrNilInputProvided indicates nil input was provided in an unacceptable context.
	ErrNilInputProvided = errors.New("nil input provided")

	// ErrInvalidIDProvided indicates nil input was provided in an unacceptable context.
	ErrInvalidIDProvided = errors.New("required ID provided is zero")

	// ErrEmptyQueryProvided indicates the user provided an empty query.
	ErrEmptyQueryProvided = errors.New("query provided was empty")

	// ErrEmptyUsernameProvided indicates the user provided an empty username for search.
	ErrEmptyUsernameProvided = errors.New("empty username provided")

	// ErrCookieRequired indicates a cookie is required.
	ErrCookieRequired = errors.New("cookie required for request")

	// ErrNoCookiesReturned indicates nil input was provided in an unacceptable context.
	ErrNoCookiesReturned = errors.New("no cookies returned from request")
)

// prepareError standardizes our error handling by logging, tracing, and formatting an error consistently.
func prepareError(err error, logger logging.Logger, span trace.Span, descriptionFmt string, descriptionArgs ...interface{}) error {
	desc := fmt.Sprintf(descriptionFmt, descriptionArgs...)

	logging.EnsureLogger(logger).Error(err, desc)

	if err != nil {
		tracing.AttachErrorToSpan(span, err)
	}

	return fmt.Errorf("%s: %w", desc, err)
}
