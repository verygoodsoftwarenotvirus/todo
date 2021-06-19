package frontend

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_fetchWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleWebhook := fakes.BuildFakeWebhook()
		s.service.webhookIDFetcher = func(*http.Request) uint64 { return exampleWebhook.ID }

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleWebhook, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.service.fetchWebhook(s.ctx, s.sessionCtxData, req)
		assert.Equal(t, exampleWebhook, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.service.fetchWebhook(s.ctx, s.sessionCtxData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleWebhook := fakes.BuildFakeWebhook()
		s.service.webhookIDFetcher = func(*http.Request) uint64 { return exampleWebhook.ID }

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.service.fetchWebhook(s.ctx, s.sessionCtxData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildWebhookEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		exampleWebhook := fakes.BuildFakeWebhook()
		s.service.webhookIDFetcher = func(*http.Request) uint64 { return exampleWebhook.ID }

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleWebhook, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhookEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		exampleWebhook := fakes.BuildFakeWebhook()
		s.service.webhookIDFetcher = func(*http.Request) uint64 { return exampleWebhook.ID }

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return(exampleWebhook, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhookEditorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhookEditorView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching webhook", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		exampleWebhook := fakes.BuildFakeWebhook()
		s.service.webhookIDFetcher = func(*http.Request) uint64 { return exampleWebhook.ID }

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhook",
			testutils.ContextMatcher,
			exampleWebhook.ID,
			s.sessionCtxData.ActiveAccountID,
		).Return((*types.Webhook)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhookEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_fetchWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleWebhookList := fakes.BuildFakeWebhookList()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.service.fetchWebhooks(s.ctx, s.sessionCtxData, req)
		assert.Equal(t, exampleWebhookList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.service.fetchWebhooks(s.ctx, s.sessionCtxData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		actual, err := s.service.fetchWebhooks(s.ctx, s.sessionCtxData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildWebhooksTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleWebhookList := fakes.BuildFakeWebhookList()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhooksTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleWebhookList := fakes.BuildFakeWebhookList()

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleWebhookList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhooksTableView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhooksTableView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.WebhookDataManager.On(
			"GetWebhooks",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.WebhookList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)

		s.service.buildWebhooksTableView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
