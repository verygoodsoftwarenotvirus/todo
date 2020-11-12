package dbclient

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/app/database"
	mockauth "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/auth/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

const (
	defaultLimit = uint8(20)
)

func buildTestClient() (*Client, *database.MockDatabase) {
	db := database.BuildMockDatabase()
	c := &Client{
		logger:  noop.NewLogger(),
		querier: db,
	}
	return c, db
}

func TestMigrate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		authenticator := &mockauth.Authenticator{}

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, authenticator, (*database.UserCreationConfig)(nil)).Return(nil)

		c := &Client{querier: mockDB}
		assert.NoError(t, c.Migrate(ctx, authenticator, nil))

		mock.AssertExpectationsForObjects(t, authenticator, mockDB)
	})

	T.Run("bubbles up errors", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		authenticator := &mockauth.Authenticator{}

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, authenticator, (*database.UserCreationConfig)(nil)).Return(errors.New("blah"))

		c := &Client{querier: mockDB}
		assert.Error(t, c.Migrate(ctx, authenticator, nil))

		mock.AssertExpectationsForObjects(t, authenticator, mockDB)
	})
}

func TestIsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		ctx := context.Background()
		authenticator := &mockauth.Authenticator{}

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, authenticator, (*database.UserCreationConfig)(nil)).Return(nil)

		actual, err := ProvideDatabaseClient(ctx, noop.NewLogger(), mockDB, nil, authenticator, nil, true, true)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, authenticator, mockDB)
	})

	T.Run("with error migrating querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		authenticator := &mockauth.Authenticator{}

		expected := errors.New("blah")
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, authenticator, (*database.UserCreationConfig)(nil)).Return(expected)

		x, actual := ProvideDatabaseClient(ctx, noop.NewLogger(), mockDB, nil, authenticator, nil, true, true)
		assert.Nil(t, x)
		assert.Error(t, actual)
		assert.Equal(t, expected, errors.Unwrap(actual))

		mock.AssertExpectationsForObjects(t, authenticator, mockDB)
	})
}
