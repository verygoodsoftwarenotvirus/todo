package requests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	panicking "gitlab.com/verygoodsoftwarenotvirus/todo/internal/panicking"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type (
	testingType struct {
		Name string
	}

	testBreakableStruct struct {
		Thing json.Number
	}
)

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

	T.Run("with nil URL", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		encoder := encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON)
		c, err := NewBuilder(nil, logger, encoder)

		require.Nil(t, c)
		require.Error(t, err)
	})

	T.Run("with nil encoder", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		c, err := NewBuilder(mustParseURL(exampleURI), logger, nil)

		require.Nil(t, c)
		require.Error(t, err)
	})
}

func TestBuilder_URL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		assert.NotNil(t, h.builder.URL())
	})
}

func TestBuilder_SetURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		assert.NoError(t, h.builder.SetURL(&url.URL{}))
	})

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		assert.Error(t, h.builder.SetURL(nil))
	})
}

func TestBuilder_BuildURL(T *testing.T) {
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
		c := buildTestRequestBuilderWithInvalidURL()

		assert.Empty(t, c.BuildURL(ctx, nil, asciiControlChar))
	})
}

func TestBuilder_Must(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		h.builder.Must(&http.Request{}, nil)
	})

	T.Run("with panic", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()
		exampleErr := errors.New("blah")

		mockPanicker := panicking.NewMockPanicker()
		mockPanicker.On("Panic", exampleErr).Return()
		h.builder.panicker = mockPanicker

		h.builder.Must(&http.Request{}, exampleErr)

		mock.AssertExpectationsForObjects(t, mockPanicker)
	})
}

func Test_buildRawURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual, err := buildRawURL(parsedExampleURL, url.Values{}, true, "things", "and", "stuff")
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func TestBuilder_buildAPIV1URL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildTestHelper()

		assert.NotNil(t, h.builder.buildAPIV1URL(h.ctx, url.Values{}, "things", "and", "stuff"))
	})
}

func TestBuilder_buildUnversionedURL(T *testing.T) {
	T.Parallel()

	T.Run("various urls", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNonOperationalLogger()
		encoder := encoding.ProvideClientEncoder(logger, encoding.ContentTypeJSON)
		b, err := NewBuilder(mustParseURL(exampleURI), logger, encoder)

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
			actual := b.buildUnversionedURL(ctx, tc.inputQuery.ToValues(), tc.inputParts...)
			assert.Equal(t, tc.expectation, actual)
		}
	})

	T.Run("with invalid url parts", func(t *testing.T) {
		t.Parallel()
		c := buildTestRequestBuilderWithInvalidURL()
		ctx := context.Background()
		actual := c.buildUnversionedURL(ctx, nil, asciiControlChar)
		assert.Empty(t, actual)
	})
}

func TestBuilder_BuildWebsocketURL(T *testing.T) {
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

func TestBuilder_BuildHealthCheckRequest(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodGet

		c := buildTestRequestBuilder()
		actual, err := c.BuildHealthCheckRequest(ctx)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestBuilder_buildDataRequest(T *testing.T) {
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
		c := buildTestRequestBuilderWithInvalidURL()
		req, err := c.buildDataRequest(ctx, http.MethodPost, c.url.String(), exampleData)

		require.Nil(t, req)
		assert.Error(t, err)
	})
}

func Test_mustParseURL(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		mustParseURL(exampleURI)
	})
}
