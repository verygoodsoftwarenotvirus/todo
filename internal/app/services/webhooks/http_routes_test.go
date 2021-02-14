package webhooks

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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	mocktypes "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebhooksService_List(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return types.SessionInfoFromUser(exampleUser), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhookList := fakes.BuildFakeWebhookList()

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.WebhookList{}))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with no rows returned", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), sql.ErrNoRows)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.WebhookList{}))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with error fetching webhooks from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhooks",
			mock.Anything,
			exampleUser.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}

func TestValidateWebhook(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		assert.NoError(t, validateWebhook(exampleInput))
	})

	T.Run("with invalid method", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleInput.Method = " MEATLOAF "

		assert.Error(t, validateWebhook(exampleInput))
	})

	T.Run("with invalid url", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)
		exampleInput.URL = "%zzzzz"

		assert.Error(t, validateWebhook(exampleInput))
	})
}

func TestWebhooksService_Create(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return types.SessionInfoFromUser(exampleUser), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		mc := &mockmetrics.UnitCounter{}
		mc.On("Increment", mock.MatchedBy(testutil.ContextMatcher()))
		s.webhookCounter = mc

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			mock.Anything,
			mock.IsType(&types.WebhookCreationInput{}),
		).Return(exampleWebhook, nil)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponseWithStatus", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Webhook{}), http.StatusCreated)
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

		mock.AssertExpectationsForObjects(t, mc, wd, ed)
	})

	T.Run("with invalid webhook request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleWebhook.URL = "%zzzzz"
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On(
			"EncodeErrorResponse",
			mock.Anything,
			mock.Anything,
			fmt.Sprintf(`invalid url provided: parse %q: invalid URL escape "%%zz"`, exampleWebhook.URL),
			http.StatusBadRequest,
		)
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
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("without input attached", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

	T.Run("with error creating webhook", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookCreationInputFromWebhook(exampleWebhook)

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"CreateWebhook",
			mock.Anything,
			mock.IsType(&types.WebhookCreationInput{}),
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}

func TestWebhooksService_Read(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return types.SessionInfoFromUser(exampleUser), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return(exampleWebhook, nil)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Webhook{}))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with no such webhook in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with error fetching webhook from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}

func TestWebhooksService_Update(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return types.SessionInfoFromUser(exampleUser), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return(exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			mock.Anything,
			mock.IsType(&types.Webhook{}),
		).Return(nil)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()), mock.IsType(&types.Webhook{}))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("without update input", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeInvalidInputResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

	T.Run("with no rows fetching webhook", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return((*types.Webhook)(nil), sql.ErrNoRows)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with error updating webhook", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID
		exampleInput := fakes.BuildFakeWebhookUpdateInputFromWebhook(exampleWebhook)

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"GetWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return(exampleWebhook, nil)

		wd.On(
			"UpdateWebhook",
			mock.Anything,
			mock.IsType(&types.Webhook{}),
		).Return(errors.New("blah"))
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}

func TestWebhooksService_Archive(T *testing.T) {
	T.Parallel()

	exampleUser := fakes.BuildFakeUser()
	sessionInfoFetcher := func(_ *http.Request) (*types.SessionInfo, error) {
		return types.SessionInfoFromUser(exampleUser), nil
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID

		mc := &mockmetrics.UnitCounter{}
		mc.On("Decrement", mock.MatchedBy(testutil.ContextMatcher())).Return()
		s.webhookCounter = mc

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return(nil)
		s.webhookDataManager = wd

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

		mock.AssertExpectationsForObjects(t, mc, wd)
	})

	T.Run("with no webhook in database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return(sql.ErrNoRows)
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeNotFoundResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})

	T.Run("with error reading from database", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		s := buildTestService()
		s.sessionInfoFetcher = sessionInfoFetcher

		exampleWebhook := fakes.BuildFakeWebhook()
		exampleWebhook.BelongsToUser = exampleUser.ID

		s.webhookIDFetcher = func(req *http.Request) uint64 {
			return exampleWebhook.ID
		}

		wd := &mocktypes.WebhookDataManager{}
		wd.On(
			"ArchiveWebhook",
			mock.Anything,
			exampleWebhook.ID,
			exampleUser.ID,
		).Return(errors.New("blah"))
		s.webhookDataManager = wd

		ed := mockencoding.NewMockEncoderDecoder()
		ed.On("EncodeUnspecifiedInternalServerErrorResponse", mock.MatchedBy(testutil.ContextMatcher()), mock.MatchedBy(testutil.ResponseWriterMatcher()))
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

		mock.AssertExpectationsForObjects(t, wd, ed)
	})
}
