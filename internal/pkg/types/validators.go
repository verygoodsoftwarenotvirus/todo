package types

import (
	"fmt"
	"net/url"

	"github.com/RussellLuo/validating/v2"
)

var _ validating.Validator = (*urlValidator)(nil)

type urlValidator struct{}

func (uv *urlValidator) Validate(field validating.Field) validating.Errors {
	if u, ok := field.ValuePtr.(string); ok {
		if _, err := url.Parse(u); err != nil {
			return validating.NewErrors(field.Name, "parse error", err.Error())
		}

		return nil
	}

	return validating.NewErrors(field.Name, "type error", "URL field is the wrong type")
}

var _ validating.Validator = (*userAccountStatusValidator)(nil)

type userAccountStatusValidator struct{}

func (slv *userAccountStatusValidator) Validate(field validating.Field) validating.Errors {
	if s, ok := field.ValuePtr.(userAccountStatus); ok && !IsValidAccountStatus(string(s)) {
		return validating.NewErrors(field.Name, "invalid value", fmt.Sprintf("%q is not a valid User account status", s))
	}

	return nil
}

var _ validating.Validator = (*minimumStringLengthValidator)(nil)

type minimumStringLengthValidator struct {
	minLength uint
}

func (slv *minimumStringLengthValidator) Validate(field validating.Field) validating.Errors {
	if s, ok := field.ValuePtr.(string); ok {
		if uint(len(s)) >= slv.minLength {
			return nil
		}

		return validating.NewErrors(field.Name, "invalid length", fmt.Sprintf("field should be at least %d characters long", slv.minLength))
	}

	return validating.NewErrors(field.Name, "type error", "string field is the wrong type")
}

var _ validating.Validator = (*exactStringLengthValidator)(nil)

type exactStringLengthValidator struct {
	length uint
}

func (slv *exactStringLengthValidator) Validate(field validating.Field) validating.Errors {
	if s, ok := field.ValuePtr.(string); ok {
		if uint(len(s)) == slv.length {
			return nil
		}

		return validating.NewErrors(field.Name, "invalid length", fmt.Sprintf("field should be at least %d characters long", slv.length))
	}

	return validating.NewErrors(field.Name, "type error", "string field is the wrong type")
}

var _ validating.Validator = (*minimumStringSliceLengthValidator)(nil)

type minimumStringSliceLengthValidator struct {
	minLength uint
}

func (slv *minimumStringSliceLengthValidator) Validate(field validating.Field) validating.Errors {
	if s, ok := field.ValuePtr.(*[]string); ok {
		if uint(len(*s)) >= slv.minLength {
			return nil
		}

		return validating.NewErrors(field.Name, "invalid length", fmt.Sprintf("field should be at least %d entries long", slv.minLength))
	}

	return validating.NewErrors(field.Name, "type error", "string slice field is the wrong type")
}
