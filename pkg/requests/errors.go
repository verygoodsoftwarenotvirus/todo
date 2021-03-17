package requests

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrNotFound is a handy error to return when we receive a 404 response.
	ErrNotFound = fmt.Errorf("%d: not found", http.StatusNotFound)

	// ErrInvalidRequestInput is a handy error to return when we receive a 400 response.
	ErrInvalidRequestInput = fmt.Errorf("%d: bad request", http.StatusBadRequest)

	// ErrNoURLProvided is a handy error to return when we expect a *url.URL and don't receive one.
	ErrNoURLProvided = errors.New("no URL provided")

	// ErrBanned is a handy error to return when we receive a 401 response.
	ErrBanned = fmt.Errorf("%d: banned", http.StatusForbidden)

	// ErrUnauthorized is a handy error to return when we receive a 401 response.
	ErrUnauthorized = fmt.Errorf("%d: not authorized", http.StatusUnauthorized)

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
