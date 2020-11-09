package sqlite

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (s *Sqlite) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.CycleCookieSecretEvent,
		Context: map[string]interface{}{
			auditLogUserAssignmentKey: userID,
		},
	}

	s.createAuditLogEntry(ctx, entry)
}
