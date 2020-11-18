package mariadb

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (m *MariaDB) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64) {
	m.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient))
}
