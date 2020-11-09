package sqlite

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/stretchr/testify/assert"
)

func TestSqlite_LogCycleCookieSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		s, mockDB := buildTestService(t)

		exampleInput := fakemodels.BuildFakeUser()
		exampleAuditLogEntry := &models.AuditLogEntry{
			EventType: models.CycleCookieSecretEvent,
			Context: map[string]interface{}{
				auditLogUserAssignmentKey: exampleInput.ID,
			},
		}

		expectedQuery, expectedArgs := s.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		mockDB.ExpectExec(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...)

		s.LogCycleCookieSecretEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
