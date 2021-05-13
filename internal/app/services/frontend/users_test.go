package frontend

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_fetchUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()
		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		actual, err := s.fetchUsers(ctx, req)
		assert.Equal(t, exampleUserList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)
		s.useFakeData = true

		ctx := context.Background()

		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		actual, err := s.fetchUsers(ctx, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		ctx := context.Background()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.UserList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		actual, err := s.fetchUsers(ctx, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildUsersTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template but not for search", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUserList := fakes.BuildFakeUserList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.buildUsersTableView(true, false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with base template and for search", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleQuery := "whatever"
		exampleUserList := fakes.BuildFakeUserList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"SearchForUsersByUsername",
			testutil.ContextMatcher,
			exampleQuery,
		).Return(exampleUserList.Users, nil)
		s.dataStore = mockDB

		uri := fmt.Sprintf("/users?%s=%s", types.SearchQueryKey, exampleQuery)
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, uri, nil)

		s.buildUsersTableView(true, true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with base template and for search and error performing search", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleQuery := "whatever"
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"SearchForUsersByUsername",
			testutil.ContextMatcher,
			exampleQuery,
		).Return([]*types.User(nil), errors.New("blah"))
		s.dataStore = mockDB

		uri := fmt.Sprintf("/users?%s=%s", types.SearchQueryKey, exampleQuery)
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, uri, nil)

		s.buildUsersTableView(true, true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template but for search", func(t *testing.T) {
		t.Parallel()

		s := buildTestService(t)

		exampleUserList := fakes.BuildFakeUserList()
		exampleSessionContextData := fakes.BuildFakeSessionContextData()
		s.sessionContextDataFetcher = func(req *http.Request) (*types.SessionContextData, error) {
			return exampleSessionContextData, nil
		}

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.buildUsersTableView(false, false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.buildUsersTableView(true, false)(res, req)

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
		mockDB.UserDataManager.On(
			"GetUsers",
			testutil.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.UserList)(nil), errors.New("blah"))
		s.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.buildUsersTableView(true, false)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
