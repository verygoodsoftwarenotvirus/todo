package mock

import (
	"context"
)

// LogCycleCookieSecretEvent implements our interface.
func (m *AuditLogEntryDataManager) LogCycleCookieSecretEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogSuccessfulLoginEvent implements our interface.
func (m *AuditLogEntryDataManager) LogSuccessfulLoginEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogBannedUserLoginAttemptEvent implements our interface.
func (m *AuditLogEntryDataManager) LogBannedUserLoginAttemptEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUnsuccessfulLoginBadPasswordEvent implements our interface.
func (m *AuditLogEntryDataManager) LogUnsuccessfulLoginBadPasswordEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogUnsuccessfulLoginBad2FATokenEvent implements our interface.
func (m *AuditLogEntryDataManager) LogUnsuccessfulLoginBad2FATokenEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}

// LogLogoutEvent implements our interface.
func (m *AuditLogEntryDataManager) LogLogoutEvent(ctx context.Context, userID uint64) {
	m.Called(ctx, userID)
}
