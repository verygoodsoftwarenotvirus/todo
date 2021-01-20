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
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
)

const (
	defaultLimit = uint8(20)
)

// begin helper funcs

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

type expecterSqlmockWrapper struct {
	sqlmock.Sqlmock
}

func (e *expecterSqlmockWrapper) AssertExpectations(t mock.TestingT) bool {
	return assert.NoError(t, e.Sqlmock.ExpectationsWereMet(), "not all database expectations were met")
}

func buildTestClient(t *testing.T) (*Client, *expecterSqlmockWrapper) {
	t.Helper()

	db, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	c := &Client{
		db:              db,
		logger:          noop.NewLogger(),
		timeTeller:      &queriers.StandardTimeTeller{},
		tracer:          tracing.NewTracer("test"),
		sqlQueryBuilder: database.BuildMockSQLQueryBuilder(),
	}

	return c, &expecterSqlmockWrapper{Sqlmock: sqlMock}
}

// end helper funcs

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
