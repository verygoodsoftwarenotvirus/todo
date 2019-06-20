package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildGetItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodGet
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		expectedID := uint64(1)

		actual, err := c.BuildGetItemRequest(ctx, expectedID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Item{
			ID:      1,
			Name:    "example",
			Details: "blah",
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t,
						strings.HasSuffix(
							req.URL.String(),
							strconv.Itoa(int(expected.ID)),
						),
					)
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", expected.ID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.GetItem(ctx, expected.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildGetItemsRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodGet
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetItemsRequest(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.ItemList{
			Items: []models.Item{
				{
					ID:      1,
					Name:    "example",
					Details: "blah",
				},
			},
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/api/v1/items", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(expected))
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.GetItems(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildCreateItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodPost
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleInput := &models.ItemCreationInput{
			Name:    "expected name",
			Details: "expected details",
		}
		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateItemRequest(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Item{
			ID:      1,
			Name:    "example",
			Details: "blah",
		}

		exampleInput := &models.ItemCreationInput{
			Name:    expected.Name,
			Details: expected.Details,
		}

		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/api/v1/items", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)

					var x *models.ItemCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))
					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(expected))
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		c := buildTestClient(t, ts)

		actual, err := c.CreateItem(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildUpdateItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodPut
		ctx := context.Background()

		exampleInput := &models.Item{
			Name:    "changed name",
			Details: "changed details",
		}

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateItemRequest(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.Item{
			ID:      1,
			Name:    "example",
			Details: "blah",
		}
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", expected.ID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPut)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).UpdateItem(ctx, expected)

		assert.NoError(t, err, "no error should be returned")
	})
}

func TestV1Client_BuildArchiveItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expectedMethod := http.MethodDelete
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		expectedID := uint64(1)
		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveItemRequest(ctx, expectedID)

		require.NotNil(t, actual)
		require.NotNil(t, actual.URL)
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedID)))
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t,
			actual.Method,
			expectedMethod,
			"request should be a %s request",
			expectedMethod,
		)
	})
}

func TestV1Client_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(1)
		ctx := context.Background()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", expected), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodDelete)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveItem(ctx, expected)

		assert.NoError(t, err, "no error should be returned")
	})
}
