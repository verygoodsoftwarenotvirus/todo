package querier

import (
	"context"
	"errors"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClient_GetAuditLogEntriesForAccountUserMembership(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()
		exampleAuditLogEntriesList := fakes.BuildFakeAuditLogEntryList()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAccountUserMembershipQuery", exampleAccountUserMembership.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildMockRowsFromAuditLogEntries(
				false,
				exampleAuditLogEntriesList.Entries...,
			))

		actual, err := c.GetAuditLogEntriesForAccountUserMembership(ctx, exampleAccountUserMembership.ID)
		assert.NoError(t, err)
		assert.Equal(t, exampleAuditLogEntriesList.Entries, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with error executing query", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAccountUserMembershipQuery", exampleAccountUserMembership.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnError(errors.New("blah"))

		actual, err := c.GetAuditLogEntriesForAccountUserMembership(ctx, exampleAccountUserMembership.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})

	T.Run("with erroneous response from database", func(t *testing.T) {
		t.Parallel()

		exampleAccountUserMembership := fakes.BuildFakeAccountUserMembership()

		ctx := context.Background()
		mockQueryBuilder := database.BuildMockSQLQueryBuilder()
		c, db := buildTestClient(t)

		fakeQuery, fakeArgs := fakes.BuildFakeSQLQuery()
		mockQueryBuilder.AccountUserMembershipSQLQueryBuilder.
			On("BuildGetAuditLogEntriesForAccountUserMembershipQuery", exampleAccountUserMembership.ID).
			Return(fakeQuery, fakeArgs)
		c.sqlQueryBuilder = mockQueryBuilder

		db.ExpectQuery(formatQueryForSQLMock(fakeQuery)).
			WithArgs(interfaceToDriverValue(fakeArgs)...).
			WillReturnRows(buildErroneousMockRow())

		actual, err := c.GetAuditLogEntriesForAccountUserMembership(ctx, exampleAccountUserMembership.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, db, mockQueryBuilder)
	})
}
