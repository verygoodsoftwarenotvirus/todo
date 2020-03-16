package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleID := fake.Uint64()
		expected := &models.User{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleID).Return(expected, nil)

		actual, err := c.GetUser(ctx, exampleID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleUsername := "username"
		expected := &models.User{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserByUsername", mock.Anything, exampleUsername).Return(expected, nil)

		actual, err := c.GetUserByUsername(ctx, exampleUsername)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetUserCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := fake.Uint64()
		filter := models.DefaultQueryFilter()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserCount", mock.Anything, filter).Return(expected, nil)

		actual, err := c.GetUserCount(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()
		expected := fake.Uint64()
		filter := (*models.QueryFilter)(nil)

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserCount", mock.Anything, filter).Return(expected, nil)

		actual, err := c.GetUserCount(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.UserList{}
		filter := models.DefaultQueryFilter()

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, filter).Return(expected, nil)

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		ctx := context.Background()
		expected := &models.UserList{}
		filter := (*models.QueryFilter)(nil)

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, filter).Return(expected, nil)

		actual, err := c.GetUsers(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.UserInput{}
		expected := &models.User{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("CreateUser", mock.Anything, exampleInput).Return(expected, nil)

		actual, err := c.CreateUser(ctx, exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := &models.User{}
		var expected error

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, exampleInput).Return(expected, nil)

		err := c.UpdateUser(ctx, exampleInput)
		assert.NoError(t, err)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		exampleInput := fake.Uint64()
		var expected error

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleInput).Return(expected, nil)

		err := c.ArchiveUser(ctx, exampleInput)
		assert.NoError(t, err)

		mockDB.AssertExpectations(t)
	})
}
