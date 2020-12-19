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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestItemsService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItemList := fakes.BuildFakeItemList()

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.Anything, exampleUser.ID, mock.AnythingOfType("*types.QueryFilter")).Return(exampleItemList, nil)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.ItemList"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.Anything, exampleUser.ID, mock.AnythingOfType("*types.QueryFilter")).Return((*types.ItemList)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.ItemList"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with error fetching items from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItems", mock.Anything, exampleUser.ID, mock.AnythingOfType("*types.QueryFilter")).Return((*types.ItemList)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return(exampleItemIDs, nil)
		s.search = si

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.Anything, exampleUser.ID, exampleLimit, exampleItemIDs).Return(exampleItemList, nil)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("[]types.Item"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, si, itemDataManager, ed)
	})

	T.Run("with error conducting search", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return([]uint64{}, errors.New("blah"))
		s.search = si

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

		mock.AssertExpectationsForObjects(t, si, ed)
	})

	T.Run("with now rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return(exampleItemIDs, nil)
		s.search = si

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.Anything, exampleUser.ID, exampleLimit, exampleItemIDs).Return([]types.Item{}, sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("[]types.Item"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, si, itemDataManager, ed)
	})

	T.Run("with error fetching from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakes.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return(exampleItemIDs, nil)
		s.search = si

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.Anything, exampleUser.ID, exampleLimit, exampleItemIDs).Return([]types.Item{}, errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

		mock.AssertExpectationsForObjects(t, si, itemDataManager, ed)
	})
}

func TestItemsService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.Anything, mock.AnythingOfType("*types.ItemCreationInput")).Return(exampleItem, nil)
		s.itemDataManager = itemDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		si := &mocksearch.IndexManager{}
		si.On("Index", mock.Anything, exampleItem.ID, exampleItem).Return(nil)
		s.search = si

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogItemCreationEvent", mock.Anything, exampleItem)
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponseWithStatus", mock.Anything, mock.AnythingOfType("*types.Item"), http.StatusCreated)
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

		mock.AssertExpectationsForObjects(t, itemDataManager, mc, si, ed)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNoInputResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemCreationInputFromItem(exampleItem)

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.Anything, mock.AnythingOfType("*types.ItemCreationInput")).Return((*types.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.Anything, exampleItem.ID, exampleUser.ID).Return(true, nil)
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with no such item in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.Anything, exampleItem.ID, exampleUser.ID).Return(false, sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.Anything, exampleItem.ID, exampleUser.ID).Return(false, errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
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

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(exampleItem, nil)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.Item"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("with no such item in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*types.Item)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*types.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.Anything, mock.AnythingOfType("*types.Item")).Return(nil)
		s.itemDataManager = itemDataManager

		si := &mocksearch.IndexManager{}
		si.On("Index", mock.Anything, exampleItem.ID, exampleItem).Return(nil)
		s.search = si

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogItemUpdateEvent", mock.Anything, exampleUser.ID, exampleItem.ID, mock.AnythingOfType("[]types.FieldChangeSummary"))
		s.auditLog = auditLog

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*types.Item"))
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

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNoInputResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*types.Item)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*types.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.Anything, mock.AnythingOfType("*types.Item")).Return(errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return exampleUser.ToSessionInfo(), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(nil)
		s.itemDataManager = itemDataManager

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogItemArchiveEvent", mock.Anything, exampleUser.ID, exampleItem.ID)
		s.auditLog = auditLog

		si := &mocksearch.IndexManager{}
		si.On("Delete", mock.Anything, exampleItem.ID).Return(nil)
		s.search = si

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
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

		mock.AssertExpectationsForObjects(t, itemDataManager, mc)
	})

	T.Run("with no item in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeNotFoundResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(errors.New("blah"))
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.Anything)
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
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleItem := fakes.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mocktypes.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(nil)
		s.itemDataManager = itemDataManager

		auditLog := &mocktypes.AuditLogDataManager{}
		auditLog.On("LogItemArchiveEvent", mock.Anything, exampleUser.ID, exampleItem.ID)
		s.auditLog = auditLog

		si := &mocksearch.IndexManager{}
		si.On("Delete", mock.Anything, exampleItem.ID).Return(errors.New("blah"))
		s.search = si

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
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

		mock.AssertExpectationsForObjects(t, itemDataManager, mc)
	})
}
