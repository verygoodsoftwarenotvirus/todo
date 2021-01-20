package superclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
)

// LogCycleCookieSecretEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogCycleCookieSecretEvent called")

	c.createAuditLogEntry(ctx, audit.BuildCycleCookieSecretEvent(userID))
}

// LogSuccessfulLoginEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogSuccessfulLoginEvent called")

	c.createAuditLogEntry(ctx, audit.BuildSuccessfulLoginEventEntry(userID))
}

// LogBannedUserLoginAttemptEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogBannedUserLoginAttemptEvent called")

	c.createAuditLogEntry(ctx, audit.BuildBannedUserLoginAttemptEventEntry(userID))
}

// LogUnsuccessfulLoginBadPasswordEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUnsuccessfulLoginBadPasswordEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBadPasswordEventEntry(userID))
}

// LogUnsuccessfulLoginBad2FATokenEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUnsuccessfulLoginBad2FATokenEvent called")

	c.createAuditLogEntry(ctx, audit.BuildUnsuccessfulLoginBad2FATokenEventEntry(userID))
}

// LogLogoutEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogLogoutEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogLogoutEvent called")

	c.createAuditLogEntry(ctx, audit.BuildLogoutEventEntry(userID))
}
