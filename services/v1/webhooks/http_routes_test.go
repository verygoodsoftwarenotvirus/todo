package webhooks

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	mencoding "gitlab.com/verygoodsoftwarenotvirus/todo/internal/encoding/v1/mock"
	mmetrics "gitlab.com/verygoodsoftwarenotvirus/todo/internal/metrics/v1/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	mmodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/mock"

	mockman "gitlab.com/verygoodsoftwarenotvirus/newsman/mock"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebhooksService_List(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
		).Return(expected, nil)

		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.List(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
		).Return((*models.WebhookList)(nil), sql.ErrNoRows)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.List(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})

	T.Run("with error fetching webhooks from database", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
		).Return((*models.WebhookList)(nil), errors.New("blah"))
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.List(res, req)

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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhooks",
			mock.Anything,
			mock.Anything,
			requestingUser.ID,
		).Return(expected, nil)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.List(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestValidateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		exampleInput := &models.WebhookInput{
			Method: http.MethodPost,
			URL:    "https://todo.verygoodsoftwarenotvirus.ru",
		}

		expected := http.StatusOK
		actual, err := validateWebhook(exampleInput)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid method", func(t *testing.T) {
		exampleInput := &models.WebhookInput{
			Method: ` MEATLOAF `,
			URL:    "https://todo.verygoodsoftwarenotvirus.ru",
		}

		expected := http.StatusBadRequest
		actual, err := validateWebhook(exampleInput)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	T.Run("with invalid url", func(t *testing.T) {
		exampleInput := &models.WebhookInput{
			Method: http.MethodPost,
			URL:    "%zzzzz",
		}

		expected := http.StatusBadRequest
		actual, err := validateWebhook(exampleInput)

		assert.Error(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestWebhooksService_Create(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		r := &eventMan{
			Reporter: &mockman.Reporter{},
		}
		r.Reporter.On("Report", mock.Anything).Return()
		r.On("TuneIn", mock.Anything).Return()
		s.newsman = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mmodels.WebhookDataManager{}
		id.On(
			"CreateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(expected, nil)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Create(res, req)

		assert.Equal(t, res.Code, http.StatusCreated)
	})

	T.Run("with invalid webhook request", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		r := &eventMan{
			Reporter: &mockman.Reporter{},
		}
		r.Reporter.On("Report", mock.Anything).Return()
		r.On("TuneIn", mock.Anything).Return()
		s.newsman = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mmodels.WebhookDataManager{}
		id.On(
			"CreateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(expected, nil)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Method: http.MethodPost,
			URL:    "%zzzzz",
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Create(res, req)

		assert.Equal(t, res.Code, http.StatusBadRequest)
	})

	T.Run("without input attached", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Create(res, req)

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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"CreateWebhook",
			mock.Anything,
			mock.Anything,
		).Return((*models.Webhook)(nil), errors.New("blah"))
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Create(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		r := &eventMan{Reporter: &mockman.Reporter{}}
		r.Reporter.On("Report", mock.Anything).Return()
		r.On("TuneIn", mock.Anything).Return()
		s.newsman = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		id := &mmodels.WebhookDataManager{}
		id.On(
			"CreateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(expected, nil)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Create(res, req)

		assert.Equal(t, res.Code, http.StatusCreated)
	})
}

func TestWebhooksService_Read(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Read(res, req)

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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), sql.ErrNoRows)
		s.webhookDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Read(res, req)

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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), errors.New("blah"))
		s.webhookDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Read(res, req)

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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Read(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestWebhooksService_Update(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID: 123, Name: "name",
		}

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		r := &eventMan{Reporter: &mockman.Reporter{}}
		r.Reporter.On("Report", mock.Anything).Return()
		s.newsman = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mmodels.WebhookDataManager{}

		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)

		id.On(
			"UpdateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(nil)

		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Update(res, req)

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

		s.Update(res, req)

		assert.Equal(t, res.Code, http.StatusBadRequest)
	})

	T.Run("with no rows fetching webhook", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID: 123, Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mmodels.WebhookDataManager{}

		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), sql.ErrNoRows)

		s.webhookDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Update(res, req)

		assert.Equal(t, res.Code, http.StatusNotFound)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID: 123, Name: "name",
		}

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mmodels.WebhookDataManager{}

		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return((*models.Webhook)(nil), errors.New("blah"))

		s.webhookDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Update(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error updating webhook", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID: 123, Name: "name",
		}

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		r := &eventMan{Reporter: &mockman.Reporter{}}
		r.Reporter.On("Report", mock.Anything).Return()
		s.newsman = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mmodels.WebhookDataManager{}

		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)

		id.On(
			"UpdateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(errors.New("blah"))

		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Update(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})

	T.Run("with error encoding response", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID: 123, Name: "name",
		}

		mc := &mmetrics.UnitCounter{}
		mc.On("Increment", mock.Anything)
		s.webhookCounter = mc

		r := &eventMan{Reporter: &mockman.Reporter{}}
		r.Reporter.On("Report", mock.Anything).Return()
		s.newsman = r

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mmodels.WebhookDataManager{}

		id.On(
			"GetWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(expected, nil)

		id.On(
			"UpdateWebhook",
			mock.Anything,
			mock.Anything,
		).Return(nil)

		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(errors.New("blah"))
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		exampleInput := &models.WebhookInput{
			Name: expected.Name,
		}
		req = req.WithContext(context.WithValue(req.Context(), MiddlewareCtxKey, exampleInput))

		s.Update(res, req)

		assert.Equal(t, res.Code, http.StatusOK)
	})
}

func TestWebhooksService_Delete(T *testing.T) {
	T.Parallel()

	T.Run("normal operation", func(t *testing.T) {
		s := buildTestService()
		requestingUser := &models.User{ID: 1}
		expected := &models.Webhook{
			ID:   123,
			Name: "name",
		}

		r := &eventMan{Reporter: &mockman.Reporter{}}
		r.Reporter.On("Report", mock.Anything).Return()
		s.newsman = r

		mc := &mmetrics.UnitCounter{}
		mc.On("Decrement").Return()
		s.webhookCounter = mc

		s.userIDFetcher = func(req *http.Request) uint64 {
			return requestingUser.ID
		}

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return expected.ID
		}

		id := &mmodels.WebhookDataManager{}
		id.On(
			"DeleteWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(nil)
		s.webhookDatabase = id

		ed := &mencoding.EncoderDecoder{}
		ed.On("EncodeResponse", mock.Anything, mock.Anything).
			Return(nil)
		s.encoderDecoder = ed

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Delete(res, req)

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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"DeleteWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(sql.ErrNoRows)
		s.webhookDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Delete(res, req)

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

		id := &mmodels.WebhookDataManager{}
		id.On(
			"DeleteWebhook",
			mock.Anything,
			expected.ID,
			requestingUser.ID,
		).Return(errors.New("blah"))
		s.webhookDatabase = id

		res := httptest.NewRecorder()
		req, err := http.NewRequest(
			http.MethodGet,
			"http://todo.verygoodsoftwarenotvirus.ru",
			nil,
		)
		require.NotNil(t, req)
		require.NoError(t, err)

		s.Delete(res, req)

		assert.Equal(t, res.Code, http.StatusInternalServerError)
	})
}
