package frontend

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/authorization"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_buildUserSettingsView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
		).Return(exampleUser, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildUserSettingsView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUser := fakes.BuildFakeUser()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
		).Return(exampleUser, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildUserSettingsView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildUserSettingsView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching user from database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutil.ContextMatcher,
			exampleSessionContextData.Requester.ID,
		).Return((*types.User)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildUserSettingsView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAccountSettingsView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
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
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAccountSettingsView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
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
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAccountSettingsView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAccountSettingsView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching account from database", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
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
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAccountSettingsView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAdminSettingsView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAdminSettingsView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAdminSettingsView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		s.sessionContextDataFetcher = testutil.BrokenSessionContextDataFetcher

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAdminSettingsView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		exampleSessionContextData.Requester.ServiceRole = authorization.ServiceUserRole
		exampleSessionContextData.Requester.ServiceAdminPermission = 0
		s.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.buildAdminSettingsView(true)(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}
