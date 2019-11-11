package dbclient

import (
	"context"
	"errors"
	"testing"

	database "gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/noop"
)

func buildTestClient() (*Client, *database.MockDatabase) {
	db := database.BuildMockDatabase()
	c := &Client{
		logger:  noop.ProvideNoopLogger(),
		querier: db,
	}
	return c, db
}

func TestMigrate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(nil)

		c := &Client{querier: mockDB}
		actual := c.Migrate(context.Background())
		assert.NoError(t, actual)
	})

	T.Run("bubbles up errors", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(errors.New("blah"))

		c := &Client{querier: mockDB}
		actual := c.Migrate(context.Background())
		assert.Error(t, actual)
	})
}

func TestIsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("IsReady", mock.Anything).Return(true)

		c := &Client{querier: mockDB}
		c.IsReady(context.Background())
		mockDB.AssertExpectations(t)
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(nil)

		actual, err := ProvideDatabaseClient(context.Background(), nil, mockDB, false, noop.ProvideNoopLogger())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error migrating querier", func(t *testing.T) {
		expected := errors.New("blah")
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(expected)

		x, actual := ProvideDatabaseClient(context.Background(), nil, mockDB, false, noop.ProvideNoopLogger())
		assert.Nil(t, x)
		assert.Error(t, actual)
		assert.Equal(t, expected, actual)
	})
}
