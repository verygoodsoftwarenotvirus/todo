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
	mockmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebhooksService_List(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.WebhookList{
			Webhooks: []models.Webhook{
				{
					ID:   123,
					Name: "name",
				},
			},
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
		).Return(expected, nil)
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
		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
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
		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with error fetching webhooks from database", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
		).Return((*models.WebhookList)(nil), errors.New("blah"))
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
		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.WebhookList{
			Webhooks: []models.Webhook{},
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
		).Return(expected, nil)
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
		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestValidateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleInput := &models.WebhookCreationInput{
			Method: http.MethodPost,
			URL:    "https://todo.verygoodsoftwarenotvirus.ru",
		}

		assert.NoError(t, validateWebhook(exampleInput))
	})

	T.Run("with invalid method", func(t *testing.T) {
		exampleInput := &models.WebhookCreationInput{
			Method: " MEATLOAF ",
			URL:    "https://todo.verygoodsoftwarenotvirus.ru",
		}

		assert.Error(t, validateWebhook(exampleInput))
	})

	T.Run("with invalid url", func(t *testing.T) {
		exampleInput := &models.WebhookCreationInput{
			Method: http.MethodPost,
			URL:    "%zzzzz",
		}

		assert.Error(t, validateWebhook(exampleInput))
	})
}

func TestWebhooksService_Create(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

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
		).Return(expected, nil)
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

		exampleInput := &models.WebhookCreationInput{
			Name:   expected.Name,
			Method: http.MethodPatch,
		}
		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusCreated)
	})

	T.Run("with invalid webhook request", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

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
		).Return(expected, nil)
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

		exampleInput := &models.WebhookCreationInput{
			Method: http.MethodPost,
			URL:    "%zzzzz",
		}
		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusBadRequest)
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

	T.Run("with error creating webhook", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

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

		exampleInput := &models.WebhookCreationInput{
			Method: http.MethodPatch,
			Name:   expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

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
		).Return(expected, nil)
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

		exampleInput := &models.WebhookCreationInput{
			Method: http.MethodPatch,
			Name:   expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), CreateMiddlewareCtxKey, exampleInput))

		s.CreateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusCreated)
	})
}

func TestWebhooksService_Read(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)
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
		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with no such webhook in database", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
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
		assert.Equal(t, res.Code, http.StatusNotFound)
	})

	T.Run("with error fetching webhook from database", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
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
		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)
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
		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestWebhooksService_Update(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)

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

		exampleInput := &models.WebhookUpdateInput{
			Name: expected.Name,
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

	T.Run("with no rows fetching webhook", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
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

		exampleInput := &models.WebhookUpdateInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusNotFound)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
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

		exampleInput := &models.WebhookUpdateInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error updating webhook", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)

		wd.On(
			"UpdateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("blah"))
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

		exampleInput := &models.WebhookUpdateInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)

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

		exampleInput := &models.WebhookUpdateInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), UpdateMiddlewareCtxKey, exampleInput))

		s.UpdateHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestWebhooksService_Archive(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement").Return()
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
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

		s.ArchiveHandler()(res, req)
		assert.Equal(t, res.Code, http.StatusNoContent)
	})

	T.Run("with no webhook in database", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			expected.ID,
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
		assert.Equal(t, res.Code, http.StatusNotFound)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		wd := &mockmodels.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			expected.ID,
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
		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})
}
