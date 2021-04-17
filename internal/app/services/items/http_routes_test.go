package items

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/metrics/mock"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/search/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/util/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestParseBool(t *testing.T) {
	t.Parallel()

	expectations := map[string]bool{
		"1":     true,
		"fart":  false,
		"true":  true,
		"troo":  false,
		"t":     true,
		"false": false,
	}

	for input, expected := range expectations {
		assert.Equal(t, expected, parseBool(input))
	}
}

func TestItemsService_CreateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleItem,
		).Return(nil)
		helper.service.search = indexManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Item{}),
			http.StatusCreated,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, unitCounter, indexManager, encoderDecoder)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req = testutil.BuildTestRequest(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error creating item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
		).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error indexing item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"CreateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.ItemCreationInput{}),
		).Return(helper.exampleItem, nil)
		helper.service.itemDataManager = itemDataManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Increment", testutil.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleItem,
		).Return(errors.New("blah"))
		helper.service.search = indexManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeResponseWithStatus",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Item{}),
			http.StatusCreated,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.CreateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusCreated, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, unitCounter, indexManager, encoderDecoder)
	})
}

func TestItemsService_ReadHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		h := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			h.exampleItem.ID, h.exampleAccount.ID,
		).Return(h.exampleItem, nil)
		h.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Item{}))
		h.service.encoderDecoder = encoderDecoder

		h.service.ReadHandler(h.res, h.req)

		assert.Equal(t, http.StatusOK, h.res.Code, "expected %d in status response, got %d", http.StatusOK, h.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such item in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return((*types.Item)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ReadHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()
		h := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			h.exampleItem.ID, h.exampleAccount.ID,
		).Return((*types.Item)(nil), errors.New("blah"))
		h.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher)
		h.service.encoderDecoder = encoderDecoder

		h.service.ReadHandler(h.res, h.req)

		assert.Equal(t, http.StatusInternalServerError, h.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_ExistenceHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ItemExists",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(true, nil)
		helper.service.itemDataManager = itemDataManager

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no result in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ItemExists",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(false, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error checking database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ItemExists",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(false, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ExistenceHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_ListHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleItemList := fakes.BuildFakeItemList()

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.ItemList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("standard for admin", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{types.AdminQueryKey: []string{"true"}}.Encode()

		helper.exampleUser.ServiceAdminPermission = testutil.BuildMaxServiceAdminPerms()
		helper.service.sessionContextDataFetcher = func(_ *http.Request) (*types.SessionContextData, error) {
			sessionCtxData, err := types.SessionContextDataFromUser(
				helper.exampleUser,
				helper.exampleAccount.ID,
				map[uint64]*types.UserAccountMembershipInfo{
					helper.exampleAccount.ID: {
						AccountName: helper.exampleAccount.Name,
						Permissions: testutil.BuildMaxUserPerms(),
					},
				},
			)
			require.NoError(t, err)

			return sessionCtxData, nil
		}

		exampleItemList := fakes.BuildFakeItemList()

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsForAdmin",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleItemList, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.ItemList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.ItemList{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving items from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItems",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.ItemList)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ListHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_SearchHandler(T *testing.T) {
	T.Parallel()

	exampleQuery := "whatever"
	exampleLimit := uint8(123)
	exampleItemList := fakes.BuildFakeItemList()
	exampleItemIDs := []uint64{}
	for _, x := range exampleItemList.Items {
		exampleItemIDs = append(exampleItemIDs, x.ID)
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{
			types.SearchQueryKey: []string{exampleQuery},
			types.LimitQueryKey:  []string{strconv.Itoa(int(exampleLimit))},
		}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return(exampleItemList.Items, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType([]*types.Item{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, encoderDecoder)
	})

	T.Run("standard for admin", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{
			types.SearchQueryKey: []string{exampleQuery},
			types.AdminQueryKey:  []string{"true"},
			types.LimitQueryKey:  []string{strconv.Itoa(int(exampleLimit))},
		}.Encode()

		helper.exampleUser.ServiceAdminPermission = testutil.BuildMaxServiceAdminPerms()
		helper.service.sessionContextDataFetcher = func(_ *http.Request) (*types.SessionContextData, error) {
			sessionCtxData, err := types.SessionContextDataFromUser(
				helper.exampleUser,
				helper.exampleAccount.ID,
				map[uint64]*types.UserAccountMembershipInfo{
					helper.exampleAccount.ID: {
						AccountName: helper.exampleAccount.Name,
						Permissions: testutil.BuildMaxUserPerms(),
					},
				},
			)
			require.NoError(t, err)

			return sessionCtxData, nil
		}

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"SearchForAdmin",
			testutil.ContextMatcher,
			exampleQuery,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDsForAdmin",
			testutil.ContextMatcher,
			exampleLimit,
			exampleItemIDs,
		).Return(exampleItemList.Items, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType([]*types.Item{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with error conducting search", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{types.SearchQueryKey: []string{exampleQuery}}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return([]uint64{}, errors.New("blah"))
		helper.service.search = indexManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, indexManager, encoderDecoder)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{
			types.SearchQueryKey: []string{exampleQuery},
			types.LimitQueryKey:  []string{strconv.Itoa(int(exampleLimit))},
		}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return([]*types.Item{}, sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType([]*types.Item{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code, "expected %d in status response, got %d", http.StatusOK, helper.res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req.URL.RawQuery = url.Values{
			types.SearchQueryKey: []string{exampleQuery},
			types.LimitQueryKey:  []string{strconv.Itoa(int(exampleLimit))},
		}.Encode()

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Search",
			testutil.ContextMatcher,
			exampleQuery,
			helper.exampleAccount.ID,
		).Return(exampleItemIDs, nil)
		helper.service.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItemsWithIDs",
			testutil.ContextMatcher,
			helper.exampleAccount.ID,
			exampleLimit,
			exampleItemIDs,
		).Return([]*types.Item{}, errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.SearchHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, encoderDecoder)
	})
}

func TestItemsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		h := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			h.exampleItem.ID, h.exampleAccount.ID,
		).Return(h.exampleItem, nil)
		itemDataManager.On(
			"UpdateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.Item{}),
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		h.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			h.exampleItem.ID, h.exampleItem,
		).Return(nil)
		h.service.search = indexManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Item{}))
		h.service.encoderDecoder = encoderDecoder

		h.service.UpdateHandler(h.res, h.req)

		assert.Equal(t, http.StatusOK, h.res.Code, "expected %d in status response, got %d", http.StatusOK, h.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("without input attached to context", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		helper.req = testutil.BuildTestRequest(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeInvalidInputResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusBadRequest, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return((*types.Item)(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving item from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return((*types.Item)(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error updating item", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(helper.exampleItem, nil)
		itemDataManager.On(
			"UpdateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.Item{}),
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.UpdateHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error updating search index", func(t *testing.T) {
		t.Parallel()
		h := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetItem",
			testutil.ContextMatcher,
			h.exampleItem.ID, h.exampleAccount.ID,
		).Return(h.exampleItem, nil)
		itemDataManager.On(
			"UpdateItem",
			testutil.ContextMatcher,
			mock.IsType(&types.Item{}),
			mock.IsType([]*types.FieldChangeSummary{}),
		).Return(nil)
		h.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Index",
			testutil.ContextMatcher,
			h.exampleItem.ID, h.exampleItem,
		).Return(errors.New("blah"))
		h.service.search = indexManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType(&types.Item{}))
		h.service.encoderDecoder = encoderDecoder

		h.service.UpdateHandler(h.res, h.req)

		assert.Equal(t, http.StatusOK, h.res.Code, "expected %d in status response, got %d", http.StatusOK, h.res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, encoderDecoder)
	})
}

func TestItemsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		h := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			h.exampleItem.ID, h.exampleAccount.ID,
		).Return(nil)
		h.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Delete",
			testutil.ContextMatcher,
			h.exampleItem.ID,
		).Return(nil)
		h.service.search = indexManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		h.service.itemCounter = unitCounter

		h.service.ArchiveHandler(h.res, h.req)

		assert.Equal(t, http.StatusNoContent, h.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, unitCounter)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		)
		helper.service.encoderDecoder = encoderDecoder

		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with no such item in the database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error saving as archived", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error removing from search index", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"ArchiveItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID, helper.exampleAccount.ID,
		).Return(nil)
		helper.service.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On(
			"Delete",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return(errors.New("blah"))
		helper.service.search = indexManager

		unitCounter := &mockmetrics.UnitCounter{}
		unitCounter.On("Decrement", testutil.ContextMatcher).Return()
		helper.service.itemCounter = unitCounter

		helper.service.ArchiveHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNoContent, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, unitCounter)
	})
}

func TestAccountsService_AuditEntryHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		exampleAuditLogEntries := fakes.BuildFakeAuditLogEntryList().Entries

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetAuditLogEntriesForItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return(exampleAuditLogEntries, nil)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"RespondWithData",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			mock.IsType([]*types.AuditLogEntry{}),
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusOK, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error retrieving session context data", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)
		helper.service.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
			"unauthenticated",
			http.StatusUnauthorized,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusUnauthorized, helper.res.Code)
		mock.AssertExpectationsForObjects(t, encoderDecoder)
	})

	T.Run("with sql.ErrNoRows", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetAuditLogEntriesForItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return([]*types.AuditLogEntry(nil), sql.ErrNoRows)
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeNotFoundResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusNotFound, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On(
			"GetAuditLogEntriesForItem",
			testutil.ContextMatcher,
			helper.exampleItem.ID,
		).Return([]*types.AuditLogEntry(nil), errors.New("blah"))
		helper.service.itemDataManager = itemDataManager

		encoderDecoder := mockencoding.NewMockEncoderDecoder()
		encoderDecoder.On(
			"EncodeUnspecifiedInternalServerErrorResponse",
			testutil.ContextMatcher,
			testutil.ResponseWriterMatcher,
		).Return()
		helper.service.encoderDecoder = encoderDecoder

		helper.service.AuditEntryHandler(helper.res, helper.req)

		assert.Equal(t, http.StatusInternalServerError, helper.res.Code)
		mock.AssertExpectationsForObjects(t, itemDataManager, encoderDecoder)
	})
}
