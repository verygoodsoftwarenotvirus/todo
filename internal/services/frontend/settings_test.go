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
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_buildUserSettingsView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUser := fakes.BuildFakeUser()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			exampleSessionContextData.Requester.UserID,
		).Return(exampleUser, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildUserSettingsView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUser := fakes.BuildFakeUser()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			exampleSessionContextData.Requester.UserID,
		).Return(exampleUser, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildUserSettingsView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		s.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildUserSettingsView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching user from database", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUser",
			testutils.ContextMatcher,
			exampleSessionContextData.Requester.UserID,
		).Return((*types.User)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildUserSettingsView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAccountSettingsView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.UserID,
		).Return(exampleAccount, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAccountSettingsView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.UserID,
		).Return(exampleAccount, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAccountSettingsView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		s.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAccountSettingsView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching account from database", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleAccount := fakes.BuildFakeAccount()
		exampleSessionContextData := fakes.BuildFakeSessionContextDataForAccount(exampleAccount)
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.AccountDataManager.On(
			"GetAccount",
			testutils.ContextMatcher,
			exampleSessionContextData.ActiveAccountID,
			exampleSessionContextData.Requester.UserID,
		).Return((*types.Account)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAccountSettingsView(true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildAdminSettingsView(T *testing.T) {
	T.Parallel()

	T.Run("with base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		exampleSessionContextData.Requester.ServicePermissions = authorization.NewServiceRolePermissionChecker(authorization.ServiceAdminRole.String())
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAdminSettingsView(true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("without base template", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		exampleSessionContextData.Requester.ServicePermissions = authorization.NewServiceRolePermissionChecker(authorization.ServiceAdminRole.String())
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAdminSettingsView(false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)
	})

	T.Run("with error fetching session context data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		s.service.sessionContextDataFetcher = testutils.BrokenSessionContextDataFetcher

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAdminSettingsView(true)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with non-admin user", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleSessionContextData := fakes.BuildFakeSessionContextData()

		exampleSessionContextData.Requester.ServicePermissions = authorization.NewServiceRolePermissionChecker(authorization.ServiceUserRole.String())
		s.service.sessionContextDataFetcher = func(*http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/whatever", nil)

		s.service.buildAdminSettingsView(true)(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Code)
	})
}
