package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AdminUserDataManager = (*Client)(nil)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64, reason string) {
	c.querier.LogUserBanEvent(ctx, banGiver, banRecipient, reason)
}

// LogAccountTerminationEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string) {
	c.querier.LogAccountTerminationEvent(ctx, terminator, terminee, reason)
}

// UpdateUserAccountStatus marks a user's account as banned.
func (c *Client) UpdateUserAccountStatus(ctx context.Context, userID uint64, input types.AccountStatusUpdateInput) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("UpdateUserAccountStatus called")

	return c.querier.UpdateUserAccountStatus(ctx, userID, input)
}
