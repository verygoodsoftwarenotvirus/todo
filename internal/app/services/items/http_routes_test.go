package items

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestItemsService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItemList := fakes.BuildFakeItemList()

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, mock.IsType(&types.QueryFilter{})).Return(exampleItemList, nil)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.ItemList{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, mock.IsType(&types.QueryFilter{})).Return((*types.ItemList)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.ItemList{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error fetching items from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, mock.IsType(&types.QueryFilter{})).Return((*types.ItemList)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_SearchHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, exampleAccount.ID).Return(exampleItemIDs, nil)
		s.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleLimit, exampleItemIDs).Return(exampleItemList, nil)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType([]*types.Item{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("http://todo.verygoodsoftwarenotvirus.ru?q=%s&limit=%d", exampleQuery, exampleLimit),
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.SearchHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, ed)
	})

	T.Run("with error conducting search", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, exampleAccount.ID).Return([]uint64{}, errors.New("blah"))
		s.search = indexManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("http://todo.verygoodsoftwarenotvirus.ru?q=%s&limit=%d", exampleQuery, exampleLimit),
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.SearchHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, ed)
	})

	T.Run("with now rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, exampleAccount.ID).Return(exampleItemIDs, nil)
		s.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleLimit, exampleItemIDs).Return([]*types.Item{}, sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType([]*types.Item{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("http://todo.verygoodsoftwarenotvirus.ru?q=%s&limit=%d", exampleQuery, exampleLimit),
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.SearchHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, ed)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Search", mock.MatchedBy(testutil.ContextMatcher), exampleQuery, exampleAccount.ID).Return(exampleItemIDs, nil)
		s.search = indexManager

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.MatchedBy(testutil.ContextMatcher), exampleAccount.ID, exampleLimit, exampleItemIDs).Return([]*types.Item{}, errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("http://todo.verygoodsoftwarenotvirus.ru?q=%s&limit=%d", exampleQuery, exampleLimit),
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.SearchHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, indexManager, itemDataManager, ed)
	})
}

func TestItemsService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.ItemCreationInput{})).Return(exampleItem, nil)
		s.itemDataManager = itemDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher))
		s.itemCounter = mc

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Index", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleItem).Return(nil)
		s.search = indexManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Item{}), http.StatusCreated)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, mc, indexManager, ed)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with error creating item", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.ItemCreationInput{})).Return((*types.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_ExistenceHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(true, nil)
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with no such item in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(false, sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error fetching item from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(false, errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_ReadHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(exampleItem, nil)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Item{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with no such item in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return((*types.Item)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error fetching item from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return((*types.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Item{}), mock.IsType([]types.FieldChangeSummary{})).Return(nil)
		s.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Index", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleItem).Return(nil)
		s.search = indexManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("RespondWithData", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Item{}))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code, "expected %d in status response, got %d", http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)

		mock.AssertExpectationsForObjects(t, ed)
	})

	T.Run("with no rows fetching item", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return((*types.Item)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return((*types.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error updating item", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.MatchedBy(testutil.ContextMatcher), mock.IsType(&types.Item{}), mock.IsType([]types.FieldChangeSummary{})).Return(errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})
}

func TestItemsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	exampleUser, exampleAccount, examplePerms := fakes.BuildUserTestPrerequisites()
	requestContextFetcher := func(_ *http.Request) (*types.RequestContext, error) {
		reqCtx, err := types.RequestContextFromUser(exampleUser, exampleAccount.ID, examplePerms)
		require.NoError(T, err)
		return reqCtx, nil
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(nil)
		s.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Delete", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID).Return(nil)
		s.search = indexManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
		s.itemCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, mc)
	})

	T.Run("with no item in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher), mock.MatchedBy(testutil.ResponseWriterMatcher()))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error removing from search index", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.requestContextFetcher = requestContextFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToAccount = exampleAccount.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID, exampleAccount.ID).Return(nil)
		s.itemDataManager = itemDataManager

		indexManager := &mocksearch.IndexManager{}
		indexManager.On("Delete", mock.MatchedBy(testutil.ContextMatcher), exampleItem.ID).Return(errors.New("blah"))
		s.search = indexManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher)).Return()
		s.itemCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, indexManager, mc)
	})
}
