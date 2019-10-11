package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// argIsNotPointer checks an argument and returns whether or not it is a pointer
func argIsNotPointer(i interface{}) (notAPointer bool, err error) {
	if i == nil || reflect.TypeOf(i).Kind() != reflect.Ptr {
		return true, errors.New("value is not a pointer")
	}
	return false, nil
}

// argIsNotNil checks an argument and returns whether or not it is nil
func argIsNotNil(i interface{}) (isNil bool, err error) {
	if i == nil {
		return true, errors.New("value is nil")
	}
	return false, nil
}

// argIsNotPointerOrNil does what it says on the tin. This function
// is primarily useful for detecting if a destination value is valid
// before decoding an HTTP response, for instance
func argIsNotPointerOrNil(i interface{}) error {
	if nn, err := argIsNotNil(i); nn || err != nil {
		return err
	}

	if np, err := argIsNotPointer(i); np || err != nil {
		return err
	}

	return nil
}

// unmarshalBody takes an HTTP response and JSON decodes its
// body into a destination value. `dest` must be a non-nil
// pointer to an object. Ideally, response is also not nil.
// The error returned here should only ever be received in
// testing, and should never be encountered by an end-user.
func unmarshalBody(res *http.Response, dest interface{}) error {
	if err := argIsNotPointerOrNil(dest); err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := &models.ErrorResponse{}
		if err = json.Unmarshal(bodyBytes, &apiErr); err != nil {
			return fmt.Errorf("unmarshaling error: %w", err)
		}
		return apiErr
	}

	if err = json.Unmarshal(bodyBytes, &dest); err != nil {
		return fmt.Errorf("unmarshaling body: %w", err)
	}

	return nil
}

// createBodyFromStruct takes any value in and returns an io.Reader
// for placement within http.NewRequest's last argument.
func createBodyFromStruct(in interface{}) (io.Reader, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(out), nil
}
