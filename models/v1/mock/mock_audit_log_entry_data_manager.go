package mock

import (
	"context"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.AuditLogDataManager = (*AuditLogDataManager)(nil)

// AuditLogDataManager is a mocked models.AuditLogDataManager for testing.
type AuditLogDataManager struct {
	mock.Mock
}

// GetAuditLogEntry is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntry(ctx context.Context, entryID uint64) (*models.AuditLogEntry, error) {
	args := m.Called(ctx, entryID)
	return args.Get(0).(*models.AuditLogEntry), args.Error(1)
}

// GetAllAuditLogEntriesCount is a mock function.
func (m *AuditLogDataManager) GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllAuditLogEntries is a mock function.
func (m *AuditLogDataManager) GetAllAuditLogEntries(ctx context.Context, results chan []models.AuditLogEntry) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

// GetAuditLogEntries is a mock function.
func (m *AuditLogDataManager) GetAuditLogEntries(ctx context.Context, filter *models.QueryFilter) (*models.AuditLogEntryList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.AuditLogEntryList), args.Error(1)
}

// CreateAuditLogEntry is a mock function.
func (m *AuditLogDataManager) CreateAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) {
	m.Called(ctx, input)
}

// LogCycleCookieSecretEvent implements our interface.
func (m *AuditLogDataManager) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogSuccessfulLoginEvent implements our interface.
func (m *AuditLogDataManager) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
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
func (m *AuditLogDataManager) LogItemCreationEvent(ctx context.Context, item *models.Item) {
	m.Called(ctx, item)
}

// LogItemUpdateEvent implements our interface.
func (m *AuditLogDataManager) LogItemUpdateEvent(ctx context.Context, userID, itemID uint64, changes []models.FieldChangeEvent) {
	m.Called(ctx, userID, itemID, changes)
}

// LogItemArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogItemArchiveEvent(ctx context.Context, userID, itemID uint64) {
	m.Called(ctx, userID, itemID)
}

// LogOAuth2ClientCreationEvent implements our interface.
func (m *AuditLogDataManager) LogOAuth2ClientCreationEvent(ctx context.Context, client *models.OAuth2Client) {
	m.Called(ctx, client)
}

// LogOAuth2ClientArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogOAuth2ClientArchiveEvent(ctx context.Context, userID, clientID uint64) {
	m.Called(ctx, userID, clientID)
}

// LogWebhookCreationEvent implements our interface.
func (m *AuditLogDataManager) LogWebhookCreationEvent(ctx context.Context, webhook *models.Webhook) {
	m.Called(ctx, webhook)
}

// LogWebhookUpdateEvent implements our interface.
func (m *AuditLogDataManager) LogWebhookUpdateEvent(ctx context.Context, userID, webhookID uint64, changes []models.FieldChangeEvent) {
	m.Called(ctx, userID, webhookID, changes)
}

// LogWebhookArchiveEvent implements our interface.
func (m *AuditLogDataManager) LogWebhookArchiveEvent(ctx context.Context, userID, webhookID uint64) {
	m.Called(ctx, userID, webhookID)
}

// LogUserCreationEvent implements our interface.
func (m *AuditLogDataManager) LogUserCreationEvent(ctx context.Context, user *models.User) {
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
