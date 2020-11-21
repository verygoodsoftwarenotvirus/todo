package postgres

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (p *Postgres) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient))
}
