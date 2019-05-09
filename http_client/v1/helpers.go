package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// Error is an error wrapper we can expose to the end user
type Error struct {
	Err     error
	FromAPI *models.ErrorResponse
}

func (ce *Error) Error() string {
	if ce.Err != nil {
		return ce.Err.Error()
	} else if ce.FromAPI != nil {
		return ce.FromAPI.Error()
	}

	return ""
}

////////////////////////////////////////////////////////
//                                                    //
//                 Helper Functions                   //
//                                                    //
////////////////////////////////////////////////////////

//func mapToQueryValues(in map[string]string) url.Values {
//	out := url.Values{}
//	for k, v := range in {
//		out.Set(k, v)
//	}
//	return out
//}

func argIsNotPointerOrNil(i interface{}) error {
	if nn, err := argIsNotNil(i); nn || err != nil {
		return err
	}
	if np, err := argIsNotPointer(i); np || err != nil {
		return err
	}
	return nil
}

// argIsNotPointer looks like a normal function, but its error value is actually an error you can wrap around
func argIsNotPointer(i interface{}) (np bool, err error) {
	if reflect.TypeOf(i).Kind() != reflect.Ptr {
		return true, errors.New("value is not a pointer")
	}
	return
}

// argIsNotNil looks like a normal function, but its error value is actually an error you can wrap around
func argIsNotNil(i interface{}) (nn bool, err error) {
	if i == nil {
		return true, errors.New("value is nil")
	}
	return
}

func unmarshalBody(res *http.Response, dest interface{}) error {
	ce := &Error{}

	// These paths should only ever be reached in tests, and should never be encountered by an end user.
	if err := argIsNotPointerOrNil(dest); err != nil {
		ce.Err = err
		return ce
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ce.Err = err
		return ce
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		apiErr := &models.ErrorResponse{}
		// eating this error because it would have been caught above
		err = json.Unmarshal(bodyBytes, &apiErr)
		if err != nil {
			return &Error{Err: err}
		}
		return &Error{FromAPI: apiErr}
	}

	err = json.Unmarshal(bodyBytes, &dest)
	if err != nil {
		return &Error{Err: err}
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
