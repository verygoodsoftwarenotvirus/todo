package querier

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AdminUserDataManager = (*Client)(nil)

// UpdateUserAccountStatus updates a user's account status.
func (c *Client) UpdateUserAccountStatus(ctx context.Context, userID uint64, input types.UserReputationUpdateInput) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, userID)
	c.logger.WithValue(keys.UserIDKey, userID).Debug("UpdateUserAccountStatus called")

	query, args := c.sqlQueryBuilder.BuildSetUserStatusQuery(userID, input)

	return c.performWriteQueryIgnoringReturn(ctx, c.db, "user status update query", query, args)
}

// LogUserBanEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogUserBanEvent(ctx context.Context, banGiver, banRecipient uint64, reason string) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, banRecipient)

	c.createAuditLogEntry(ctx, c.db, audit.BuildUserBanEventEntry(banGiver, banRecipient, reason))
}

// LogAccountTerminationEvent saves a UserBannedEvent in the audit log table.
func (c *Client) LogAccountTerminationEvent(ctx context.Context, terminator, terminee uint64, reason string) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachUserIDToSpan(span, terminee)

	c.createAuditLogEntry(ctx, c.db, audit.BuildAccountTerminationEventEntry(terminator, terminee, reason))
}

// LogCycleCookieSecretEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.createAuditLogEntry(ctx, c.db, audit.BuildCycleCookieSecretEvent(userID))
}

// LogSuccessfulLoginEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.createAuditLogEntry(ctx, c.db, audit.BuildSuccessfulLoginEventEntry(userID))
}

// LogBannedUserLoginAttemptEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.createAuditLogEntry(ctx, c.db, audit.BuildBannedUserLoginAttemptEventEntry(userID))
}

// LogUnsuccessfulLoginBadPasswordEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.createAuditLogEntry(ctx, c.db, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(userID))
}

// LogUnsuccessfulLoginBad2FATokenEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.createAuditLogEntry(ctx, c.db, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(userID))
}

// LogLogoutEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogLogoutEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.createAuditLogEntry(ctx, c.db, audit.BuildLogoutEventEntry(userID))
}
