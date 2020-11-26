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
func (q *MariaDB) buildSetUserStatusQuery(userID uint64, input types.AccountStatusUpdateInput) (query string, args []interface{}) {
	var err error

	query, args, err = q.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableAccountStatusColumn, input.NewStatus).
		Set(queriers.UsersTableStatusExplanationColumn, input.Reason).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	q.logQueryBuildingError(err)

	return query, args
}

// updateUserAccountStatus updates a user's account status.
func (q *MariaDB) updateUserAccountStatus(ctx context.Context, query string, args []interface{}) error {
	res, err := q.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if count, err := res.RowsAffected(); count == 0 || err != nil {
		return sql.ErrNoRows
	}

	return nil
}

// UpdateUserAccountStatus updates a user's account status.
func (q *MariaDB) UpdateUserAccountStatus(ctx context.Context, userID uint64, input types.AccountStatusUpdateInput) error {
	query, args := q.buildSetUserStatusQuery(userID, input)

	return q.updateUserAccountStatus(ctx, query, args)
}

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (q *MariaDB) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64, reason string) {
	q.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient, reason))
}

// LogAccountTerminationEvent saves a UserBannedEvent in the audit log table.
func (q *MariaDB) LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string) {
	q.createAuditLogEntry(ctx, audit.BuildAccountTerminationEventEntry(terminator, terminee, reason))
}
