package types

import (
	"fmt"
	"net/url"

	v "github.com/RussellLuo/validating/v2"
)

var _ v.Validator = (*urlValidator)(nil)

type urlValidator struct{}

func (uv *urlValidator) Validate(field v.Field) v.Errors {
	if u, ok := field.ValuePtr.(string); ok {
		if _, err := url.Parse(u); err != nil {
			return v.NewErrors(field.Name, "parse error", err.Error())
		}

		return nil
	}

	return v.NewErrors(field.Name, "type error", "URL field is the wrong type")
}

var _ v.Validator = (*userAccountStatusValidator)(nil)

type userAccountStatusValidator struct{}

func (slv *userAccountStatusValidator) Validate(field v.Field) v.Errors {
	if s, ok := field.ValuePtr.(userAccountStatus); ok && !IsValidAccountStatus(string(s)) {
		return v.NewErrors(field.Name, "invalid value", fmt.Sprintf("%q is not a valid user account status", s))
	}

	return nil
}

var _ v.Validator = (*minimumStringLengthValidator)(nil)

type minimumStringLengthValidator struct {
	minLength int
}

func (slv *minimumStringLengthValidator) Validate(field v.Field) v.Errors {
	if s, ok := field.ValuePtr.(string); ok {
		if len(s) >= slv.minLength {
			return nil
		}

		return v.NewErrors(field.Name, "invalid length", fmt.Sprintf("field should be at least %d characters long", slv.minLength))
	}

	return v.NewErrors(field.Name, "type error", "string field is the wrong type")
}

var _ v.Validator = (*minimumStringSliceLengthValidator)(nil)

type minimumStringSliceLengthValidator struct {
	minLength int
}

func (slv *minimumStringSliceLengthValidator) Validate(field v.Field) v.Errors {
	if s, ok := field.ValuePtr.(*[]string); ok {
		if len(*s) >= slv.minLength {
			return nil
		}

		return v.NewErrors(field.Name, "invalid length", fmt.Sprintf("field should be at least %d entries long", slv.minLength))
	}

	return v.NewErrors(field.Name, "type error", "string slice field is the wrong type")
}
