package mock

import (
	"context"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/mock"
)

var _ types.AuditLogEntryDataManager = (*AuditLogEntryDataManager)(nil)

// AuditLogEntryDataManager is a mocked types.AuditLogEntryDataManager for testing.
type AuditLogEntryDataManager struct {
	mock.Mock
}

// LogUserBanEvent implements our interface.
func (m *AuditLogEntryDataManager) LogUserBanEvent(ctx context.Context, banGiver, banReceiver uint64, reason string) {
	m.Called(ctx, banGiver, banReceiver, reason)
}

// LogAccountTerminationEvent implements our interface.
func (m *AuditLogEntryDataManager) LogAccountTerminationEvent(ctx context.Context, adminID, accountID uint64, reason string) {
	m.Called(ctx, adminID, accountID, reason)
}

// GetAuditLogEntry is a mock function.
func (m *AuditLogEntryDataManager) GetAuditLogEntry(ctx context.Context, entryID uint64) (*types.AuditLogEntry, error) {
	args := m.Called(ctx, entryID)
	return args.Get(0).(*types.AuditLogEntry), args.Error(1)
}

// GetAllAuditLogEntriesCount is a mock function.
func (m *AuditLogEntryDataManager) GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllAuditLogEntries is a mock function.
func (m *AuditLogEntryDataManager) GetAllAuditLogEntries(ctx context.Context, results chan []*types.AuditLogEntry, bucketSize uint16) error {
	args := m.Called(ctx, results, bucketSize)
	return args.Error(0)
}

// GetAuditLogEntries is a mock function.
func (m *AuditLogEntryDataManager) GetAuditLogEntries(ctx context.Context, filter *types.QueryFilter) (*types.AuditLogEntryList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*types.AuditLogEntryList), args.Error(1)
}
