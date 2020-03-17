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

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildItemExistsRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodHead
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		expectedItemID := fake.Uint64()
		actual, err := c.BuildItemExistsRequest(ctx, expectedItemID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedItemID)))
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_ItemExists(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expected := true

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.String(), strconv.Itoa(int(expectedItemID))))
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", expectedItemID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodHead)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.ItemExists(ctx, expectedItemID)

		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})

	T.Run("nonexistent", func(t *testing.T) {
		ctx := context.Background()
		expectedItemID := fake.Uint64()
		expected := false

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.String(), strconv.Itoa(int(expectedItemID))))
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", expectedItemID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodHead)

					res.WriteHeader(http.StatusNotFound)
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.ItemExists(ctx, expectedItemID)

		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, expected, actual)
	})
}

func TestV1Client_BuildGetItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expectedMethod := http.MethodGet
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		expectedItemID := fake.Uint64()
		actual, err := c.BuildGetItemRequest(ctx, expectedItemID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", expectedItemID)))
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Item{
			ID: fake.Uint64(),
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.String(), strconv.Itoa(int(expected.ID))))
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
		ctx := context.Background()
		expectedMethod := http.MethodGet
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetItemsRequest(ctx, nil)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.ItemList{
			Items: []models.Item{
				{
					ID: fake.Uint64(),
				},
			},
		}

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
		ctx := context.Background()
		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(nil)

		exampleInput := &models.ItemCreationInput{
			Name:    fake.Word(),
			Details: fake.Word(),
		}
		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateItemRequest(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_CreateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Item{
			ID:      fake.Uint64(),
			Name:    fake.Word(),
			Details: fake.Word(),
		}
		exampleInput := &models.ItemCreationInput{
			Name:    expected.Name,
			Details: expected.Details,
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, "/api/v1/items", "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPost)

					var x *models.ItemCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))
					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(expected))
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
		ctx := context.Background()
		expectedMethod := http.MethodPut
		exampleInput := &models.Item{
			ID:      fake.Uint64(),
			Name:    fake.Word(),
			Details: fake.Word(),
		}

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateItemRequest(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.Item{
			ID: fake.Uint64(),
		}

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", expected.ID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodPut)
					assert.NoError(t, json.NewEncoder(res).Encode(&models.Item{}))
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
		ctx := context.Background()
		expectedMethod := http.MethodDelete
		ts := httptest.NewTLSServer(nil)

		exampleItemID := fake.Uint64()
		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveItemRequest(ctx, exampleItemID)

		require.NotNil(t, actual)
		require.NotNil(t, actual.URL)
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", exampleItemID)))
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleItemID := fake.Uint64()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", exampleItemID), "expected and actual path don't match")
					assert.Equal(t, req.Method, http.MethodDelete)
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveItem(ctx, exampleItemID)
		assert.NoError(t, err, "no error should be returned")
	})
}
