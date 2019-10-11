package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

const (
	exampleURI = "https://todo.verygoodsoftwarenotvirus.ru"
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

	l := noop.ProvideNoopLogger()
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

// end helper funcs

func TestV1Client_AuthenticatedClient(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual := c.AuthenticatedClient()

		assert.Equal(t, ts.Client(), actual, "AuthenticatedClient should return the assigned authedClient")
	})
}

func TestV1Client_PlainClient(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual := c.PlainClient()

		assert.Equal(t, ts.Client(), actual, "PlainClient should return the assigned plainClient")
	})
}

func TestV1Client_TokenSource(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ts := httptest.NewTLSServer(nil)
		c, err := NewClient(
			context.Background(),
			"",
			"",
			mustParseURL(exampleURI),
			noop.ProvideNoopLogger(),
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
		ts := httptest.NewTLSServer(nil)
		c, err := NewClient(
			context.Background(),
			"",
			"",
			mustParseURL(exampleURI),
			noop.ProvideNoopLogger(),
			ts.Client(),
			[]string{"*"},
			false,
		)

		require.NotNil(t, c)
		require.NoError(t, err)
	})

	T.Run("with client but invalid timeout", func(t *testing.T) {
		c, err := NewClient(
			context.Background(),
			"",
			"",
			mustParseURL(exampleURI),
			noop.ProvideNoopLogger(),
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
		c, err := NewSimpleClient(
			context.Background(),
			mustParseURL(exampleURI),
			true,
		)
		assert.NotNil(t, c)
		assert.NoError(t, err)
	})
}

func TestV1Client_executeRequest(T *testing.T) {
	T.Parallel()

	T.Run("with error", func(t *testing.T) {
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

		req, err := http.NewRequest(expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res, err := c.executeRawRequest(ctx, &http.Client{Timeout: time.Second}, req)
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestBuildURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()

		u, _ := url.Parse(exampleURI)
		c, err := NewClient(
			context.Background(),
			"",
			"",
			u,
			noop.ProvideNoopLogger(),
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
}

func TestV1Client_BuildWebsocketURL(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		u, _ := url.Parse(exampleURI)
		c, err := NewClient(
			context.Background(),
			"",
			"",
			u,
			noop.ProvideNoopLogger(),
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
		expectedMethod := http.MethodGet
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildHealthCheckRequest()

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_IsUp(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, http.MethodGet)
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		c := buildTestClient(t, ts)
		actual := c.IsUp()
		assert.True(t, actual)
	})

	T.Run("with bad status code", func(t *testing.T) {
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, http.MethodGet)
					res.WriteHeader(http.StatusInternalServerError)
				},
			),
		)

		c := buildTestClient(t, ts)
		actual := c.IsUp()
		assert.False(t, actual)
	})

	T.Run("with timeout", func(t *testing.T) {
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
		actual := c.IsUp()
		assert.False(t, actual)
	})
}

func TestV1Client_buildDataRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		expectedMethod := http.MethodPost
		req, err := c.buildDataRequest(expectedMethod, ts.URL, &testingType{Name: "name"})

		require.NotNil(t, req)
		assert.NoError(t, err)
		assert.Equal(t, expectedMethod, req.Method)
	})
}

func TestV1Client_makeRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(&argleBargle{Name: "name"}))
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequest(expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.executeRequest(ctx, req, &argleBargle{})
		assert.NoError(t, err)
	})

	T.Run("with 404", func(t *testing.T) {
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

		req, err := http.NewRequest(expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Equal(t, ErrNotFound, c.executeRequest(ctx, req, &argleBargle{}))
	})
}

func TestV1Client_makeUnauthedDataRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(&argleBargle{Name: "name"}))
				},
			),
		)
		c := buildTestClient(t, ts)

		in, out := &argleBargle{}, &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequest(expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.executeUnathenticatedDataRequest(ctx, req, out)
		assert.NoError(t, err)
	})

	T.Run("with 404", func(t *testing.T) {
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

		in, out := &argleBargle{}, &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequest(expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.executeUnathenticatedDataRequest(ctx, req, out)
		assert.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	T.Run("with timeout", func(t *testing.T) {
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

		in, out := &argleBargle{}, &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequest(expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		c.plainClient.Timeout = 500 * time.Millisecond
		assert.Error(t, c.executeUnathenticatedDataRequest(ctx, req, out))
	})

	T.Run("with nil as output", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		in := &argleBargle{}

		body, err := createBodyFromStruct(in)
		require.NoError(t, err)
		require.NotNil(t, body)

		req, err := http.NewRequest(expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.executeUnathenticatedDataRequest(ctx, req, testingType{})
		assert.Error(t, err)
	})
}

func TestV1Client_retrieve(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.Method, expectedMethod)
					require.NoError(t, json.NewEncoder(res).Encode(&argleBargle{Name: "name"}))
				},
			),
		)
		c := buildTestClient(t, ts)

		req, err := http.NewRequest(expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.retrieve(ctx, req, &argleBargle{})
		assert.NoError(t, err)
	})

	T.Run("with nil passed in", func(t *testing.T) {
		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		req, err := http.NewRequest(http.MethodPost, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.retrieve(ctx, req, nil)
		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
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

		req, err := http.NewRequest(expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c.authedClient.Timeout = 500 * time.Millisecond
		err = c.retrieve(ctx, req, &argleBargle{})
		assert.Error(t, err)
	})

	T.Run("with 404", func(t *testing.T) {
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

		req, err := http.NewRequest(expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Equal(t, ErrNotFound, c.retrieve(ctx, req, &argleBargle{}))
	})
}
