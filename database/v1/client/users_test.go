package dbclient

import (
	"context"
	"testing"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleID := uint64(123)
		expected := &models.User{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUser", mock.Anything, exampleID).Return(expected, nil)

		actual, err := c.GetUser(context.Background(), exampleID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetUserByUsername(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleUsername := "username"
		expected := &models.User{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserByUsername", mock.Anything, exampleUsername).Return(expected, nil)

		actual, err := c.GetUserByUsername(context.Background(), exampleUsername)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetUserCount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserCount", mock.Anything, models.DefaultQueryFilter()).Return(expected, nil)

		actual, err := c.GetUserCount(context.Background(), models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		expected := uint64(123)

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUserCount", mock.Anything, (*models.QueryFilter)(nil)).Return(expected, nil)

		actual, err := c.GetUserCount(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_GetUsers(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		expected := &models.UserList{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, models.DefaultQueryFilter()).Return(expected, nil)

		actual, err := c.GetUsers(context.Background(), models.DefaultQueryFilter())
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})

	T.Run("with nil filter", func(t *testing.T) {
		expected := &models.UserList{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("GetUsers", mock.Anything, (*models.QueryFilter)(nil)).Return(expected, nil)

		actual, err := c.GetUsers(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_CreateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleInput := &models.UserInput{}
		expected := &models.User{}

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("CreateUser", mock.Anything, exampleInput).Return(expected, nil)

		actual, err := c.CreateUser(context.Background(), exampleInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_UpdateUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleInput := &models.User{}
		var expected error

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("UpdateUser", mock.Anything, exampleInput).Return(expected, nil)

		err := c.UpdateUser(context.Background(), exampleInput)
		assert.NoError(t, err)

		mockDB.AssertExpectations(t)
	})
}

func TestClient_ArchiveUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		exampleInput := uint64(123)
		var expected error

		c, mockDB := buildTestClient()
		mockDB.UserDataManager.On("ArchiveUser", mock.Anything, exampleInput).Return(expected, nil)

		err := c.ArchiveUser(context.Background(), exampleInput)
		assert.NoError(t, err)

		mockDB.AssertExpectations(t)
	})
}
