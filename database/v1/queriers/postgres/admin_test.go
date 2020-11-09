package postgres

import (
	"context"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgres_LogCycleCookieSecretEvent(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		p, mockDB := buildTestService(t)

		exampleInput := fakemodels.BuildFakeUser()
		exampleAuditLogEntry := &models.AuditLogEntry{
			EventType: models.CycleCookieSecretEvent,
			Context: map[string]interface{}{
				auditLogActionAssignmentKey: exampleInput.ID,
			},
		}

		expectedQuery, expectedArgs := p.buildCreateAuditLogEntryQuery(exampleAuditLogEntry)
		exampleRows := sqlmock.NewRows([]string{"id", "created_on"}).AddRow(exampleInput.ID, exampleInput.CreatedOn)
		mockDB.ExpectQuery(formatQueryForSQLMock(expectedQuery)).
			WithArgs(interfaceToDriverValue(expectedArgs)...).
			WillReturnRows(exampleRows)

		p.LogCycleCookieSecretEvent(ctx, exampleInput.ID)

		assert.NoError(t, mockDB.ExpectationsWereMet(), "not all database expectations were met")
	})
}
