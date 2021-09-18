package mysql

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

var _ types.AdminUserDataManager = (*SQLQuerier)(nil)

const setUserReputationQuery = `
	UPDATE users SET reputation = ?, reputation_explanation = ?, last_updated_on = UNIX_TIMESTAMP() WHERE archived_on IS NULL AND id = ?
`

// UpdateUserReputation updates a user's account status.
func (q *SQLQuerier) UpdateUserReputation(ctx context.Context, userID string, input *types.UserReputationUpdateInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachUserIDToSpan(span, userID)

	args := []interface{}{
		input.NewReputation,
		input.Reason,
		input.TargetUserID,
	}

	if err := q.performWriteQuery(ctx, q.db, "user status update query", setUserReputationQuery, args); err != nil {
		return observability.PrepareError(err, logger, span, "user status update")
	}

	logger.Info("user reputation updated")

	return nil
}
