package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
)

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64) {
	c.querier.LogUserBanEvent(ctx, banGiver, banRecipient)
}

// BanUserAccount marks a user's account as banned.
func (c *Client) BanUserAccount(ctx context.Context, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("BanUserAccount called")

	return c.querier.BanUserAccount(ctx, userID)
}

// TerminateUserAccount marks a user's account as terminated.
func (c *Client) TerminateUserAccount(ctx context.Context, userID uint64) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue("user_id", userID).Debug("BanUserAccount called")

	return c.querier.BanUserAccount(ctx, userID)
}
