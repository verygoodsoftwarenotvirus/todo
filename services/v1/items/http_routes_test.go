package items

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/encoding/mock"
	mockmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/metrics/mock"
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

	requestingUser := fakemodels.BuildFakeUser()

	userIDFetcher := func(_ *http.Request) uint64 {
		return requestingUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItemList := fakemodels.BuildFakeItemList()

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return(exampleItemList, nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, idm, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return((*models.ItemList)(nil), sql.ErrNoRows)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, idm, ed)
	})

	T.Run("with error fetching items from database", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return((*models.ItemList)(nil), errors.New("blah"))
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItemList := fakemodels.BuildFakeItemList()

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return(exampleItemList, nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ListHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, idm, ed)
	})
}

func TestItemsService_CreateHandler(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	userIDFetcher := func(_ *http.Request) uint64 {
		return requestingUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		idm := &mockmodels.ItemDataManager{}
		idm.On("CreateItem", mock.Anything, mock.Anything).Return(exampleItem, nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, mc, r, idm, ed)
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

		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with error creating item", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		idm := &mockmodels.ItemDataManager{}
		idm.On("CreateItem", mock.Anything, mock.Anything).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemCreationInputFromItem(exampleItem)

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		idm := &mockmodels.ItemDataManager{}
		idm.On("CreateItem", mock.Anything, mock.Anything).Return(exampleItem, nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)

		assert.Equal(t, http.StatusCreated, res.Code)

		mock.AssertExpectationsForObjects(t, mc, r, idm, ed)
	})
}

func TestItemsService_ExistenceHandler(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	userIDFetcher := func(_ *http.Request) uint64 {
		return requestingUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("ItemExists", mock.Anything, exampleItem.ID, requestingUser.ID).Return(true, nil)
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with no such item in database", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("ItemExists", mock.Anything, exampleItem.ID, requestingUser.ID).Return(false, sql.ErrNoRows)
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler()(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error fetching item from database", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("ItemExists", mock.Anything, exampleItem.ID, requestingUser.ID).Return(false, errors.New("blah"))
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ExistenceHandler()(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})
}

func TestItemsService_ReadHandler(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	userIDFetcher := func(_ *http.Request) uint64 {
		return requestingUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(exampleItem, nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, idm, ed)
	})

	T.Run("with no such item in database", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return((*models.Item)(nil), sql.ErrNoRows)
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error fetching item from database", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(exampleItem, nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, idm, ed)
	})
}

func TestItemsService_UpdateHandler(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	userIDFetcher := func(_ *http.Request) uint64 {
		return requestingUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(exampleItem, nil)
		idm.On("UpdateItem", mock.Anything, mock.Anything).Return(nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, r, idm, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		s := buildTestService()

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.UpdateHandler()(res, req)

		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("with no rows fetching item", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return((*models.Item)(nil), sql.ErrNoRows)
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error updating item", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(exampleItem, nil)
		idm.On("UpdateItem", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		exampleInput := fakemodels.BuildFakeItemUpdateInputFromItem(exampleItem)

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		idm := &mockmodels.ItemDataManager{}
		idm.On("GetItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(exampleItem, nil)
		idm.On("UpdateItem", mock.Anything, mock.Anything).Return(nil)
		s.itemDatabase = idm

		ed := &mockencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, r, idm, ed)
	})
}

func TestItemsService_ArchiveHandler(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	userIDFetcher := func(_ *http.Request) uint64 {
		return requestingUser.ID
	}

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.itemCounter = mc

		idm := &mockmodels.ItemDataManager{}
		idm.On("ArchiveItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(nil)
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler()(res, req)

		assert.Equal(t, http.StatusNoContent, res.Code)

		mock.AssertExpectationsForObjects(t, mc, r, idm)
	})

	T.Run("with no item in database", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("ArchiveItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(sql.ErrNoRows)
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler()(res, req)

		assert.Equal(t, http.StatusNotFound, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService()
		s.userIDFetcher = userIDFetcher

		exampleItem := fakemodels.BuildFakeItem()
		s.itemIDFetcher = func(req *http.Request) uint64 {
			return exampleItem.ID
		}

		idm := &mockmodels.ItemDataManager{}
		idm.On("ArchiveItem", mock.Anything, exampleItem.ID, requestingUser.ID).Return(errors.New("blah"))
		s.itemDatabase = idm

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler()(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, idm)
	})
}
