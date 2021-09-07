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

func TestService_fetchAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		s.service.apiClientIDFetcher = func(*http.Request) string {
			return exampleAPIClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			s.sessionCtxData.Requester.UserID,
		).Return(exampleAPIClient, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.service.fetchAPIClient(s.ctx, s.sessionCtxData, req)
		assert.Equal(t, exampleAPIClient, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.service.fetchAPIClient(s.ctx, s.sessionCtxData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching apiClient", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		s.service.apiClientIDFetcher = func(*http.Request) string {
			return exampleAPIClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			s.sessionCtxData.Requester.UserID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.service.fetchAPIClient(s.ctx, s.sessionCtxData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAPIClientEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		s.service.apiClientIDFetcher = func(*http.Request) string {
			return exampleAPIClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			s.sessionCtxData.Requester.UserID,
		).Return(exampleAPIClient, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		s.service.apiClientIDFetcher = func(*http.Request) string {
			return exampleAPIClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			s.sessionCtxData.Requester.UserID,
		).Return(exampleAPIClient, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientEditorView(false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientEditorView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching apiClient", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClient := fakes.BuildFakeAPIClient()
		s.service.apiClientIDFetcher = func(*http.Request) string {
			return exampleAPIClient.ID
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutils.ContextMatcher,
			exampleAPIClient.ID,
			s.sessionCtxData.Requester.UserID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_fetchAPIClients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.service.fetchAPIClients(s.ctx, s.sessionCtxData, req)
		assert.Equal(t, exampleAPIClientList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.service.fetchAPIClients(s.ctx, s.sessionCtxData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.service.fetchAPIClients(s.ctx, s.sessionCtxData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAPIClientsTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientsTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientsTableView(false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientsTableView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.service.buildAPIClientsTableView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
