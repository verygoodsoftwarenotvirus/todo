package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleURI       = "https://todo.verygoodsoftwarenotvirus.ru"
	asciiControlChar = string(byte(127))
)

// begin test helpers

type (
	argleBargle struct {
		Name string
	}

	valuer map[string][]string
)

func (v valuer) ToValues() url.Values {
	return url.Values(v)
}

type requestSpec struct {
	path              string
	method            string
	query             string
	pathArgs          []interface{}
	bodyShouldBeEmpty bool
}

func newRequestSpec(bodyShouldBeEmpty bool, method, query, path string, pathArgs ...interface{}) *requestSpec {
	return &requestSpec{
		path:              path,
		pathArgs:          pathArgs,
		method:            method,
		query:             query,
		bodyShouldBeEmpty: bodyShouldBeEmpty,
	}
}

func assertErrorMatches(t *testing.T, err1, err2 error) {
	t.Helper()

	assert.True(t, errors.Is(err1, err2))
}

func assertRequestQuality(t *testing.T, req *http.Request, spec *requestSpec) {
	t.Helper()

	expectedPath := fmt.Sprintf(spec.path, spec.pathArgs...)

	require.NotNil(t, req, "provided req must not be nil")
	require.NotNil(t, spec, "provided spec must not be nil")

	bodyBytes, err := httputil.DumpRequest(req, true)
	require.NotEmpty(t, bodyBytes)
	require.NoError(t, err)

	if spec.bodyShouldBeEmpty {
		bodyLines := strings.Split(string(bodyBytes), "\n")
		require.NotEmpty(t, bodyLines)
		assert.Empty(t, bodyLines[len(bodyLines)-1])
	}

	assert.Equal(t, spec.query, req.URL.Query().Encode(), "expected query to be %q, but was %q instead", spec.query, req.URL.Query().Encode())
	assert.Equal(t, expectedPath, req.URL.Path, "expected path to be %q, but was %q instead", expectedPath, req.URL.Path)
	assert.Equal(t, spec.method, req.Method, "expected method to be %q, but was %q instead", spec.method, req.Method)
}

// createBodyFromStruct takes any value in and returns an io.Reader for placement within http.NewRequest's last argument.
func createBodyFromStruct(t *testing.T, in interface{}) io.Reader {
	t.Helper()

	out, err := json.Marshal(in)
	require.NoError(t, err)

	return bytes.NewReader(out)
}

func buildTestClient(t *testing.T, ts *httptest.Server) *Client {
	t.Helper()

	require.NotNil(t, ts)

	client, err := NewClient(
		mustParseURL(""),
		UsingLogger(logging.NewNonOperationalLogger()),
	)
	require.NoError(t, err)
	require.NotNil(t, client)

	require.NoError(t, client.requestBuilder.SetURL(mustParseURL(ts.URL)))
	client.unauthenticatedClient = ts.Client()
	client.authedClient = ts.Client()

	return client
}

func buildSimpleTestClient(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()

	ts := httptest.NewTLSServer(nil)

	return buildTestClient(t, ts), ts
}

func buildTestClientWithInvalidURL(t *testing.T) *Client {
	t.Helper()

	l := logging.NewNonOperationalLogger()
	u := mustParseURL("https://verygoodsoftwarenotvirus.ru")
	u.Scheme = fmt.Sprintf(`%s://`, asciiControlChar)

	c, err := NewClient(u, UsingLogger(l), UsingDebug(true))
	require.NotNil(t, c)
	require.NoError(t, err)

	return c
}

func buildTestClientWithStatusCodeResponse(t *testing.T, spec *requestSpec, code int) (*Client, *httptest.Server) {
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			t.Helper()

			assertRequestQuality(t, req, spec)

			res.WriteHeader(code)
		},
	))

	return buildTestClient(t, ts), ts
}

func buildTestClientWithInvalidResponse(t *testing.T, spec *requestSpec) *Client {
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			t.Helper()

			assertRequestQuality(t, req, spec)

			require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
		},
	))

	return buildTestClient(t, ts)
}

func buildTestClientWithJSONResponse(t *testing.T, spec *requestSpec, outputBody interface{}) (*Client, *httptest.Server) {
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			t.Helper()

			assertRequestQuality(t, req, spec)

			assert.NoError(t, json.NewEncoder(res).Encode(outputBody))
		},
	))

	return buildTestClient(t, ts), ts
}

func buildTestClientWithRequestBodyValidation(t *testing.T, spec *requestSpec, inputBody, expectedInput, outputBody interface{}) *Client {
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			t.Helper()

			assertRequestQuality(t, req, spec)

			require.NoError(t, json.NewDecoder(req.Body).Decode(&inputBody))
			assert.Equal(t, expectedInput, inputBody)
			require.NoError(t, json.NewEncoder(res).Encode(outputBody))
		},
	))

	return buildTestClient(t, ts)
}

func buildTestClientThatWaitsTooLong(t *testing.T, spec *requestSpec) (*Client, *httptest.Server) {
	ts := httptest.NewTLSServer(http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {
			t.Helper()

			assertRequestQuality(t, req, spec)

			time.Sleep(24 * time.Hour)
		},
	))

	c := buildTestClient(t, ts)

	require.NoError(t, c.SetOptions(UsingTimeout(time.Millisecond)))

	return c, ts
}

// end test helpers
