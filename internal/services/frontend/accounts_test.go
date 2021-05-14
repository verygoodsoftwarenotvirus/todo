package frontend

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_fetchAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.ID,
		).Return(exampleAccount, nil)
		s.dataStore = mockDB

		actual, err := s.fetchAccount(ctx, exampleSessionContextData)
		assert.Equal(t, exampleAccount, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		actual, err := s.fetchAccount(ctx, exampleSessionContextData)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching account", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.ID,
		).Return((*types.Account)(nil), errors.New("blah"))
		s.dataStore = mockDB

		actual, err := s.fetchAccount(ctx, exampleSessionContextData)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAccountEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.ID,
		).Return(exampleAccount, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildAccountEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.ID,
		).Return(exampleAccount, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildAccountEditorView(false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildAccountEditorView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching item", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutil.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.ID,
		).Return((*types.Account)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/items", nil)

		s.buildAccountEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_fetchAccounts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		exampleAccountList := fakes.BuildFakeAccountList()

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		actual, err := s.fetchAccounts(ctx, exampleSessionContextData, req)
		assert.Equal(t, exampleAccountList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		actual, err := s.fetchAccounts(ctx, exampleSessionContextData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		actual, err := s.fetchAccounts(ctx, exampleSessionContextData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAccountsTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAccountList := fakes.BuildFakeAccountList()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.buildAccountsTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAccountList := fakes.BuildFakeAccountList()
		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.buildAccountsTableView(false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.buildAccountsTableView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.buildAccountsTableView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
