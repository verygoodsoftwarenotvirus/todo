package frontend

import (
	"errors"
	"fmt"
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

func TestService_fetchUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutils.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		actual, err := s.service.fetchUsers(s.ctx, req)
		assert.Equal(t, exampleUserList, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with fake mode", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)
		s.service.useFakeData = true

		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		actual, err := s.service.fetchUsers(s.ctx, req)
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutils.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.UserList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		actual, err := s.service.fetchUsers(s.ctx, req)
		assert.Nil(t, actual)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestService_buildUsersTableView(T *testing.T) {
	T.Parallel()

	T.Run("with base template but not for search", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutils.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.service.buildUsersTableView(true, false)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with base template and for search", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleQuery := "whatever"
		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"SearchForUsersByUsername",
			testutils.ContextMatcher,
			exampleQuery,
		).Return(exampleUserList.Users, nil)
		s.service.dataStore = mockDB

		uri := fmt.Sprintf("/users?%s=%s", types.SearchQueryKey, exampleQuery)
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, uri, nil)

		s.service.buildUsersTableView(true, true)(res, req)

		assert.Equal(t, http.StatusOK, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with base template and for search and error performing search", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleQuery := "whatever"

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"SearchForUsersByUsername",
			testutils.ContextMatcher,
			exampleQuery,
		).Return([]*types.User(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		uri := fmt.Sprintf("/users?%s=%s", types.SearchQueryKey, exampleQuery)
		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, uri, nil)

		s.service.buildUsersTableView(true, true)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("without base template but for search", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		exampleUserList := fakes.BuildFakeUserList()

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutils.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return(exampleUserList, nil)
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.service.buildUsersTableView(false, false)(res, req)

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
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.service.buildUsersTableView(true, false)(res, req)

		assert.Equal(t, unauthorizedRedirectResponseCode, res.Code)
	})

	T.Run("with error fetching data", func(t *testing.T) {
		t.Parallel()

		s := buildTestHelper(t)

		mockDB := database.BuildMockDatabase()
		mockDB.UserDataManager.On(
			"GetUsers",
			testutils.ContextMatcher,
			mock.IsType(&types.QueryFilter{}),
		).Return((*types.UserList)(nil), errors.New("blah"))
		s.service.dataStore = mockDB

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users", nil)

		s.service.buildUsersTableView(true, false)(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Code)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
