package requests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
)

// mustParseURL parses a url or otherwise panics.
func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}

	return u
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
