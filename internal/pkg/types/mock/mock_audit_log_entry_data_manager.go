package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AuditLogDataManager = (*AuditLogDataManager)(nil)

// AuditLogDataManager is a mocked types.AuditLogDataManager for testing.
type AuditLogDataManager struct {
	mock.Mock
}

// LogUserBanEvent implements our interface.
func (m *AuditLogDataManager) LogUserBanEvent(ctx context.Context, banGiver, banReceiver uint64) {
	m.Called(ctx, banGiver, banReceiver)
}

// GetAuditLogEntry is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	args := m.Called(ctx, entryID)
	return args.Get(0).(*types.AuditLogEntry), args.Error(1)
}

// GetAllAuditLogEntriesCount is a mock function.
func (m *AuditLogDataManager) GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllAuditLogEntries is a mock function.
func (m *AuditLogDataManager) GetAllAuditLogEntries(ctx context.Context, results chan []types.AuditLogEntry) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

// GetAuditLogEntries is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*types.AuditLogEntryList), args.Error(1)
}

// LogCycleCookieSecretEvent implements our interface.
func (m *AuditLogDataManager) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogSuccessfulLoginEvent implements our interface.
func (m *AuditLogDataManager) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogBannedUserLoginAttemptEvent implements our interface.
func (m *AuditLogDataManager) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUnsuccessfulLoginBadPasswordEvent implements our interface.
func (m *AuditLogDataManager) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUnsuccessfulLoginBad2FATokenEvent implements our interface.
func (m *AuditLogDataManager) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogLogoutEvent implements our interface.
func (m *AuditLogDataManager) LogLogoutEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogItemCreationEvent implements our interface.
func (m *AuditLogDataManager) LogItemCreationEvent(ctx context.Context, item *types.Item) {
	m.Called(ctx, item)
}

// LogItemUpdateEvent implements our interface.
func (m *AuditLogDataManager) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []types.FieldChangeSummary) {
	m.Called(ctx, userID, itemID, changes)
}

// LogItemArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	m.Called(ctx, userID, itemID)
}

// GetAuditLogEntriesForItem is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntriesForItem(ctx context.Context, itemID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, itemID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}

// LogOAuth2ClientCreationEvent implements our interface.
func (m *AuditLogDataManager) LogOAuth2ClientCreationEvent(ctx context.Context, client *types.OAuth2Client) {
	m.Called(ctx, client)
}

// LogOAuth2ClientArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	m.Called(ctx, userID, clientID)
}

// GetAuditLogEntriesForOAuth2Client is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntriesForOAuth2Client(ctx context.Context, clientID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, clientID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}

// LogWebhookCreationEvent implements our interface.
func (m *AuditLogDataManager) LogWebhookCreationEvent(ctx context.Context, webhook *types.Webhook) {
	m.Called(ctx, webhook)
}

// LogWebhookUpdateEvent implements our interface.
func (m *AuditLogDataManager) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []types.FieldChangeSummary) {
	m.Called(ctx, userID, webhookID, changes)
}

// LogWebhookArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	m.Called(ctx, userID, webhookID)
}

// GetAuditLogEntriesForWebhook is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntriesForWebhook(ctx context.Context, webhookID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, webhookID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}

// LogUserCreationEvent implements our interface.
func (m *AuditLogDataManager) LogUserCreationEvent(ctx context.Context, user *types.User) {
	m.Called(ctx, user)
}

// LogUserVerifyTwoFactorSecretEvent implements our interface.
func (m *AuditLogDataManager) LogUserVerifyTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUserUpdateTwoFactorSecretEvent implements our interface.
func (m *AuditLogDataManager) LogUserUpdateTwoFactorSecretEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUserUpdatePasswordEvent implements our interface.
func (m *AuditLogDataManager) LogUserUpdatePasswordEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUserArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogUserArchiveEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// GetAuditLogEntriesForUser is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntriesForUser(ctx context.Context, userID uint64) ([]types.AuditLogEntry, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]types.AuditLogEntry), args.Error(1)
}
