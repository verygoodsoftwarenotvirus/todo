package mock

import (
	"context"
)

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
