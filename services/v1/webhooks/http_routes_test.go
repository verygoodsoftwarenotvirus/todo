package webhooks

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
)

func TestWebhooksService_List(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			requestingUser.ID,
			mock.Anything,
		).Return(exampleWebhookList, nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			requestingUser.ID,
			mock.Anything,
		).Return((*models.WebhookList)(nil), sql.ErrNoRows)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with error fetching webhooks from database", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			requestingUser.ID,
			mock.Anything,
		).Return((*models.WebhookList)(nil), errors.New("blah"))
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		exampleWebhookList := fakemodels.BuildFakeWebhookList()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			requestingUser.ID,
			mock.Anything,
		).Return(exampleWebhookList, nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}

func TestValidateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		assert.NoError(t, validateWebhook(exampleInput))
	})

	T.Run("with invalid method", func(t *testing.T) {
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleInput.Method = " MEATLOAF "

		assert.Error(t, validateWebhook(exampleInput))
	})

	T.Run("with invalid url", func(t *testing.T) {
		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleInput.URL = "%zzzzz"

		assert.Error(t, validateWebhook(exampleInput))
	})
}

func TestWebhooksService_Create(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(exampleWebhook, nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, mc, wd, ed)
	})

	T.Run("with invalid webhook request", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleWebhook.URL = "%zzzzz"
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

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
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("without input attached", func(t *testing.T) {
		s := buildTestService()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

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

	T.Run("with error creating webhook", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			mock.Anything,
			mock.Anything,
		).Return((*models.Webhook)(nil), errors.New("blah"))
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(exampleWebhook, nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, mc, wd, ed)
	})
}

func TestWebhooksService_Read(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(exampleWebhook, nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with no such webhook in database", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), sql.ErrNoRows)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error fetching webhook from database", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), errors.New("blah"))
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(exampleWebhook, nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}

func TestWebhooksService_Update(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd, ed)
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

	T.Run("with no rows fetching webhook", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), sql.ErrNoRows)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), errors.New("blah"))
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error updating webhook", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("blah"))
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()
		exampleInput := fakemodels.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}

func TestWebhooksService_Archive(T *testing.T) {
	T.Parallel()

	requestingUser := fakemodels.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.Anything).Return()
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(nil)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, mc, wd)
	})

	T.Run("with no webhook in database", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(sql.ErrNoRows)
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService()

		exampleWebhook := fakemodels.BuildFakeWebhook()

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			exampleWebhook.ID,
			requestingUser.ID,
		).Return(errors.New("blah"))
		s.webhookDatabase = wd

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

		mock.AssertExpectationsForObjects(t, wd)
	})
}
