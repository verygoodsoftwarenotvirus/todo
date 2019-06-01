package dbclient

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/noop"
	"testing"
)

func TestMigrate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).
			Return(nil)

		c := &Client{
			database: mockDB,
		}

		actual := c.Migrate(context.Background())

		assert.NoError(t, actual)
	})

	T.Run("bubbles up errors", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).
			Return(errors.New("blah"))

		c := &Client{
			database: mockDB,
		}

		actual := c.Migrate(context.Background())

		assert.Error(t, actual)
	})
}

func TestIsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("IsReady", mock.Anything).
			Return(true)

		c := &Client{
			database: mockDB,
		}

		c.IsReady(context.Background())

		mockDB.AssertExpectations(t)
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).
			Return(nil)

		actual, err := ProvideDatabaseClient(nil, mockDB, false, noop.ProvideNoopLogger())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with error migrating database", func(t *testing.T) {
		expected := errors.New("blah")
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything).
			Return(expected)

		x, actual := ProvideDatabaseClient(nil, mockDB, false, noop.ProvideNoopLogger())
		assert.Nil(t, x)
		assert.Error(t, actual)
		assert.Equal(t, expected, actual)
	})
}
