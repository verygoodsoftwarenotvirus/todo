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
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	mocknewsman "gitlab.com/verygoodsoftwarenotvirus/newsman/mock"
)

func TestItemsService_List(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.ItemList{
			Items: []models.Item{
				{
					ID: 123,
				},
			},
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return(expected, nil)
		s.itemDatabase = id

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

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return((*models.ItemList)(nil), sql.ErrNoRows)
		s.itemDatabase = id

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

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with error fetching items from database", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return((*models.ItemList)(nil), errors.New("blah"))
		s.itemDatabase = id

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

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.ItemList{
			Items: []models.Item{},
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItems", mock.Anything, requestingUser.ID, mock.Anything).Return(expected, nil)
		s.itemDatabase = id

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

		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestItemsService_Create(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("CreateItem", mock.Anything, mock.Anything).Return(expected, nil)
		s.itemDatabase = id

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

		exampleInput := &models.ItemCreationInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusCreated)
	})

	T.Run("without input attached", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

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

		s.CreateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusBadRequest)
	})

	T.Run("with error creating item", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("CreateItem", mock.Anything, mock.Anything).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDatabase = id

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

		exampleInput := &models.ItemCreationInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("CreateItem", mock.Anything, mock.Anything).Return(expected, nil)
		s.itemDatabase = id

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

		exampleInput := &models.ItemCreationInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusCreated)
	})
}

func TestItemsService_Read(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return(expected, nil)
		s.itemDatabase = id

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

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with no such item in database", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return((*models.Item)(nil), sql.ErrNoRows)
		s.itemDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusNotFound)
	})

	T.Run("with error fetching item from database", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ReadHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return(expected, nil)
		s.itemDatabase = id

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

		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestItemsService_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return(expected, nil)
		id.On("UpdateItem", mock.Anything, mock.Anything).Return(nil)
		s.itemDatabase = id

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

		exampleInput := &models.ItemUpdateInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
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

		assert.Equal(t, res.Code, http.StatusBadRequest)
	})

	T.Run("with no rows fetching item", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return((*models.Item)(nil), sql.ErrNoRows)
		s.itemDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.ItemUpdateInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusNotFound)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return((*models.Item)(nil), errors.New("blah"))
		s.itemDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.ItemUpdateInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error updating item", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return(expected, nil)
		id.On("UpdateItem", mock.Anything, mock.Anything).Return(errors.New("blah"))
		s.itemDatabase = id

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

		exampleInput := &models.ItemUpdateInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.itemCounter = mc

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("GetItem", mock.Anything, expected.ID, requestingUser.ID).Return(expected, nil)
		id.On("UpdateItem", mock.Anything, mock.Anything).Return(nil)
		s.itemDatabase = id

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

		exampleInput := &models.ItemUpdateInput{
			Name:    expected.Name,
			Details: expected.Details,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestItemsService_Archive(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		r := &mocknewsman.Reporter{}
		r.On("Report", mock.Anything).Return()
		s.reporter = r

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement").Return()
		s.itemCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("ArchiveItem", mock.Anything, expected.ID, requestingUser.ID).Return(nil)
		s.itemDatabase = id

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

		s.ArchiveHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusNoContent)
	})

	T.Run("with no item in database", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("ArchiveItem", mock.Anything, expected.ID, requestingUser.ID).Return(sql.ErrNoRows)
		s.itemDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusNotFound)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService()

		requestingUser := &models.User{ID: 1}
		expected := &models.Item{
			ID: 123,
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.itemIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mockmodels.ItemDataManager{}
		id.On("ArchiveItem", mock.Anything, expected.ID, requestingUser.ID).Return(errors.New("blah"))
		s.itemDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.ArchiveHandler()(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})
}
