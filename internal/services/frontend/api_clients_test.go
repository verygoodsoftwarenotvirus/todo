package frontend

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	mockrouting "gitlab.com/verygoodsoftwarenotvirus/todo/internal/routing/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_fetchAPIClient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			apiClientIDURLParamKey,
			"API client",
		).Return(func(req *http.Request) uint64 {
			return exampleAPIClient.ID
		})
		s.routeParamManager = rpm

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			exampleAPIClient.ID,
			exampleSessionContextData.Requester.ID,
		).Return(exampleAPIClient, nil)
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.fetchAPIClient(ctx, exampleSessionContextData, req)
		assert.Equal(t, exampleAPIClient, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.fetchAPIClient(ctx, exampleSessionContextData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching apiClient", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			apiClientIDURLParamKey,
			"API client",
		).Return(func(req *http.Request) uint64 {
			return exampleAPIClient.ID
		})
		s.routeParamManager = rpm

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			exampleAPIClient.ID,
			exampleSessionContextData.Requester.ID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.fetchAPIClient(ctx, exampleSessionContextData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})
}

func TestService_buildAPIClientEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			apiClientIDURLParamKey,
			"API client",
		).Return(func(req *http.Request) uint64 {
			return exampleAPIClient.ID
		})
		s.routeParamManager = rpm

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			exampleAPIClient.ID,
			exampleSessionContextData.Requester.ID,
		).Return(exampleAPIClient, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			apiClientIDURLParamKey,
			"API client",
		).Return(func(req *http.Request) uint64 {
			return exampleAPIClient.ID
		})
		s.routeParamManager = rpm

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			exampleAPIClient.ID,
			exampleSessionContextData.Requester.ID,
		).Return(exampleAPIClient, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientEditorView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientEditorView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching apiClient", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleAPIClient := fakes.BuildFakeAPIClient()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamIDFetcher",
			mock.Anything,
			apiClientIDURLParamKey,
			"API client",
		).Return(func(req *http.Request) uint64 {
			return exampleAPIClient.ID
		})
		s.routeParamManager = rpm

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClientByDatabaseID",
			testutil.ContextMatcher,
			exampleAPIClient.ID,
			exampleSessionContextData.Requester.ID,
		).Return((*types.APIClient)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB, rpm)
	})
}

func TestService_fetchAPIClients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		exampleAPIClientList := fakes.BuildFakeAPIClientList()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.fetchAPIClients(ctx, exampleSessionContextData, req)
		assert.Equal(t, exampleAPIClientList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.fetchAPIClients(ctx, exampleSessionContextData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		actual, err := s.fetchAPIClients(ctx, exampleSessionContextData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAPIClientsTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientsTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAPIClientList := fakes.BuildFakeAPIClientList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAPIClientList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientsTableView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return nil, errors.New("blah")
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientsTableView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.APIClientDataManager.On(
			"GetAPIClients",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.APIClientList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api_clients", nil)

		s.buildAPIClientsTableView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
