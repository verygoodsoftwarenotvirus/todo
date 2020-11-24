package mariadb

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// buildSetUserStatusQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (m *MariaDB) buildSetUserStatusQuery(userID uint64, input types.AccountStatusUpdateInput) (query string, args []interface{}) {
	var err error

	query, args, err = m.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableAccountStatusColumn, input.NewStatus).
		Set(queriers.UsersTableStatusExplanationColumn, input.Reason).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	m.logQueryBuildingError(err)

	return query, args
}

// updateUserAccountStatus updates a user's account status.
func (m *MariaDB) updateUserAccountStatus(ctx context.Context, query string, args []interface{}) error {
	res, err := m.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, err := res.RowsAffected(); count == 0 || err != nil {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateUserAccountStatus updates a user's account status.
func (m *MariaDB) UpdateUserAccountStatus(ctx context.Context, userID uint64, input types.AccountStatusUpdateInput) error {
	query, args := m.buildSetUserStatusQuery(userID, input)

	return m.updateUserAccountStatus(ctx, query, args)
}

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (m *MariaDB) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64, reason string) {
	m.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient, reason))
}

// LogAccountTerminationEvent saves a UserBannedEvent in the audit log table.
func (m *MariaDB) LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string) {
	m.createAuditLogEntry(ctx, audit.BuildAccountTerminationEventEntry(terminator, terminee, reason))
}
