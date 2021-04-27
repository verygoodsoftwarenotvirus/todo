package httpclient

import (
	"errors"
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response.
	ErrNotFound = errors.New("404: not found")

	// ErrInvalidRequestInput is a handy error to return when we receive a 400 response.
	ErrInvalidRequestInput = errors.New("400: bad request")

	// ErrBanned is a handy error to return when we receive a 403 response.
	ErrBanned = errors.New("403: banned")

	// ErrInternalServerError is a handy error to return when we receive a 500 response.
	ErrInternalServerError = errors.New("500: internal server error")

	// ErrUnauthorized is a handy error to return when we receive a 401 response.
	ErrUnauthorized = errors.New("401: not authorized")

	// ErrNoURLProvided is a handy error to return when we expect a *url.URL and don't receive one.
	ErrNoURLProvided = errors.New("no URL provided")

	// ErrInvalidTOTPToken is an error for when our TOTP validation request goes awry.
	ErrInvalidTOTPToken = errors.New("invalid TOTP token")

	// ErrNilInputProvided indicates nil input was provided in an unacceptable context.
	ErrNilInputProvided = errors.New("nil input provided")

	// ErrEmptyInputProvided indicates empty input was provided in an unacceptable context.
	ErrEmptyInputProvided = errors.New("empty input provided")

	// ErrInvalidIDProvided indicates a required ID was passed in as zero.
	ErrInvalidIDProvided = errors.New("required ID provided is zero")

	// ErrEmptyQueryProvided indicates the user provided an empty query.
	ErrEmptyQueryProvided = errors.New("query provided was empty")

	// ErrEmptyUsernameProvided indicates the user provided an empty username for search.
	ErrEmptyUsernameProvided = errors.New("empty username provided")

	// ErrCookieRequired indicates a cookie is required.
	ErrCookieRequired = errors.New("cookie required for request")

	// ErrNoCookiesReturned indicates nil input was provided in an unacceptable context.
	ErrNoCookiesReturned = errors.New("no cookies returned from request")

	// ErrInvalidAvatarSize indicates an invalid avatar was provided.
	ErrInvalidAvatarSize = errors.New("invalid avatar size")

	// ErrInvalidImageExtension indicates an invalid image extension was provided.
	ErrInvalidImageExtension = errors.New("invalid image extension")

	// ErrNilResponse indicates we received a nil response.
	ErrNilResponse = errors.New("nil response")

	// ErrArgumentIsNotPointer indicates we received a non-pointer interface argument.
	ErrArgumentIsNotPointer = errors.New("value is not a pointer")
)
