package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"golang.org/x/oauth2/clientcredentials"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// MustParseURL parses a URL or otherwise panics.
func MustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return u
}

// BuildClientCredentialsConfig builds a clientcredentials.Config.
func BuildClientCredentialsConfig(u *url.URL, clientID, clientSecret string, scopes ...string) *clientcredentials.Config {
	conf := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		EndpointParams: url.Values{
			"client_id":     []string{clientID},
			"client_secret": []string{clientSecret},
		},
		TokenURL: tokenEndpoint(u).TokenURL,
	}

	return conf
}

// argIsNotPointer checks an argument and returns whether or not it is a pointer.
func argIsNotPointer(i interface{}) (notAPointer bool, err error) {
	if i == nil || reflect.TypeOf(i).Kind() != reflect.Ptr {
		return true, errors.New("value is not a pointer")
	}

	return false, nil
}

// argIsNotNil checks an argument and returns whether or not it is nil.
func argIsNotNil(i interface{}) (isNil bool, err error) {
	if i == nil {
		return true, ErrNilInputProvided
	}

	return false, nil
}

// argIsNotPointerOrNil does what it says on the tin. This function
// is primarily useful for detecting if a destination value is valid
// before decoding an HTTP response, for instance.
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
func (c *V1Client) unmarshalBody(ctx context.Context, res *http.Response, dest interface{}) error {
	_, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if err := argIsNotPointerOrNil(dest); err != nil {
		return err
	}

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		apiErr := &types.ErrorResponse{}

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
