package dbclient

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

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
		tracer:  tracing.NewTracer("test"),
	}

	return c, db
}

func TestMigrate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, (*types.TestUserCreationConfig)(nil)).Return(nil)

		c, _ := buildTestClient()
		c.querier = mockDB

		assert.NoError(t, c.Migrate(ctx, nil))

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("bubbles up errors", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _ := buildTestClient()

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, (*types.TestUserCreationConfig)(nil)).Return(errors.New("blah"))

		c.querier = mockDB

		assert.Error(t, c.Migrate(ctx, nil))

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestIsReady(T *testing.T) {
	T.Parallel()

	T.Run("obligatory", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _ := buildTestClient()

		mockDB := database.BuildMockDatabase()
		mockDB.On("IsReady", mock.Anything).Return(true)

		c.querier = mockDB
		c.IsReady(ctx)

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}

func TestProvideDatabaseClient(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, (*types.TestUserCreationConfig)(nil)).Return(nil)

		actual, err := ProvideDatabaseClient(ctx, noop.NewLogger(), mockDB, nil, nil, true, true)
		assert.NotNil(t, actual)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("with error migrating querier", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		expected := errors.New("blah")
		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, (*types.TestUserCreationConfig)(nil)).Return(expected)

		x, actual := ProvideDatabaseClient(ctx, noop.NewLogger(), mockDB, nil, nil, true, true)
		assert.Nil(t, x)
		assert.Error(t, actual)
		assert.Equal(t, expected, errors.Unwrap(actual))

		mock.AssertExpectationsForObjects(t, mockDB)
	})
}
