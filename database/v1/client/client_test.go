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
		ctx := context.Background()
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(nil)

		c := &Client{querier: mockDB}
		actual := c.Migrate(ctx)
		assert.NoError(t, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("bubbles up errors", func(t *testing.T) {
		ctx := context.Background()
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(errors.New("blah"))

		c := &Client{querier: mockDB}
		actual := c.Migrate(ctx)
		assert.Error(t, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestIsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		ctx := context.Background()
		mockDB := database.BuildMockDatabase()
		mockDB.On("IsReady", mock.Anything).Return(true)

		c := &Client{querier: mockDB}
		c.IsReady(ctx)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		ctx := context.Background()
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(nil)

		actual, err := ProvideDatabaseClient(ctx, nil, mockDB, false, noop.ProvideNoopLogger())
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error migrating querier", func(t *testing.T) {
		ctx := context.Background()
		expected := errors.New("blah")
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).Return(expected)

		x, actual := ProvideDatabaseClient(ctx, nil, mockDB, false, noop.ProvideNoopLogger())
		assert.Nil(t, x)
		assert.Error(t, actual)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
