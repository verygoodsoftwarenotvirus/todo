package postgres

import (
	"errors"
)

var (
	// ErrNilInputProvided indicates nil input was provided in an unacceptable context.
	ErrNilInputProvided = errors.New("nil input provided")

	// ErrNilTransactionProvided indicates nil transaction was provided in an unacceptable context.
	ErrNilTransactionProvided = errors.New("empty input provided")

	// ErrEmptyInputProvided indicates empty input was provided in an unacceptable context.
	ErrEmptyInputProvided = errors.New("empty input provided")

	// ErrInvalidIDProvided indicates a required ID was passed in as zero.
	ErrInvalidIDProvided = errors.New("required ID provided is zero")
)
