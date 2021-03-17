package requests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItems(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(itemsTestSuite))
}

type itemsBaseSuite struct {
	suite.Suite

	ctx             context.Context
	exampleItem     *types.Item
	exampleInput    *types.ItemCreationInput
	exampleItemList *types.ItemList
}

var _ suite.SetupTestSuite = (*itemsBaseSuite)(nil)

func (s *itemsBaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleItem = fakes.BuildFakeItem()
	s.exampleInput = fakes.BuildFakeItemCreationInputFromItem(s.exampleItem)
	s.exampleItemList = fakes.BuildFakeItemList()
}

type itemsTestSuite struct {
	suite.Suite

	itemsBaseSuite
}

func TestV1Client_BuildItemExistsRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)
		exampleItem := fakes.BuildFakeItem()
		actual, err := c.BuildItemExistsRequest(ctx, exampleItem.ID)
		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, exampleItem.ID)

		assert.NoError(t, err)
		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)

		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, exampleItem.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetItemRequest(ctx, exampleItem.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetItemsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		filter := (*types.QueryFilter)(nil)
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		c := buildTestClient(t, ts)
		actual, err := c.BuildGetItemsRequest(ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildSearchItemsRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items/search"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		limit := types.DefaultQueryFilter().Limit
		exampleQuery := "whatever"
		spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)
		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildSearchItemsRequest(ctx, exampleQuery, limit)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildCreateItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleAccount := fakes.BuildFakeAccount()
		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		ts := httptest.NewTLSServer(nil)

		c := buildTestClient(t, ts)
		actual, err := c.BuildCreateItemRequest(ctx, exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildUpdateItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleItem := fakes.BuildFakeItem()
		ts := httptest.NewTLSServer(nil)
		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, exampleItem.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildUpdateItemRequest(ctx, exampleItem)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildArchiveItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPathFormat = "/api/v1/items/%d"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ts := httptest.NewTLSServer(nil)
		exampleItem := fakes.BuildFakeItem()
		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, exampleItem.ID)

		c := buildTestClient(t, ts)
		actual, err := c.BuildArchiveItemRequest(ctx, exampleItem.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func TestV1Client_BuildGetAuditLogForItemRequest(T *testing.T) {
	T.Parallel()

	const expectedPath = "/api/v1/items/%d/audit"

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		exampleItem := fakes.BuildFakeItem()
		ts := httptest.NewTLSServer(nil)
		c := buildTestClient(t, ts)

		actual, err := c.BuildGetAuditLogForItemRequest(ctx, exampleItem.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, exampleItem.ID)
		assertRequestQuality(t, actual, spec)
	})
}
