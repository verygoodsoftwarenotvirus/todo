package dbclient

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)

		actual, err := c.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetUserWithUnverifiedTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserWithUnverifiedTwoFactorSecret", mock.Anything, exampleUser.ID).Return(exampleUser, nil)

		actual, err := c.GetUserWithUnverifiedTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_VerifyUserTwoFactorSecret(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("VerifyUserTwoFactorSecret", mock.Anything, exampleUser.ID).Return(nil)

		err := c.VerifyUserTwoFactorSecret(ctx, exampleUser.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserByUsername", mock.Anything, exampleUser.Username).Return(exampleUser, nil)

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_SearchForUsersByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUsername := fakes.BuildFakeUser().Username
		exampleUserList := fakes.BuildFakeUserList().Users

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("SearchForUsersByUsername", mock.Anything, exampleUsername).Return(exampleUserList, nil)

		actual, err := c.SearchForUsersByUsername(ctx, exampleUsername)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetAllUsersCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleCount := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetAllUsersCount", mock.Anything).Return(exampleCount, nil)

		actual, err := c.GetAllUsersCount(ctx)
		assert.NoError(t, err)
		assert.Equal(t, exampleCount, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUserList := fakes.BuildFakeUserList()
		filter := types.DefaultQueryFilter()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, filter).Return(exampleUserList, nil)

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUserList := fakes.BuildFakeUserList()
		filter := (*types.QueryFilter)(nil)

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, filter).Return(exampleUserList, nil)

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := fakes.BuildFakeUserDataStoreCreationInputFromUser(exampleUser)

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("CreateUser", mock.Anything, exampleInput).Return(exampleUser, nil)

		actual, err := c.CreateUser(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		var expected error

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, exampleUser).Return(expected)

		err := c.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_UpdateUserPassword(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		var expected error

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("UpdateUserPassword", mock.Anything, exampleUser.ID, exampleUser.HashedPassword).Return(expected)

		err := c.UpdateUserPassword(ctx, exampleUser.ID, exampleUser.HashedPassword)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleUser.ID).Return(nil)

		err := c.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
