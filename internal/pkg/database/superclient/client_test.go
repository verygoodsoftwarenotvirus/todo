package superclient

import (
	"context"
	"database/sql/driver"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

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

var (
	dummyQuery          = "SELECT example FROM testing WHERE things = ? AND stuff = ?"
	formattedDummyQuery = formatQueryForSQLMock(dummyQuery)
	dummyArgs           = []interface{}{"things", "stuff"}
)

func formatQueryForSQLMock(query string) string {
	return strings.NewReplacer(
		"$", `\$`,
		"(", `\(`,
		")", `\)`,
		"=", `\=`,
		"*", `\*`,
		".", `\.`,
		"+", `\+`,
		"?", `\?`,
		",", `\,`,
		"-", `\-`,
		"[", `\[`,
		"]", `\]`,
	).Replace(query)
}

func interfaceToDriverValue(in []interface{}) []driver.Value {
	out := []driver.Value{}

	for _, x := range in {
		out = append(out, driver.Value(x))
	}

	return out
}

// end helper funcs

func buildTestClient(t *testing.T) (*Client, sqlmock.Sqlmock, *database.MockDatabase) {
	t.Helper()

	db, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	mdb := database.BuildMockDatabase()

	c := &Client{
		logger:          noop.NewLogger(),
		querier:         mdb,
		db:              db,
		sqlQueryBuilder: database.BuildMockSQLQueryBuilder(),
		tracer:          tracing.NewTracer("test"),
	}

	return c, sqlMock, mdb
}

func TestMigrate(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		mockDB := database.BuildMockDatabase()
		mockDB.On("Migrate", mock.Anything, (*types.TestUserCreationConfig)(nil)).Return(nil)

		c, _, _ := buildTestClient(t)
		c.querier = mockDB

		assert.NoError(t, c.Migrate(ctx, nil))

		mock.AssertExpectationsForObjects(t, mockDB)
	})

	T.Run("bubbles up errors", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		c, _, _ := buildTestClient(t)

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

		c, _, _ := buildTestClient(t)

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
