package httpclient

import (
	"context"
	"encoding/json"
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

type itemsTestSuite struct {
	suite.Suite

	ctx             context.Context
	exampleItem     *types.Item
	exampleInput    *types.ItemCreationInput
	exampleItemList *types.ItemList
}

var _ suite.SetupTestSuite = (*itemsTestSuite)(nil)

func (s *itemsTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.exampleItem = fakes.BuildFakeItem()
	s.exampleInput = fakes.BuildFakeItemCreationInputFromItem(s.exampleItem)
	s.exampleItemList = fakes.BuildFakeItemList()
}

func (s *itemsTestSuite) TestV1Client_ItemExists() {
	const expectedPathFormat = "/api/v1/items/%d"

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodHead, "", expectedPathFormat, s.exampleItem.ID)
		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusOK)
			},
		))

		c := buildTestClient(t, ts)
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

	s.Run("happy path", func() {
		t := s.T()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleItem))
			},
		))

		c := buildTestClient(t, ts)
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

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetItem(s.ctx, s.exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_GetItems() {
	const expectedPath = "/api/v1/items"

	spec := newRequestSpec(true, http.MethodGet, "includeArchived=false&limit=20&page=1&sortBy=asc", expectedPath)

	s.Run("happy path", func() {
		t := s.T()

		filter := (*types.QueryFilter)(nil)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleItemList))
			},
		))

		c := buildTestClient(t, ts)
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

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetItems(s.ctx, filter)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_SearchItems() {
	const expectedPath = "/api/v1/items/search"

	exampleQuery := "whatever"
	spec := newRequestSpec(true, http.MethodGet, "limit=20&q=whatever", expectedPath)

	s.Run("happy path", func() {
		t := s.T()

		limit := types.DefaultQueryFilter().Limit

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleItemList.Items))
			},
		))

		c := buildTestClient(t, ts)
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

	s.Run("happy path", func() {
		t := s.T()

		limit := types.DefaultQueryFilter().Limit
		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.SearchItems(s.ctx, exampleQuery, limit)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}

func (s *itemsTestSuite) TestV1Client_CreateItem() {
	const expectedPath = "/api/v1/items"

	spec := newRequestSpec(false, http.MethodPost, "", expectedPath)

	s.Run("happy path", func() {
		t := s.T()
		ctx := context.Background()

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				var x *types.ItemCreationInput
				require.NoError(t, json.NewDecoder(req.Body).Decode(&x))

				s.exampleInput.BelongsToAccount = 0
				assert.Equal(t, s.exampleInput, x)

				require.NoError(t, json.NewEncoder(res).Encode(s.exampleItem))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.CreateItem(ctx, s.exampleInput)

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

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(false, http.MethodPut, "", expectedPathFormat, s.exampleItem.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				assert.NoError(t, json.NewEncoder(res).Encode(s.exampleItem))
			},
		))

		err := buildTestClient(t, ts).UpdateItem(s.ctx, s.exampleItem)
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

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, http.MethodDelete, "", expectedPathFormat, s.exampleItem.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				res.WriteHeader(http.StatusOK)
			},
		))

		err := buildTestClient(t, ts).ArchiveItem(s.ctx, s.exampleItem.ID)
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

	s.Run("happy path", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleItem.ID)
		exampleAuditLogEntryList := fakes.BuildFakeAuditLogEntryList().Entries

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode(exampleAuditLogEntryList))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForItem(s.ctx, s.exampleItem.ID)

		require.NotNil(t, actual)
		assert.NoError(t, err, "no error should be returned")
		assert.Equal(t, exampleAuditLogEntryList, actual)
	})

	s.Run("with invalid response", func() {
		t := s.T()

		spec := newRequestSpec(true, expectedMethod, "", expectedPath, s.exampleItem.ID)

		ts := httptest.NewTLSServer(http.HandlerFunc(
			func(res http.ResponseWriter, req *http.Request) {
				assertRequestQuality(t, req, spec)

				require.NoError(t, json.NewEncoder(res).Encode("BLAH"))
			},
		))

		c := buildTestClient(t, ts)
		actual, err := c.GetAuditLogForItem(s.ctx, s.exampleItem.ID)

		assert.Nil(t, actual)
		assert.Error(t, err, "error should be returned")
	})
}
