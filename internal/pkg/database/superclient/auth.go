package superclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
)

// LogCycleCookieSecretEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogCycleCookieSecretEvent called")

	c.querier.LogCycleCookieSecretEvent(ctx, userID)
}

// LogSuccessfulLoginEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogSuccessfulLoginEvent called")

	c.querier.LogSuccessfulLoginEvent(ctx, userID)
}

// LogBannedUserLoginAttemptEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogBannedUserLoginAttemptEvent called")

	c.querier.LogBannedUserLoginAttemptEvent(ctx, userID)
}

// LogUnsuccessfulLoginBadPasswordEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUnsuccessfulLoginBadPasswordEvent called")

	c.querier.LogUnsuccessfulLoginBadPasswordEvent(ctx, userID)
}

// LogUnsuccessfulLoginBad2FATokenEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogUnsuccessfulLoginBad2FATokenEvent called")

	c.querier.LogUnsuccessfulLoginBad2FATokenEvent(ctx, userID)
}

// LogLogoutEvent implements our AuditLogEntryDataManager interface.
func (c *Client) LogLogoutEvent(ctx context.Context, userID uint64) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	c.logger.WithValue(keys.UserIDKey, userID).Debug("LogLogoutEvent called")

	c.querier.LogLogoutEvent(ctx, userID)
}
