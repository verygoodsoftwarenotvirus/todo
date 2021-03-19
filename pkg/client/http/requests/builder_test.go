package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
)

const (
	exampleURI       = "https://todo.verygoodsoftwarenotvirus.ru"
	asciiControlChar = string(byte(127))
)

var (
	parsedExampleURL *url.URL
)

func init() {
	var err error

	parsedExampleURL, err = url.Parse(exampleURI)
	if err != nil {
		panic(err)
	}
}

// begin test helpers

type (
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

func buildTestRequestBuilder() *Builder {
	l := logging.NewNonOperationalLogger()

	return &Builder{
		url:     parsedExampleURL,
		logger:  l,
		tracer:  tracing.NewTracer("test"),
		encoder: encoding.ProvideClientEncoder(l, encoding.ContentTypeJSON),
	}
}

func buildTestRequestBuilderWithInvalidURL(t *testing.T) *Builder {
	t.Helper()

	l := logging.NewNonOperationalLogger()
	u := mustParseURL("https://verygoodsoftwarenotvirus.ru")
	u.Scheme = fmt.Sprintf(`%s://`, asciiControlChar)

	return &Builder{
		url:     u,
		logger:  l,
		tracer:  tracing.NewTracer("test"),
		encoder: encoding.ProvideClientEncoder(l, encoding.ContentTypeJSON),
	}
}

// end test helpers

func TestNewBuilder(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		encoder := encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON)
		c, err := NewBuilder(mustParseURL(exampleURI), logger, encoder)

		require.NotNil(t, c)
		require.NoError(t, err)
	})
}

func TestBuildURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		encoder := encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON)

		c, _ := NewBuilder(mustParseURL(exampleURI), logger, encoder)
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
		c := buildTestRequestBuilderWithInvalidURL(t)

		assert.Empty(t, c.BuildURL(ctx, nil, asciiControlChar))
	})
}

func TestBuildVersionlessURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		encoder := encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON)
		c, err := NewBuilder(mustParseURL(exampleURI), logger, encoder)

		require.NoError(t, err)

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
		c := buildTestRequestBuilderWithInvalidURL(t)
		ctx := context.Background()
		actual := c.buildVersionlessURL(ctx, nil, asciiControlChar)
		assert.Empty(t, actual)
	})
}

func TestV1Client_BuildWebsocketURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		logger := logging.NewNonOperationalLogger()
		encoder := encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON)
		c, err := NewBuilder(mustParseURL(exampleURI), logger, encoder)

		require.NoError(t, err)

		expected := "ws://todo.verygoodsoftwarenotvirus.ru/api/v1/things/and/stuff"
		actual := c.BuildWebsocketURL(ctx, "things", "and", "stuff")

		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildHealthCheckRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodGet

		c := buildTestRequestBuilder()
		actual, err := c.BuildHealthCheckRequest(ctx)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

type (
	testingType struct {
		Name string
	}

	testBreakableStruct struct {
		Thing json.Number
	}
)

func TestV1Client_buildDataRequest(T *testing.T) {
	T.Parallel()

	exampleData := &testingType{Name: "whatever"}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		c := buildTestRequestBuilder()
		expectedMethod := http.MethodPost
		req, err := c.buildDataRequest(ctx, expectedMethod, exampleURI, exampleData)

		require.NotNil(t, req)
		assert.NoError(t, err)

		assert.Equal(t, expectedMethod, req.Method)
		assert.Equal(t, exampleURI, req.URL.String())
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()

		c := buildTestRequestBuilder()
		x := &testBreakableStruct{Thing: "stuff"}
		req, err := c.buildDataRequest(ctx, http.MethodPost, exampleURI, x)

		require.Nil(t, req)
		assert.Error(t, err)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		c := buildTestRequestBuilderWithInvalidURL(t)
		req, err := c.buildDataRequest(ctx, http.MethodPost, c.url.String(), exampleData)

		require.Nil(t, req)
		assert.Error(t, err)
	})
}
