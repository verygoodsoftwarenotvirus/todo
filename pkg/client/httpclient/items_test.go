package httpclient

import (
	"context"
	"net/http"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestItems(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(itemsTestSuite))
}

type itemsBaseSuite struct {
	suite.Suite

	ctx         context.Context
	exampleItem *types.Item
}

var _ suite.SetupTestSuite = (*itemsBaseSuite)(nil)

func (s *itemsBaseSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleItem = fakes.BuildFakeItem()
}

type itemsTestSuite struct {
	suite.Suite

	itemsBaseSuite
}

func (s *itemsTestSuite) TestClient_ItemExists() {
	const expectedPathFormat = "/api/v1/items/%s"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, s.exampleItem.ID)

		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)
		actual, err := c.ItemExists(s.ctx, s.exampleItem.ID)

		assert.NoError(t, err)
		assert.True(t, actual)
	})

	s.Run("with invalid item ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		actual, err := c.ItemExists(s.ctx, "")

		assert.Error(t, err)
		assert.False(t, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.ItemExists(s.ctx, s.exampleItem.ID)

		assert.Error(t, err)
		assert.False(t, actual)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)
		actual, err := c.ItemExists(s.ctx, s.exampleItem.ID)

		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func (s *itemsTestSuite) TestClient_GetItem() {
	const expectedPathFormat = "/api/v1/items/%s"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleItem.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleItem)
		actual, err := c.GetItem(s.ctx, s.exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, s.exampleItem, actual)
	})

	s.Run("with invalid item ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		actual, err := c.GetItem(s.ctx, "")

		require.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItem(s.ctx, s.exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "", expectedPathFormat, s.exampleItem.ID)
		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetItem(s.ctx, s.exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *itemsTestSuite) TestClient_GetItems() {
	const expectedPath = "/api/v1/items"

	s.Run("standard", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		exampleItemList := fakes.BuildFakeItemList()

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, exampleItemList)
		actual, err := c.GetItems(s.ctx, filter)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList, actual)
	})

	s.Run("with error building request", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		c := buildTestClientWithInvalidURL(t)
		actual, err := c.GetItems(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)
		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.GetItems(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *itemsTestSuite) TestClient_SearchItems() {
	const expectedPath = "/api/v1/items/search"

	exampleQuery := "whatever"

	s.Run("standard", func() {
		t := s.T()

		exampleItemList := fakes.BuildFakeItemList()

		spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, exampleItemList.Items)
		actual, err := c.SearchItems(s.ctx, exampleQuery, 0)

		require.NotNil(t, actual)
		assert.NoError(t, err)
		assert.Equal(t, exampleItemList.Items, actual)
	})

	s.Run("with empty query", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		actual, err := c.SearchItems(s.ctx, "", 0)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.SearchItems(s.ctx, exampleQuery, 0)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})

	s.Run("with bad response from server", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)
		c := buildTestClientWithInvalidResponse(t, spec)
		actual, err := c.SearchItems(s.ctx, exampleQuery, 0)

		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}

func (s *itemsTestSuite) TestClient_CreateItem() {
	const expectedPath = "/api/v1/items"

	s.Run("standard", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeItemCreationInput()
		exampleInput.BelongsToAccount = ""

		spec := newRequestSpec(false, http.MethodPost, "", expectedPath)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleItem)

		actual, err := c.CreateItem(s.ctx, exampleInput)
		require.NotNil(t, actual)
		assert.NoError(t, err)

		assert.Equal(t, s.exampleItem.ID, actual)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		actual, err := c.CreateItem(s.ctx, nil)
		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	s.Run("with invalid input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)
		exampleInput := &types.ItemCreationInput{}

		actual, err := c.CreateItem(s.ctx, exampleInput)
		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeItemCreationInputFromItem(s.exampleItem)

		c := buildTestClientWithInvalidURL(t)

		actual, err := c.CreateItem(s.ctx, exampleInput)
		assert.Empty(t, actual)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		exampleInput := fakes.BuildFakeItemCreationInputFromItem(s.exampleItem)
		c, _ := buildTestClientThatWaitsTooLong(t)

		actual, err := c.CreateItem(s.ctx, exampleInput)
		assert.Empty(t, actual)
		assert.Error(t, err)
	})
}

func (s *itemsTestSuite) TestClient_UpdateItem() {
	const expectedPathFormat = "/api/v1/items/%s"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleItem.ID)
		c, _ := buildTestClientWithJSONResponse(t, spec, s.exampleItem)

		err := c.UpdateItem(s.ctx, s.exampleItem)
		assert.NoError(t, err)
	})

	s.Run("with nil input", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		err := c.UpdateItem(s.ctx, nil)
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		err := c.UpdateItem(s.ctx, s.exampleItem)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		err := c.UpdateItem(s.ctx, s.exampleItem)
		assert.Error(t, err)
	})
}

func (s *itemsTestSuite) TestClient_ArchiveItem() {
	const expectedPathFormat = "/api/v1/items/%s"

	s.Run("standard", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleItem.ID)
		c, _ := buildTestClientWithStatusCodeResponse(t, spec, http.StatusOK)

		err := c.ArchiveItem(s.ctx, s.exampleItem.ID)
		assert.NoError(t, err)
	})

	s.Run("with invalid item ID", func() {
		t := s.T()

		c, _ := buildSimpleTestClient(t)

		err := c.ArchiveItem(s.ctx, "")
		assert.Error(t, err)
	})

	s.Run("with error building request", func() {
		t := s.T()

		c := buildTestClientWithInvalidURL(t)

		err := c.ArchiveItem(s.ctx, s.exampleItem.ID)
		assert.Error(t, err)
	})

	s.Run("with error executing request", func() {
		t := s.T()

		c, _ := buildTestClientThatWaitsTooLong(t)

		err := c.ArchiveItem(s.ctx, s.exampleItem.ID)
		assert.Error(t, err)
	})
}
