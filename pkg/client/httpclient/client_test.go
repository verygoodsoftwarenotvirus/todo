package httpclient

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClient_AuthenticatedClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, ts := buildSimpleTestClient(t)

		assert.Equal(t, ts.Client(), c.AuthenticatedClient(), "AuthenticatedClient should return the assigned authedClient")
	})
}

func TestClient_PlainClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, ts := buildSimpleTestClient(t)

		assert.Equal(t, ts.Client(), c.PlainClient(), "PlainClient should return the assigned unauthenticatedClient")
	})
}

func TestNewClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, err := NewClient(
			mustParseURL(exampleURI),
			UsingLogger(logging.NewNoopLogger()),
		)

		require.NotNil(t, c)
		require.NoError(t, err)
	})

	T.Run("with nil URL", func(t *testing.T) {
		t.Parallel()

		c, err := NewClient(
			nil,
			UsingLogger(logging.NewNoopLogger()),
		)

		require.Nil(t, c)
		require.Error(t, err)
	})
}

func TestClient_RequestBuilder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := buildSimpleTestClient(t)

		assert.NotNil(t, c.RequestBuilder())
	})
}

func TestClient_BuildURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()

		c, _ := NewClient(mustParseURL(exampleURI))
		ctx := context.Background()

		testCases := []struct {
			inputQuery  valuer
			expectation string
			inputParts  []string
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
			actual := c.BuildURL(ctx, tc.inputQuery.ToValues(), tc.inputParts...)
			assert.Equal(t, tc.expectation, actual)
		}
	})

	T.Run("with invalid url parts", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := buildTestClientWithInvalidURL(t)

		assert.Empty(t, c.BuildURL(ctx, nil, asciiControlChar))
	})
}

func TestClient_CloseRequestBody(T *testing.T) {
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

		c, _ := NewClient(mustParseURL(exampleURI))
		assert.NotNil(t, c)

		c.closeResponseBody(ctx, res)

		mock.AssertExpectationsForObjects(t, rc)
	})
}

func TestBuildVersionlessURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()

		c, _ := NewClient(mustParseURL(exampleURI))

		testCases := []struct {
			inputQuery  valuer
			expectation string
			inputParts  []string
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
			ctx := context.Background()
			actual := c.buildVersionlessURL(ctx, tc.inputQuery.ToValues(), tc.inputParts...)
			assert.Equal(t, tc.expectation, actual)
		}
	})

	T.Run("with invalid url parts", func(t *testing.T) {
		t.Parallel()
		c := buildTestClientWithInvalidURL(t)
		ctx := context.Background()
		actual := c.buildVersionlessURL(ctx, nil, asciiControlChar)
		assert.Empty(t, actual)
	})
}

func TestClient_BuildWebsocketURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, _ := NewClient(mustParseURL(exampleURI))
		ctx := context.Background()

		expected := "ws://todo.verygoodsoftwarenotvirus.ru/api/v1/things/and/stuff"
		actual := c.BuildWebsocketURL(ctx, nil, "things", "and", "stuff")

		assert.Equal(t, expected, actual)
	})
}

func TestClient_IsUp(T *testing.T) {
	T.Parallel()

	spec := newRequestSpec(true, http.MethodGet, "", "/_meta_/ready")

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		actual := c.IsUp(ctx)
		assert.True(t, actual)
	})

	T.Run("returns error with invalid url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c := buildTestClientWithInvalidURL(t)

		actual := c.IsUp(ctx)
		assert.False(t, actual)
	})

	T.Run("with bad status code", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusInternalServerError)
		actual := c.IsUp(ctx)
		assert.False(t, actual)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _ := buildTestClientThatWaitsTooLong(t)

		actual := c.IsUp(ctx)
		assert.False(t, actual)
	})
}

func TestClient_fetchAndUnmarshal(T *testing.T) {
	T.Parallel()

	exampleResponse := &argleBargle{Name: "whatever"}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(true, http.MethodPost, "", "/")
		c, ts := buildTestClientWithJSONResponse(t, spec, exampleResponse)

		req, err := http.NewRequestWithContext(ctx, spec.method, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.fetchAndUnmarshal(ctx, req, &argleBargle{})
		assert.NoError(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(true, http.MethodPost, "", "/")
		c, ts := buildTestClientThatWaitsTooLong(t)

		req, err := http.NewRequestWithContext(ctx, spec.method, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.fetchAndUnmarshal(ctx, req, &argleBargle{})
		assert.Error(t, err)
	})

	T.Run("with 401", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(true, http.MethodPost, "", "/")
		c, ts := buildTestClientWithStatusCodeResponse(t, spec, http.StatusUnauthorized)

		req, err := http.NewRequestWithContext(ctx, spec.method, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.True(t, errors.Is(c.fetchAndUnmarshal(ctx, req, &argleBargle{}), ErrUnauthorized))
	})

	T.Run("with 404", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(true, http.MethodPost, "", "/")
		c, ts := buildTestClientWithStatusCodeResponse(t, spec, http.StatusNotFound)

		req, err := http.NewRequestWithContext(ctx, spec.method, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.True(t, errors.Is(c.fetchAndUnmarshal(ctx, req, &argleBargle{}), ErrNotFound))
	})

	T.Run("with unreadable response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(true, http.MethodPost, "", "/")
		c, ts := buildTestClientWithJSONResponse(t, spec, exampleResponse)

		req, err := http.NewRequestWithContext(ctx, spec.method, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Error(t, c.fetchAndUnmarshal(ctx, req, argleBargle{}))
	})
}

func TestClient_executeRawRequest(T *testing.T) {
	T.Parallel()

	T.Run("with error", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost

		c, ts := buildTestClientThatWaitsTooLong(t)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		res, err := c.fetchResponseToRequest(ctx, &http.Client{Timeout: time.Second}, req)
		assert.Nil(t, res)
		assert.Error(t, err)
	})
}

func TestClient_checkExistence(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodHead

		spec := newRequestSpec(true, expectedMethod, "", "/")
		c, ts := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		actual, err := c.responseIsOK(ctx, req)
		assert.True(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodHead
		c, ts := buildTestClientThatWaitsTooLong(t)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c.authedClient.Timeout = 500 * time.Millisecond
		actual, err := c.responseIsOK(ctx, req)
		assert.False(t, actual)
		assert.Error(t, err)
	})
}

func TestClient_retrieve(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost
		exampleResponse := &argleBargle{Name: "whatever"}

		spec := newRequestSpec(false, expectedMethod, "", "/")
		c, ts := buildTestClientWithJSONResponse(t, spec, exampleResponse)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.fetchAndUnmarshal(ctx, req, &argleBargle{})
		assert.NoError(t, err)
	})

	T.Run("with nil passed in", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, ts := buildSimpleTestClient(t)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		err = c.fetchAndUnmarshal(ctx, req, nil)
		assert.Error(t, err)
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodPost

		c, ts := buildTestClientThatWaitsTooLong(t)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		c.authedClient.Timeout = 500 * time.Millisecond
		err = c.fetchAndUnmarshal(ctx, req, &argleBargle{})
		assert.Error(t, err)
	})

	T.Run("with 404", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		expectedMethod := http.MethodPost
		spec := newRequestSpec(true, expectedMethod, "", "/")
		c, ts := buildTestClientWithStatusCodeResponse(t, spec, http.StatusNotFound)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.True(t, errors.Is(c.fetchAndUnmarshal(ctx, req, &argleBargle{}), ErrNotFound))
	})
}

func TestClient_fetchAndUnmarshalWithoutAuthentication(T *testing.T) {
	T.Parallel()

	const expectedMethod = http.MethodPost

	exampleResponse := &argleBargle{Name: "whatever"}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(false, expectedMethod, "", "/")
		c, ts := buildTestClientWithJSONResponse(t, spec, exampleResponse)

		in, out := &argleBargle{}, &argleBargle{}

		body := createBodyFromStruct(t, in)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.fetchAndUnmarshalWithoutAuthentication(ctx, req, out)
		assert.NoError(t, err)
	})

	T.Run("with 401", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(false, expectedMethod, "", "/")
		c, ts := buildTestClientWithStatusCodeResponse(t, spec, http.StatusUnauthorized)

		in, out := &argleBargle{}, &argleBargle{}

		body := createBodyFromStruct(t, in)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.fetchAndUnmarshalWithoutAuthentication(ctx, req, out)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrUnauthorized))
	})

	T.Run("with 404", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(false, expectedMethod, "", "/")
		c, ts := buildTestClientWithStatusCodeResponse(t, spec, http.StatusNotFound)

		in, out := &argleBargle{}, &argleBargle{}

		body := createBodyFromStruct(t, in)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.fetchAndUnmarshalWithoutAuthentication(ctx, req, out)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})

	T.Run("with timeout", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, ts := buildTestClientThatWaitsTooLong(t)

		in, out := &argleBargle{}, &argleBargle{}

		body := createBodyFromStruct(t, in)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		c.unauthenticatedClient.Timeout = 500 * time.Millisecond
		assert.Error(t, c.fetchAndUnmarshalWithoutAuthentication(ctx, req, out))
	})

	T.Run("with nil as output", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, ts := buildSimpleTestClient(t)

		in := &argleBargle{}

		body := createBodyFromStruct(t, in)
		require.NotNil(t, body)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, body)
		require.NoError(t, err)
		require.NotNil(t, req)

		err = c.fetchAndUnmarshalWithoutAuthentication(ctx, req, testingType{})
		assert.Error(t, err)
	})

	T.Run("with unreadable response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		spec := newRequestSpec(false, expectedMethod, "", "/")
		c, ts := buildTestClientWithJSONResponse(t, spec, exampleResponse)

		req, err := http.NewRequestWithContext(ctx, expectedMethod, ts.URL, nil)
		require.NotNil(t, req)
		require.NoError(t, err)

		assert.Error(t, c.fetchAndUnmarshalWithoutAuthentication(ctx, req, argleBargle{}))
	})
}
