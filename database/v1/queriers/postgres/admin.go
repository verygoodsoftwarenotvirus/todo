package postgres

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// LogCycleCookieSecretEvent saves a CycleCookieSecretEvent in the audit log table.
func (p *Postgres) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	entry := &models.AuditLogEntryCreationInput{
		EventType: models.CycleCookieSecretEvent,
		Context: map[string]interface{}{
			auditLogActionAssignmentKey: userID,
		},
	}

	p.createAuditLogEntry(ctx, entry)
}
