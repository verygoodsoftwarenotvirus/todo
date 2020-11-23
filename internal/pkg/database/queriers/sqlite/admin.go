package sqlite

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (s *Sqlite) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64) {
	s.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient))
}

// buildSetUserStatusQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (s *Sqlite) buildSetUserStatusQuery(userID uint64, status string) (query string, args []interface{}) {
	var err error

	query, args, err = s.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableAccountStatusColumn, status).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	s.logQueryBuildingError(err)

	return query, args
}

// updateUserAccountStatus updates a user's account status.
func (s *Sqlite) updateUserAccountStatus(ctx context.Context, query string, args []interface{}) error {
	res, err := s.db.ExecContext(ctx, query, args...)

	if err != nil {
		return err
	}

	if count, err := res.RowsAffected(); count == 0 || err != nil {
		return sql.ErrNoRows
	}

	return nil
}

// BanUserAccount bans a user's account.
func (s *Sqlite) BanUserAccount(ctx context.Context, userID uint64) error {
	query, args := s.buildSetUserStatusQuery(userID, string(types.BannedAccountStatus))

	return s.updateUserAccountStatus(ctx, query, args)
}

// TerminateUserAccount terminates a user's account.
func (s *Sqlite) TerminateUserAccount(ctx context.Context, userID uint64) error {
	query, args := s.buildSetUserStatusQuery(userID, string(types.TerminatedAccountStatus))

	return s.updateUserAccountStatus(ctx, query, args)
}
