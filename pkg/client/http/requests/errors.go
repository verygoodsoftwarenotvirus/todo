package requests

import (
	"errors"
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
