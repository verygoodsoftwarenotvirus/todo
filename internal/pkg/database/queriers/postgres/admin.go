package postgres

import (
	"context"
	"database/sql"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/database/queriers"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/Masterminds/squirrel"
)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (p *Postgres) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64) {
	p.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient))
}

// buildSetUserStatusQuery returns a SQL query (and arguments) that would set a user's account status to banned.
func (p *Postgres) buildSetUserStatusQuery(userID uint64, status string) (query string, args []interface{}) {
	var err error

	query, args, err = p.sqlBuilder.
		Update(queriers.UsersTableName).
		Set(queriers.UsersTableAccountStatusColumn, status).
		Where(squirrel.Eq{queriers.IDColumn: userID}).
		ToSql()

	p.logQueryBuildingError(err)

	return query, args
}

// updateUserAccountStatus updates a user's account status.
func (p *Postgres) updateUserAccountStatus(ctx context.Context, query string, args []interface{}) error {
	res, err := p.db.ExecContext(ctx, query, args...)

	if err != nil {
		return err
	}

	if count, err := res.RowsAffected(); count == 0 || err != nil {
		return sql.ErrNoRows
	}

	return nil
}

// BanUserAccount bans a user's account.
func (p *Postgres) BanUserAccount(ctx context.Context, userID uint64) error {
	query, args := p.buildSetUserStatusQuery(userID, string(types.BannedAccountStatus))

	return p.updateUserAccountStatus(ctx, query, args)
}

// TerminateUserAccount terminates a user's account.
func (p *Postgres) TerminateUserAccount(ctx context.Context, userID uint64) error {
	query, args := p.buildSetUserStatusQuery(userID, string(types.TerminatedAccountStatus))

	return p.updateUserAccountStatus(ctx, query, args)
}
