package superclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AdminUserDataManager = (*Client)(nil)

// UpdateUserAccountStatus updates a user's account status.
func (c *Client) UpdateUserAccountStatus(ctx context.Context, userID uint64, input types.AccountStatusUpdateInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("UpdateUserAccountStatus called")

	query, args := c.sqlQueryBuilder.BuildSetUserStatusQuery(userID, input)

	return c.execContext(ctx, "user status update query", query, args)
}

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64, reason string) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, banRecipient)
	c.logger.WithValue("ban_recipient", banRecipient).WithValue("ban_giver", banGiver).Debug("LogUserBanEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUserBanEventEntry(banGiver, banRecipient, reason))
}

// LogAccountTerminationEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, terminee)
	c.logger.WithValue("termination_recipient", terminee).WithValue("terminator", terminator).Debug("LogAccountTerminationEvent called")

	c.createAuditLogEntry(ctx, audit.BuildAccountTerminationEventEntry(terminator, terminee, reason))
}
