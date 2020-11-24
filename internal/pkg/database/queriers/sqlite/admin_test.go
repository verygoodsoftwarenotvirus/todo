package sqlite

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSqlite_UpdateUserAccountStatus(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		exampleUser := fakes.BuildFakeUser()
		exampleInput := *fakes.BuildFakeAccountStatusUpdateInput()

		s, mockDB := buildTestService(t)
		expectedQuery, expectedArgs := s.buildSetUserStatusQuery(exampleUser.ID, exampleInput)

		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := s.UpdateUserAccountStatus(ctx, exampleUser.ID, exampleInput)
		assert.NoError(t, err)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
