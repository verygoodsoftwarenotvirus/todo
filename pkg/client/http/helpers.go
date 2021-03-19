package http

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/panicking"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var panicker = panicking.NewProductionPanicker()

// mustParseURL parses a url or otherwise panics.
func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panicker.Panic(err)
	}

	return u
}

// errorFromResponse returns library errors according to a response's status code.
func errorFromResponse(res *http.Response) error {
	if res == nil {
		return errors.New("nil response")
	}

	switch res.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusBadRequest:
		return ErrInvalidRequestInput
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrBanned
	case http.StatusInternalServerError:
		return ErrInternalServerError
	default:
		return nil
	}
}

// argIsNotPointer checks an argument and returns whether or not it is a pointer.
func argIsNotPointer(i interface{}) (bool, error) {
	if i == nil || reflect.TypeOf(i).Kind() != reflect.Ptr {
		return true, errors.New("value is not a pointer")
	}

	return false, nil
}

// argIsNotNil checks an argument and returns whether or not it is nil.
func argIsNotNil(i interface{}) (bool, error) {
	if i == nil {
		return true, ErrNilInputProvided
	}

	return false, nil
}

// argIsNotPointerOrNil does what it says on the tin. This function is primarily useful for detecting
// if a destination value is valid before decoding an HTTP response, for instance.
func argIsNotPointerOrNil(i interface{}) error {
	if nn, err := argIsNotNil(i); nn || err != nil {
		return err
	}

	if np, err := argIsNotPointer(i); np || err != nil {
		return err
	}

	return nil
}

// unmarshalBody takes an HTTP response and JSON decodes its body into a destination value. The error returned here
// should only ever be received in testing, and should never be encountered by an end-user.
func (c *Client) unmarshalBody(ctx context.Context, res *http.Response, dest interface{}) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithResponse(res)

	if err := argIsNotPointerOrNil(dest); err != nil {
		return prepareError(err, logger, span, "nil marshal target")
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return prepareError(err, logger, span, "unmarshalling error response")
	}

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := &types.ErrorResponse{
			Code: res.StatusCode,
		}

		if err = json.Unmarshal(bodyBytes, &apiErr); err != nil {
			logger.Error(err, "unmarshalling error response")
			tracing.AttachErrorToSpan(span, err)
		}

		return apiErr
	}

	if err = json.Unmarshal(bodyBytes, &dest); err != nil {
		return prepareError(err, logger, span, "unmarshalling response body")
	}

	return nil
}
