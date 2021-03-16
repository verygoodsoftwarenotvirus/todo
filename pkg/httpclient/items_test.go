package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestV1Client_ItemExists(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, exampleItem.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

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

		exampleItem := fakes.BuildFakeItem()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.ItemExists(ctx, exampleItem.ID)

		assert.Error(t, err, "error should be returned")
		assert.False(t, actual)
	})
}

func TestV1Client_GetItem(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleItem.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

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

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItem(ctx, exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleItem.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

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

func TestV1Client_GetItems(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		exampleItemList := fakes.BuildFakeItemList()

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

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

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItems(ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

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

func TestV1Client_SearchItems(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items/search"

	exampleQuery := "whatever"
	spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		limit := types.DefaultQueryFilter().Limit
		exampleItemList := fakes.BuildFakeItemList().Items
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

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

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		limit := types.DefaultQueryFilter().Limit
		c := buildTestClientWithInvalidURL(t)

		actual, err := c.SearchItems(ctx, exampleQuery, limit)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		limit := types.DefaultQueryFilter().Limit
		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

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

func TestV1Client_CreateItem(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					var x *types.ItemCreationInput
					require.NoError(t, json.NewDecoder(req.Body).Decode(&x))

					exampleInput.BelongsToAccount = 0
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

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateItem(ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_UpdateItem(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleItem.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					assert.NoError(t, json.NewEncoder(res).Encode(exampleItem))
				},
			),
		)

		err := buildTestClient(t, ts).UpdateItem(ctx, exampleItem)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()

		err := buildTestClientWithInvalidURL(t).UpdateItem(ctx, exampleItem)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_ArchiveItem(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleItem.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					res.WriteHeader(http.StatusOK)
				},
			),
		)

		err := buildTestClient(t, ts).ArchiveItem(ctx, exampleItem.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()

		err := buildTestClientWithInvalidURL(t).ArchiveItem(ctx, exampleItem.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func TestV1Client_GetAuditLogForItem(T *testing.T) {
	T.Parallel()

	const (
		expectedPath   = "/api/v1/items/%d/audit"
		expectedMethod = http.MethodGet
	)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleItem.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForItem(ctx, exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	T.Run("with invalid client url", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetAuditLogForItem(ctx, exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	T.Run("with invalid response", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, expectedMethod, "", expectedPath, exampleItem.ID)

		ts := httptest.NewTLSServer(
			http.HandlerFunc(
				func(res http.ResponseWriter, req *http.Request) {
					assertRequestQuality(t, req, spec)

					require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
				},
			),
		)

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForItem(ctx, exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
