package dbclient

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

var _ types.AuditLogDataManager = (*Client)(nil)

// GetAuditLogEntry fetches an audit log entry from the database.
func (c *Client) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntry")
	defer span.End()

	tracing.AttachAuditLogEntryIDToSpan(span, entryID)
	c.logger.WithValue("audit_log_entry_id", entryID).Debug("GetAuditLogEntry called")

	return c.querier.GetAuditLogEntry(ctx, entryID)
}

// GetAllAuditLogEntriesCount fetches the count of audit log entries from the database that meet a particular filter.
func (c *Client) GetAllAuditLogEntriesCount(ctx context.Context) (count uint64, err error) {
	ctx, span := tracing.StartSpan(ctx, "GetAllAuditLogEntriesCount")
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntriesCount called")

	return c.querier.GetAllAuditLogEntriesCount(ctx)
}

// GetAllAuditLogEntries fetches a list of all audit log entries in the database.
func (c *Client) GetAllAuditLogEntries(ctx context.Context, results chan []types.AuditLogEntry) error {
	ctx, span := tracing.StartSpan(ctx, "GetAllAuditLogEntries")
	defer span.End()

	c.logger.Debug("GetAllAuditLogEntries called")

	return c.querier.GetAllAuditLogEntries(ctx, results)
}

// GetAuditLogEntries fetches a list of audit log entries from the database that meet a particular filter.
func (c *Client) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	ctx, span := tracing.StartSpan(ctx, "GetAuditLogEntries")
	defer span.End()

	tracing.AttachFilterToSpan(span, filter)
	c.logger.Debug("GetAuditLogEntries called")

	return c.querier.GetAuditLogEntries(ctx, filter)
}

// LogCycleCookieSecretEvent implements our AuditLogDataManager interface.
func (c *Client) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogCycleCookieSecretEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogCycleCookieSecretEvent called")

	c.querier.LogCycleCookieSecretEvent(ctx, userID)
}

// LogSuccessfulLoginEvent implements our AuditLogDataManager interface.
func (c *Client) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogSuccessfulLoginEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogSuccessfulLoginEvent called")

	c.querier.LogSuccessfulLoginEvent(ctx, userID)
}

// LogUnsuccessfulLoginBadPasswordEvent implements our AuditLogDataManager interface.
func (c *Client) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogUnsuccessfulLoginBadPasswordEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogUnsuccessfulLoginBadPasswordEvent called")

	c.querier.LogUnsuccessfulLoginBadPasswordEvent(ctx, userID)
}

// LogUnsuccessfulLoginBad2FATokenEvent implements our AuditLogDataManager interface.
func (c *Client) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogUnsuccessfulLoginBad2FATokenEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogUnsuccessfulLoginBad2FATokenEvent called")

	c.querier.LogUnsuccessfulLoginBad2FATokenEvent(ctx, userID)
}

// LogLogoutEvent implements our AuditLogDataManager interface.
func (c *Client) LogLogoutEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogLogoutEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogLogoutEvent called")

	c.querier.LogLogoutEvent(ctx, userID)
}

// LogItemCreationEvent implements our AuditLogDataManager interface.
func (c *Client) LogItemCreationEvent(ctx context.Context, item *types.Item) {
	ctx, span := tracing.StartSpan(ctx, "LogItemCreationEvent")
	defer span.End()

	c.logger.WithValue("user_id", item.BelongsToUser).Debug("LogItemCreationEvent called")

	c.querier.LogItemCreationEvent(ctx, item)
}

// LogItemUpdateEvent implements our AuditLogDataManager interface.
func (c *Client) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []types.FieldChangeSummary) {
	ctx, span := tracing.StartSpan(ctx, "LogItemUpdateEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogItemUpdateEvent called")

	c.querier.LogItemUpdateEvent(ctx, userID, itemID, changes)
}

// LogItemArchiveEvent implements our AuditLogDataManager interface.
func (c *Client) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogItemArchiveEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogItemArchiveEvent called")

	c.querier.LogItemArchiveEvent(ctx, userID, itemID)
}

// LogOAuth2ClientCreationEvent implements our AuditLogDataManager interface.
func (c *Client) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	ctx, span := tracing.StartSpan(ctx, "LogOAuth2ClientCreationEvent")
	defer span.End()

	c.logger.WithValue("user_id", client.BelongsToUser).Debug("LogOAuth2ClientCreationEvent called")

	c.querier.LogOAuth2ClientCreationEvent(ctx, client)
}

// LogOAuth2ClientArchiveEvent implements our AuditLogDataManager interface.
func (c *Client) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogOAuth2ClientArchiveEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogOAuth2ClientArchiveEvent called")

	c.querier.LogOAuth2ClientArchiveEvent(ctx, userID, clientID)
}

// LogWebhookCreationEvent implements our AuditLogDataManager interface.
func (c *Client) LogWebhookCreationEvent(ctx context.Context, webhook *types.Webhook) {
	ctx, span := tracing.StartSpan(ctx, "LogWebhookCreationEvent")
	defer span.End()

	c.logger.WithValue("user_id", webhook.BelongsToUser).Debug("LogWebhookCreationEvent called")

	c.querier.LogWebhookCreationEvent(ctx, webhook)
}

// LogWebhookUpdateEvent implements our AuditLogDataManager interface.
func (c *Client) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []types.FieldChangeSummary) {
	ctx, span := tracing.StartSpan(ctx, "LogWebhookUpdateEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogWebhookUpdateEvent called")

	c.querier.LogWebhookUpdateEvent(ctx, userID, webhookID, changes)
}

// LogWebhookArchiveEvent implements our AuditLogDataManager interface.
func (c *Client) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogWebhookArchiveEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogWebhookArchiveEvent called")

	c.querier.LogWebhookArchiveEvent(ctx, userID, webhookID)
}

// LogUserCreationEvent implements our AuditLogDataManager interface.
func (c *Client) LogUserCreationEvent(ctx context.Context, user *types.User) {
	ctx, span := tracing.StartSpan(ctx, "LogUserCreationEvent")
	defer span.End()

	c.logger.WithValue("user_id", user.ID).Debug("LogUserCreationEvent called")

	c.querier.LogUserCreationEvent(ctx, user)
}

// LogUserVerifyTwoFactorSecretEvent implements our AuditLogDataManager interface.
func (c *Client) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogUserVerifyTwoFactorSecretEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogUserVerifyTwoFactorSecretEvent called")

	c.querier.LogUserVerifyTwoFactorSecretEvent(ctx, userID)
}

// LogUserUpdateTwoFactorSecretEvent implements our AuditLogDataManager interface.
func (c *Client) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogUserUpdateTwoFactorSecretEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogUserUpdateTwoFactorSecretEvent called")

	c.querier.LogUserUpdateTwoFactorSecretEvent(ctx, userID)
}

// LogUserUpdatePasswordEvent implements our AuditLogDataManager interface.
func (c *Client) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogUserUpdatePasswordEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogUserUpdatePasswordEvent called")

	c.querier.LogUserUpdatePasswordEvent(ctx, userID)
}

// LogUserArchiveEvent implements our AuditLogDataManager interface.
func (c *Client) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	ctx, span := tracing.StartSpan(ctx, "LogUserArchiveEvent")
	defer span.End()

	c.logger.WithValue("user_id", userID).Debug("LogUserArchiveEvent called")

	c.querier.LogUserArchiveEvent(ctx, userID)
}
