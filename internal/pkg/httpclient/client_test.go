package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

const (
	exampleURI       = "https://todo.verygoodsoftwarenotvirus.ru"
	asciiControlChar = string(byte(127))
)

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
	pathArgs          []interface{}
	method            string
	query             string
	bodyShouldBeEmpty bool
}

func newRequestSpec(bodyShouldBeEmpty bool, method, query, path string, pathArgs ...interface{}) requestSpec {
	return requestSpec{
		path:              path,
		pathArgs:          pathArgs,
		method:            method,
		query:             query,
		bodyShouldBeEmpty: bodyShouldBeEmpty,
	}
}

func assertRequestQuality(t *testing.T, req *http.Request, spec requestSpec) {
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

// begin helper funcs

func mustParseURL(uri string) *url.URL {
	u, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	return u
}

func buildTestClient(t *testing.T, ts *httptest.Server) *V1Client {
	t.Helper()

	l := noop.NewLogger()
	u := mustParseURL(ts.URL)
	c := ts.Client()

	return &V1Client{
		URL:          u,
		plainClient:  c,
		logger:       l,
		Debug:        true,
		authedClient: c,
	}
}

func buildTestClientWithInvalidURL(t *testing.T) *V1Client {
	t.Helper()

	l := noop.NewLogger()
	u := mustParseURL("https://verygoodsoftwarenotvirus.ru")
	u.Scheme = fmt.Sprintf(`%s://`, asciiControlChar)

	return &V1Client{
		URL:          u,
		plainClient:  http.DefaultClient,
		logger:       l,
		Debug:        true,
		authedClient: http.DefaultClient,
	}
}

// end helper funcs

func TestV1Client_EnableAdminMode(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewTLSServer(nil)

		buildTestClient(t, ts).EnableAdminMode()
	})
}

func TestV1Client_AuthenticatedClient(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual := c.AuthenticatedClient()

		assert.Equal(t, ts.Client(), actual, "AuthenticatedClient should return the assigned authedClient")
	})
}

func TestV1Client_PlainClient(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual := c.PlainClient()

		assert.Equal(t, ts.Client(), actual, "PlainClient should return the assigned plainClient")
	})
}

func TestV1Client_TokenSource(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		c, err := NewClient(
			ctx,
			"",
			"",
			mustParseURL(exampleURI),
			noop.NewLogger(),
			ts.Client(),
			[]string{"*"},
			false,
		)
		require.NoError(t, err)

		actual := c.TokenSource()

		assert.NotNil(t, actual)
	})
}

func TestNewClient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		c, err := NewClient(
			ctx,
			"",
			"",
			mustParseURL(exampleURI),
			noop.NewLogger(),
			ts.Client(),
			[]string{"*"},
			false,
		)

		require.NotNil(t, c)
		require.NoError(t, err)
	})

	T.Run("with client but invalid timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, err := NewClient(
			ctx,
			"",
			"",
			mustParseURL(exampleURI),
			noop.NewLogger(),
			&http.Client{
				Timeout: 0,
			},
			[]string{"*"},
			true,
		)

		require.NotNil(t, c)
		require.NoError(t, err)
		assert.Equal(t, c.plainClient.Timeout, defaultTimeout, "NewClient should set the default timeout")
	})
}

func TestNewSimpleClient(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, err := NewSimpleClient(
			ctx,
			mustParseURL(exampleURI),
			true,
		)
		assert.NotNil(t, c)
		assert.NoError(t, err)
	})
}

func TestV1Client_CloseRequestBody(T *testing.T) {
	T.Parallel()

	T.Run("with error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		rc := newMockReadCloser()
		rc.On("Close").Return(errors.New("blah"))

		res := &http.Response{
			Body:       rc,
			StatusCode: http.StatusOK,
		}

		c, err := NewSimpleClient(
			ctx,
			mustParseURL(exampleURI),
			true,
		)
		assert.NotNil(t, c)
		assert.NoError(t, err)

		c.closeResponseBody(res)

		mock.AssertExpectationsForObjects(t, rc)
	})
}

func TestBuildURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		u, _ := url.Parse(exampleURI)

		c, err := NewClient(
			ctx,
			"",
			"",
			u,
			noop.NewLogger(),
			nil,
			[]string{"*"},
			false,
		)
		require.NoError(t, err)

		testCases := []struct {
			expectation string
			inputParts  []string
			inputQuery  valuer
		}{
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/api/v1/things",
				inputParts:  []string{"things"},
			},
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/api/v1/stuff?key=value",
				inputQuery:  map[string][]string{"key": {"value"}},
				inputParts:  []string{"stuff"},
			},
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/api/v1/things/and/stuff?key=value1&key=value2&yek=eulav",
				inputQuery: map[string][]string{
					"key": {"value1", "value2"},
					"yek": {"eulav"},
				},
				inputParts: []string{"things", "and", "stuff"},
			},
		}

		for _, tc := range testCases {
			actual := c.BuildURL(tc.inputQuery.ToValues(), tc.inputParts...)
			assert.Equal(t, tc.expectation, actual)
		}
	})

	T.Run("with invalid URL parts", func(t *testing.T) {
		t.Parallel()
		c := buildTestClientWithInvalidURL(t)
		assert.Empty(t, c.BuildURL(nil, asciiControlChar))
	})
}

func TestBuildVersionlessURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		u, _ := url.Parse(exampleURI)

		c, err := NewClient(
			ctx,
			"",
			"",
			u,
			noop.NewLogger(),
			nil,
			[]string{"*"},
			false,
		)
		require.NoError(t, err)

		testCases := []struct {
			expectation string
			inputParts  []string
			inputQuery  valuer
		}{
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/things",
				inputParts:  []string{"things"},
			},
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/stuff?key=value",
				inputQuery:  map[string][]string{"key": {"value"}},
				inputParts:  []string{"stuff"},
			},
			{
				expectation: "https://todo.verygoodsoftwarenotvirus.ru/things/and/stuff?key=value1&key=value2&yek=eulav",
				inputQuery: map[string][]string{
					"key": {"value1", "value2"},
					"yek": {"eulav"},
				},
				inputParts: []string{"things", "and", "stuff"},
			},
		}

		for _, tc := range testCases {
			actual := c.buildVersionlessURL(tc.inputQuery.ToValues(), tc.inputParts...)
			assert.Equal(t, tc.expectation, actual)
		}
	})

	T.Run("with invalid URL parts", func(t *testing.T) {
		t.Parallel()
		c := buildTestClientWithInvalidURL(t)
		assert.Empty(t, c.buildVersionlessURL(nil, asciiControlChar))
	})
}

func TestV1Client_BuildWebsocketURL(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		u, _ := url.Parse(exampleURI)

		c, err := NewClient(
			ctx,
			"",
			"",
			u,
			noop.NewLogger(),
			nil,
			[]string{"*"},
			false,
		)
		require.NoError(t, err)

		expected := "ws://todo.verygoodsoftwarenotvirus.ru/api/v1/things/and/stuff"
		actual := c.BuildWebsocketURL("things", "and", "stuff")

		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildHealthCheckRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodGet
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildHealthCheckRequest(ctx)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_IsUp(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, http.MethodGet)
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		c := buildTestClient(t, ts)
		actual := c.IsUp(ctx)
		assert.True(t, actual)
	})

	T.Run("returns error with invalid URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c := buildTestClientWithInvalidURL(t)

		actual := c.IsUp(ctx)
		assert.False(t, actual)
	})

	T.Run("with bad status code", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, http.MethodGet)
					res.WriteHeader(http.StatusInternalServerError)
				},
			),
		)

		c := buildTestClient(t, ts)
		actual := c.IsUp(ctx)
		assert.False(t, actual)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, http.MethodGet)
					time.Sleep(10 * time.Hour)
				},
			),
		)

		c := buildTestClient(t, ts)
		c.plainClient.Timeout = 500 * time.Millisecond
		actual := c.IsUp(ctx)
		assert.False(t, actual)
	})
}

func TestV1Client_buildDataRequest(T *testing.T) {
	T.Parallel()

	exampleData := &testingType{Name: "whatever"}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		expectedMethod := http.MethodPost
		req, err := c.buildDataRequest(ctx, expectedMethod, ts.URL, exampleData)

		require.NotNil(t, req)
		assert.NoError(t, err)

		assert.Equal(t, expectedMethod, req.Method)
		assert.Equal(t, ts.URL, req.URL.String())
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		x := &testBreakableStruct{Thing: "stuff"}
		req, err := c.buildDataRequest(ctx, http.MethodPost, ts.URL, x)

		require.Nil(t, req)
		assert.Error(t, err)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := buildTestClientWithInvalidURL(t)
		req, err := c.buildDataRequest(ctx, http.MethodPost, c.URL.String(), exampleData)

		require.Nil(t, req)
		assert.Error(t, err)
	})
}

func TestV1Client_executeRequest(T *testing.T) {
	T.Parallel()

	exampleResponse := &argleBargle{Name: "whatever"}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(exampleResponse))
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.executeRequest(ctx, req, &argleBargle{})
		assert.NoError(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					time.Sleep(10 * time.Hour)
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c.authedClient.Timeout = 500 * time.Millisecond
		err = c.executeRequest(ctx, req, &argleBargle{})
		assert.Error(t, err)
	})

	T.Run("with 401", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					res.WriteHeader(http.StatusUnauthorized)
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Equal(t, ErrUnauthorized, c.executeRequest(ctx, req, &argleBargle{}))
	})

	T.Run("with 404", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					res.WriteHeader(http.StatusNotFound)
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Equal(t, ErrNotFound, c.executeRequest(ctx, req, &argleBargle{}))
	})

	T.Run("with unreadable response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(exampleResponse))
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Error(t, c.executeRequest(ctx, req, argleBargle{}))
	})
}

func TestV1Client_executeRawRequest(T *testing.T) {
	T.Parallel()

	T.Run("with error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					time.Sleep(10 * time.Hour)
				},
			),
		)

		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res, err := c.executeRawRequest(ctx, &http.Client{Timeout: time.Second}, req)
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestV1Client_checkExistence(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodHead
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					res.WriteHeader(http.StatusOK)
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual, err := c.checkExistence(ctx, req)
		assert.True(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodHead
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					time.Sleep(10 * time.Hour)
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c.authedClient.Timeout = 500 * time.Millisecond
		actual, err := c.checkExistence(ctx, req)
		assert.False(t, actual)
		assert.Error(t, err)
	})
}

func TestV1Client_retrieve(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost
		exampleResponse := &argleBargle{Name: "whatever"}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(exampleResponse))
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.retrieve(ctx, req, &argleBargle{})
		assert.NoError(t, err)
	})

	T.Run("with nil passed in", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.retrieve(ctx, req, nil)
		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					time.Sleep(10 * time.Hour)
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c.authedClient.Timeout = 500 * time.Millisecond
		err = c.retrieve(ctx, req, &argleBargle{})
		assert.Error(t, err)
	})

	T.Run("with 404", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					res.WriteHeader(http.StatusNotFound)
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Equal(t, ErrNotFound, c.retrieve(ctx, req, &argleBargle{}))
	})
}

func TestV1Client_executeUnauthenticatedDataRequest(T *testing.T) {
	T.Parallel()

	const expectedMethod = http.MethodPost

	exampleResponse := &argleBargle{Name: "whatever"}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(exampleResponse))
				},
			),
		)
		c := buildTestClient(t, ts)

		in, out := &argleBargle{}, &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.executeUnauthenticatedDataRequest(ctx, req, out)
		assert.NoError(t, err)
	})

	T.Run("with 401", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					res.WriteHeader(http.StatusUnauthorized)
				},
			),
		)
		c := buildTestClient(t, ts)

		in, out := &argleBargle{}, &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.executeUnauthenticatedDataRequest(ctx, req, out)
		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorized, err)
	})

	T.Run("with 404", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					res.WriteHeader(http.StatusNotFound)
				},
			),
		)
		c := buildTestClient(t, ts)

		in, out := &argleBargle{}, &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.executeUnauthenticatedDataRequest(ctx, req, out)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					time.Sleep(10 * time.Hour)
				},
			),
		)
		c := buildTestClient(t, ts)

		in, out := &argleBargle{}, &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		c.plainClient.Timeout = 500 * time.Millisecond
		assert.Error(t, c.executeUnauthenticatedDataRequest(ctx, req, out))
	})

	T.Run("with nil as output", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		in := &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.executeUnauthenticatedDataRequest(ctx, req, testingType{})
		assert.Error(t, err)
	})

	T.Run("with unreadable response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(exampleResponse))
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Error(t, c.executeUnauthenticatedDataRequest(ctx, req, argleBargle{}))
	})
}
