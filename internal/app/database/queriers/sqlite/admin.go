package sqlite

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (s *Sqlite) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient))
}
