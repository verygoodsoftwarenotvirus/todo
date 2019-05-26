package client

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/pkg/errors"
)

////////////////////////////////////////////////////////
//                                                    //
//                 Helper Functions                   //
//                                                    //
////////////////////////////////////////////////////////

func argIsNotPointerOrNil(i interface{}) error {
	if nn, err := argIsNotNil(i); nn || err != nil {
		return err
	}

	if np, err := argIsNotPointer(i); np || err != nil {
		return err
	}

	return nil
}

// argIsNotPointer checks an argument and returns whether or not it is a pointer
func argIsNotPointer(i interface{}) (notAPointer bool, err error) {
	if i == nil || reflect.TypeOf(i).Kind() != reflect.Ptr {
		return true, errors.New("value is not a pointer")
	}
	return
}

// argIsNotNil checks an argument and returns whether or not it is nil
func argIsNotNil(i interface{}) (isNil bool, err error) {
	if i == nil {
		return true, errors.New("value is nil")
	}
	return
}

func unmarshalBody(res *http.Response, dest interface{}) error {
	// These paths should only ever be reached in tests, and should never be encountered by an end user.
	if err := argIsNotPointerOrNil(dest); err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := &models.ErrorResponse{}
		// eating this error because it would have been caught above
		if err = json.Unmarshal(bodyBytes, &apiErr); err != nil {
			return errors.Wrap(err, "unmarshaling error")
		}
		return apiErr
	}

	s := string(bodyBytes)
	_ = s

	if err = json.Unmarshal(bodyBytes, &dest); err != nil {
		return errors.Wrap(err, "unmarshaling body")
	}

	return nil
}

func createBodyFromStruct(in interface{}) (io.Reader, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(out), nil
}
