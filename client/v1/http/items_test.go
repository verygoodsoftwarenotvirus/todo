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

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_BuildItemExistsRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodHead
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		exampleItem := fakemodels.BuildFakeItem()
		actual, err := c.BuildItemExistsRequest(ctx, exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", exampleItem.ID)))
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_ItemExists(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.String(), strconv.Itoa(int(exampleItem.ID))))
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", exampleItem.ID), "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodHead)
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.ItemExists(ctx, exampleItem.ID)

		assert.NoError(t, err, "no error should be returned")
		assert.True(t, actual)
	})

	T.Run("with erroneous response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.ItemExists(ctx, exampleItem.ID)

		assert.Error(t, err, "error should be returned")
		assert.False(t, actual)
	})
}

func TestV1Client_BuildGetItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodGet
		ts := httptest.NewTLSServer(nil)

		exampleItem := fakemodels.BuildFakeItem()

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetItemRequest(ctx, exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", exampleItem.ID)))
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_GetItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.String(), strconv.Itoa(int(exampleItem.ID))))
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", exampleItem.ID), "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(exampleItem))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetItem(ctx, exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleItem, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItem(ctx, exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.True(t, strings.HasSuffix(req.URL.String(), strconv.Itoa(int(exampleItem.ID))))
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", exampleItem.ID), "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetItem(ctx, exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildGetItemsRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*models.QueryFilter)(nil)
		expectedMethod := http.MethodGet
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetItemsRequest(ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_GetItems(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*models.QueryFilter)(nil)

		expectedPath := "/api/v1/items"

		exampleItemList := fakemodels.BuildFakeItemList()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(exampleItemList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetItems(ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleItemList, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*models.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItems(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*models.QueryFilter)(nil)

		expectedPath := "/api/v1/items"

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetItems(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildSearchItemsRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		limit := models.DefaultQueryFilter().Limit
		exampleQuery := "whatever"

		expectedMethod := http.MethodGet
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildSearchItemsRequest(ctx, exampleQuery, limit)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_SearchItems(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items/search"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		limit := models.DefaultQueryFilter().Limit
		exampleQuery := "whatever"

		exampleItemList := fakemodels.BuildFakeItemList().Items

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.URL.Query().Get(models.SearchQueryKey), exampleQuery, "expected and actual search query param do not match")
					assert.Equal(t, req.URL.Query().Get(models.LimitQueryKey), strconv.FormatUint(uint64(limit), 10), "expected and actual limit query param do not match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode(exampleItemList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.SearchItems(ctx, exampleQuery, limit)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleItemList, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		limit := models.DefaultQueryFilter().Limit
		exampleQuery := "whatever"

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.SearchItems(ctx, exampleQuery, limit)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		limit := models.DefaultQueryFilter().Limit
		exampleQuery := "whatever"

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.URL.Query().Get(models.SearchQueryKey), exampleQuery, "expected and actual search query param do not match")
					assert.Equal(t, req.URL.Query().Get(models.LimitQueryKey), strconv.FormatUint(uint64(limit), 10), "expected and actual limit query param do not match")
					assert.Equal(t, req.Method, http.MethodGet)
					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.SearchItems(ctx, exampleQuery, limit)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildCreateItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakemodels.BuildFakeUser()
		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		expectedMethod := http.MethodPost
		ts := httptest.NewTLSServer(nil)

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
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		expectedPath := "/api/v1/items"

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, expectedPath, "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodPost)

					var x *models.ItemCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))

					exampleInput.BelongsToUser = 0
					assert.Equal(t, exampleInput, x)

					require.NoError(t, json.NewEncoder(res).Encode(exampleItem))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.CreateItem(ctx, exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleItem, actual)
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateItem(ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildUpdateItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()
		expectedMethod := http.MethodPut

		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateItemRequest(ctx, exampleItem)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_UpdateItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", exampleItem.ID), "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodPut)
					assert.NoError(t, json.NewEncoder(res).Encode(exampleItem))
				},
			),
		)

		err := buildTestClient(t, ts).UpdateItem(ctx, exampleItem)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		err := buildTestClientWithInvalidURL(t).UpdateItem(ctx, exampleItem)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_BuildArchiveItemRequest(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expectedMethod := http.MethodDelete
		ts := httptest.NewTLSServer(nil)

		exampleItem := fakemodels.BuildFakeItem()

		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveItemRequest(ctx, exampleItem.ID)

		require.NotNil(t, actual)
		require.NotNil(t, actual.URL)
		assert.True(t, strings.HasSuffix(actual.URL.String(), fmt.Sprintf("%d", exampleItem.ID)))
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, actual.Method, expectedMethod, "request should be a %s request", expectedMethod)
	})
}

func TestV1Client_ArchiveItem(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assert.Equal(t, req.URL.Path, fmt.Sprintf("/api/v1/items/%d", exampleItem.ID), "expected and actual paths do not match")
					assert.Equal(t, req.Method, http.MethodDelete)
					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveItem(ctx, exampleItem.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client URL", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakemodels.BuildFakeItem()

		err := buildTestClientWithInvalidURL(t).ArchiveItem(ctx, exampleItem.ID)
		assert.Error(t, err, "error should be returned")
	})
}
