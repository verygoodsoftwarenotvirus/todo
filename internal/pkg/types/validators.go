package types

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	errInvalidType     = errors.New("unexpected type received")
	errDurationTooLong = errors.New("duration too long")
)

var _ validation.Rule = (*urlValidator)(nil)

type urlValidator struct{}

func (*urlValidator) Validate(value interface{}) error {
	raw, ok := value.(string)
	if !ok {
		return errInvalidType
	}

	if _, err := url.Parse(raw); err != nil {
		return err
	}

	return nil
}

var _ validation.Rule = (*stringDurationValidator)(nil)

type stringDurationValidator struct {
	maxDuration time.Duration
}

func (v *stringDurationValidator) Validate(value interface{}) error {
	raw, ok := value.(string)
	if !ok {
		return errInvalidType
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}

	if d > v.maxDuration {
		return fmt.Errorf("%w: %v", errDurationTooLong, d)
	}

	return nil
}
