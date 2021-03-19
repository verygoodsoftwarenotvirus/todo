package httpclient

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

func (s *itemsTestSuite) TestV1Client_ItemExists() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, s.exampleItem.ID)

		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		actual, err := c.ItemExists(s.ctx, s.exampleItem.ID)

		assert.NoError(t, err, "no error should be returned")
		assert.True(t, actual)
	})

	s.Run("with erroneous response", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.ItemExists(s.ctx, s.exampleItem.ID)

		assert.Error(t, err, "error should be returned")
		assert.False(t, actual)
	})
}

func (s *itemsTestSuite) TestV1Client_GetItem() {
	const expectedPathFormat = "/api/v1/items/%d"

	spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleItem.ID)

	s.Run("standard", func() {
		t := s.T()

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleItem)
		actual, err := c.GetItem(s.ctx, s.exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleItem, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItem(s.ctx, s.exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetItem(s.ctx, s.exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_GetItems() {
	const expectedPath = "/api/v1/items"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleItemList)
		actual, err := c.GetItems(s.ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleItemList, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItems(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("with invalid response", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetItems(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_SearchItems() {
	const expectedPath = "/api/v1/items/search"

	exampleQuery := "whatever"
	spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		limit := types.DefaultQueryFilter().Limit

		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleItemList.Items)
		actual, err := c.SearchItems(s.ctx, exampleQuery, limit)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleItemList.Items, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		limit := types.DefaultQueryFilter().Limit
		c := buildTestClientWithInvalidURL(t)

		actual, err := c.SearchItems(s.ctx, exampleQuery, limit)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})

	s.Run("standard", func() {
		t := s.T()

		limit := types.DefaultQueryFilter().Limit
		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.SearchItems(s.ctx, exampleQuery, limit)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_CreateItem() {
	const expectedPath = "/api/v1/items"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("standard", func() {
		t := s.T()

		s.exampleInput.BelongsToAccount = 0

		c := buildTestClientWithRequestBodyValidation(t, spec, &types.ItemCreationInput{}, s.exampleInput, s.exampleItem)
		actual, err := c.CreateItem(s.ctx, s.exampleInput)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, s.exampleItem, actual)
	})

	s.Run("with invalid client url", func() {
		t := s.T()
		ctx := context.Background()

		exampleItem := fakes.BuildFakeItem()
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.CreateItem(ctx, exampleInput)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_UpdateItem() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleItem.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleItem)

		err := c.UpdateItem(s.ctx, s.exampleItem)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).UpdateItem(s.ctx, s.exampleItem)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_ArchiveItem() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleItem.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		err := c.ArchiveItem(s.ctx, s.exampleItem.ID)
		assert.NoError(t, err, "no error should be returned")
	})

	s.Run("with invalid client url", func() {
		t := s.T()

		err := buildTestClientWithInvalidURL(t).ArchiveItem(s.ctx, s.exampleItem.ID)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_GetAuditLogForItem() {
	const (
		expectedPath   = "/api/v1/items/%d/audit"
		expectedMethod = http.MethodGet
	)

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleItem.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		c, _ := buildTestClientWithJSONResponse(t, spec, exampleAuditLogEntryList)
		actual, err := c.GetAuditLogForItem(s.ctx, s.exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleItem.ID)

		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetAuditLogForItem(s.ctx, s.exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
