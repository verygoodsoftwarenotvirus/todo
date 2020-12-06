package types

import (
	"errors"
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation"
)

var errInvalidType = errors.New("unexpected type received")

var _ validation.Rule = (*urlValidator)(nil)

type urlValidator struct{}

func (uv *urlValidator) Validate(value interface{}) error {
	raw, ok := value.(string)
	if !ok {
		return errInvalidType
	}

	if _, err := url.Parse(raw); err != nil {
		return err
	}

	return nil
}
