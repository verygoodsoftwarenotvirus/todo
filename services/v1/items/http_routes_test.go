package items

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
	mocksearch "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search/mock"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	mocknewsman "gitlab.com/verygoodsoftwarenotvirus/newsman/mock"
)

func TestItemsService_ListHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	userIDFetcher := func(_ *http.Request) uint64 {
		return exampleUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItemList := fakemodels.BuildFakeItemList()

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItems", mock.Anything, exampleUser.ID, mock.AnythingOfType("*models.QueryFilter")).Return(exampleItemList, nil)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.ItemList")).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItems", mock.Anything, exampleUser.ID, mock.AnythingOfType("*models.QueryFilter")).Return((*models.ItemList)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.ItemList")).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItems", mock.Anything, exampleUser.ID, mock.AnythingOfType("*models.QueryFilter")).Return((*models.ItemList)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})
}

func TestItemsService_SearchHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	userIDFetcher := func(_ *http.Request) uint64 {
		return exampleUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakemodels.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return(exampleItemIDs, nil)
		s.search = si

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.Anything, exampleUser.ID, exampleLimit, exampleItemIDs).Return(exampleItemList, nil)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("[]models.Item")).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return([]uint64{}, errors.New("blah"))
		s.search = si

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("http://todo.verygoodsoftwarenotvirus.ru?q=%s&limit=%d", exampleQuery, exampleLimit),
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.SearchHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, si)
	})

	T.Run("with now rows returned", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakemodels.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return(exampleItemIDs, nil)
		s.search = si

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.Anything, exampleUser.ID, exampleLimit, exampleItemIDs).Return([]models.Item{}, sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("[]models.Item")).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleQuery := "whatever"
		exampleLimit := uint8(123)
		exampleItemList := fakemodels.BuildFakeItemList().Items
		var exampleItemIDs []uint64
		for _, x := range exampleItemList {
			exampleItemIDs = append(exampleItemIDs, x.ID)
		}

		si := &mocksearch.IndexManager{}
		si.On("Search", mock.Anything, exampleQuery, exampleUser.ID).Return(exampleItemIDs, nil)
		s.search = si

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItemsWithIDs", mock.Anything, exampleUser.ID, exampleLimit, exampleItemIDs).Return([]models.Item{}, errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("http://todo.verygoodsoftwarenotvirus.ru?q=%s&limit=%d", exampleQuery, exampleLimit),
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.SearchHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, si, itemDataManager)
	})
}

func TestItemsService_CreateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	userIDFetcher := func(_ *http.Request) uint64 {
		return exampleUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.Anything, mock.AnythingOfType("*models.ItemCreationInput")).Return(exampleItem, nil)
		s.itemDataManager = itemDataManager

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.AnythingOfType("newsman.Event")).Return()
		s.reporter = r

		si := &mocksearch.IndexManager{}
		si.On("Index", mock.Anything, exampleItem.ID, exampleItem).Return(nil)
		s.search = si

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.Item")).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, mc, r, si, ed)
	})

	T.Run("without input attached", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error creating item", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("CreateItem", mock.Anything, mock.AnythingOfType("*models.ItemCreationInput")).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), createMiddlewareCtxKey, exampleInput))

		s.CreateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})
}

func TestItemsService_ExistenceHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	userIDFetcher := func(_ *http.Request) uint64 {
		return exampleUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.Anything, exampleItem.ID, exampleUser.ID).Return(true, nil)
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.Anything, exampleItem.ID, exampleUser.ID).Return(false, sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error fetching item from database", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("ItemExists", mock.Anything, exampleItem.ID, exampleUser.ID).Return(false, errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})
}

func TestItemsService_ReadHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	userIDFetcher := func(_ *http.Request) uint64 {
		return exampleUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(exampleItem, nil)
		s.itemDataManager = itemDataManager

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.Item")).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
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
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*models.Item)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error fetching item from database", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})
}

func TestItemsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	userIDFetcher := func(_ *http.Request) uint64 {
		return exampleUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.Anything, mock.AnythingOfType("*models.Item")).Return(nil)
		s.itemDataManager = itemDataManager

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.AnythingOfType("newsman.Event")).Return()
		s.reporter = r

		si := &mocksearch.IndexManager{}
		si.On("Index", mock.Anything, exampleItem.ID, exampleItem).Return(nil)
		s.search = si

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.AnythingOfType("*models.Item")).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, r, itemDataManager, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with no rows fetching item", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*models.Item)(nil), sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error updating item", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("GetItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(exampleItem, nil)
		itemDataManager.On("UpdateItem", mock.Anything, mock.AnythingOfType("*models.Item")).Return(errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), updateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})
}

func TestItemsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	exampleUser := fakemodels.BuildFakeUser()
	userIDFetcher := func(_ *http.Request) uint64 {
		return exampleUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(nil)
		s.itemDataManager = itemDataManager

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.AnythingOfType("newsman.Event")).Return()
		s.reporter = r

		si := &mocksearch.IndexManager{}
		si.On("Delete", mock.Anything, exampleItem.ID).Return(nil)
		s.search = si

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.itemCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, mc, r)
	})

	T.Run("with no item in database", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(sql.ErrNoRows)
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error writing to database", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(errors.New("blah"))
		s.itemDataManager = itemDataManager

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager)
	})

	T.Run("with error removing from search index", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleItem.BelongsToUser = exampleUser.ID
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		itemDataManager := &mockmodels.ItemDataManager{}
		itemDataManager.On("ArchiveItem", mock.Anything, exampleItem.ID, exampleUser.ID).Return(nil)
		s.itemDataManager = itemDataManager

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.AnythingOfType("newsman.Event")).Return()
		s.reporter = r

		si := &mocksearch.IndexManager{}
		si.On("Delete", mock.Anything, exampleItem.ID).Return(errors.New("blah"))
		s.search = si

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.itemCounter = mc

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, itemDataManager, mc, r)
	})
}
