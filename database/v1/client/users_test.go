package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleUser.ID).Return(exampleUser, nil)

		actual, err := c.GetUser(ctx, exampleUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserByUsername", mock.Anything, exampleUser.Username).Return(exampleUser, nil)

		actual, err := c.GetUserByUsername(ctx, exampleUser.Username)
		assert.NoError(t, err)
		assert.Equal(t, exampleUser, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUserList := fakemodels.BuildFakeUserList()
		filter := models.DefaultQueryFilter()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, filter).Return(exampleUserList, nil)

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, exampleUserList, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()
		exampleUserList := fakemodels.BuildFakeUserList()
		filter := (*models.QueryFilter)(nil)

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
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()
		exampleInput := fakemodels.BuildFakeUserDatabaseCreationInputFromUser(exampleUser)

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
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()
		var expected error

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, exampleUser).Return(expected)

		err := c.UpdateUser(ctx, exampleUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestClient_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUser := fakemodels.BuildFakeUser()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleUser.ID).Return(nil)

		err := c.ArchiveUser(ctx, exampleUser.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
