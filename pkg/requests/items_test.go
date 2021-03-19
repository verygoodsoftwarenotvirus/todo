package requests

import (
	"context"
	"net/http"
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
	builder         *Builder
	exampleItem     *types.Item
	exampleInput    *types.ItemCreationInput
	exampleItemList *types.ItemList
}

var _ suite.SetupTestSuite = (*itemsBaseSuite)(nil)

func (s *itemsBaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.builder = buildTestRequestBuilder()
	s.exampleItem = fakes.BuildFakeItem()
	s.exampleInput = fakes.BuildFakeItemCreationInputFromItem(s.exampleItem)
	s.exampleItemList = fakes.BuildFakeItemList()
}

type itemsTestSuite struct {
	suite.Suite

	itemsBaseSuite
}

func (s *itemsTestSuite) TestBuilder_BuildItemExistsRequest() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("standard", func() {
		t := s.T()

		actual, err := s.builder.BuildItemExistsRequest(s.ctx, s.exampleItem.ID)
		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, s.exampleItem.ID)

		assert.NoError(t, err)
		assertRequestQuality(t, actual, spec)
	})
}

func (s *itemsTestSuite) TestBuilder_BuildGetItemRequest() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleItem.ID)

		actual, err := s.builder.BuildGetItemRequest(s.ctx, s.exampleItem.ID)
		assert.NoError(t, err)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *itemsTestSuite) TestBuilder_BuildGetItemsRequest() {
	const expectedPath = "/api/v1/items"

	s.Run("standard", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)
		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

		actual, err := s.builder.BuildGetItemsRequest(s.ctx, filter)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *itemsTestSuite) TestBuilder_BuildSearchItemsRequest() {
	const expectedPath = "/api/v1/items/search"

	s.Run("standard", func() {
		t := s.T()

		limit := types.DefaultQueryFilter().Limit
		exampleQuery := "whatever"
		spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)

		actual, err := s.builder.BuildSearchItemsRequest(s.ctx, exampleQuery, limit)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *itemsTestSuite) TestBuilder_BuildCreateItemRequest() {
	const expectedPath = "/api/v1/items"

	s.Run("standard", func() {
		t := s.T()

		actual, err := s.builder.BuildCreateItemRequest(s.ctx, s.exampleInput)
		assert.NoError(t, err)

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

		assertRequestQuality(t, actual, spec)
	})
}

func (s *itemsTestSuite) TestBuilder_BuildUpdateItemRequest() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleItem.ID)

		actual, err := s.builder.BuildUpdateItemRequest(s.ctx, s.exampleItem)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *itemsTestSuite) TestBuilder_BuildArchiveItemRequest() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleItem.ID)

		actual, err := s.builder.BuildArchiveItemRequest(s.ctx, s.exampleItem.ID)
		assert.NoError(t, err, "no error should be returned")

		assertRequestQuality(t, actual, spec)
	})
}

func (s *itemsTestSuite) TestBuilder_BuildGetAuditLogForItemRequest() {
	const expectedPath = "/api/v1/items/%d/audit"

	s.Run("standard", func() {
		t := s.T()

		actual, err := s.builder.BuildGetAuditLogForItemRequest(s.ctx, s.exampleItem.ID)
		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")

		spec := newRequestSpec(true, http.MethodGet, "", expectedPath, s.exampleItem.ID)
		assertRequestQuality(t, actual, spec)
	})
}
