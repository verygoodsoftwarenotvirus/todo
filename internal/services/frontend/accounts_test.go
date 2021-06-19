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

func TestService_fetchAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAccount := fakes.BuildFakeAccount()

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			s.sessionCtxData.Requester.UserID,
		).Return(exampleAccount, nil)
		s.service.dataStore = mockDB

		actual, err := s.service.fetchAccount(s.ctx, s.sessionCtxData)
		assert.Equal(t, exampleAccount, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		actual, err := s.service.fetchAccount(s.ctx, s.sessionCtxData)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching account", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			s.sessionCtxData.Requester.UserID,
		).Return((*types.Account)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		actual, err := s.service.fetchAccount(s.ctx, s.sessionCtxData)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAccountEditorView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		exampleAccount := fakes.BuildFakeAccount()

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			s.sessionCtxData.Requester.UserID,
		).Return(exampleAccount, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountEditorView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		exampleAccount := fakes.BuildFakeAccount()

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			s.sessionCtxData.Requester.UserID,
		).Return(exampleAccount, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountEditorView(false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountEditorView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching account", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			s.sessionCtxData.ActiveAccountID,
			s.sessionCtxData.Requester.UserID,
		).Return((*types.Account)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountEditorView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_fetchAccounts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAccountList := fakes.BuildFakeAccountList()

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		actual, err := s.service.fetchAccounts(s.ctx, s.sessionCtxData, req)
		assert.Equal(t, exampleAccountList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		actual, err := s.service.fetchAccounts(s.ctx, s.sessionCtxData, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		actual, err := s.service.fetchAccounts(s.ctx, s.sessionCtxData, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAccountsTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAccountList := fakes.BuildFakeAccountList()

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountsTableView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAccountList := fakes.BuildFakeAccountList()

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleAccountList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountsTableView(false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountsTableView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccounts",
			testutils.ContextMatcher,
			s.sessionCtxData.Requester.UserID,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.AccountList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

		s.service.buildAccountsTableView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
