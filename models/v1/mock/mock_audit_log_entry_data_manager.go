package mock

import (
	"context"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/mock"
)

var _ models.AuditLogEntryDataManager = (*AuditLogEntryDataManager)(nil)

// AuditLogEntryDataManager is a mocked models.AuditLogEntryDataManager for testing.
type AuditLogEntryDataManager struct {
	mock.Mock
}

// GetAuditLogEntry is a mock function.
func (m *AuditLogEntryDataManager) GetAuditLogEntry(ctx context.Context, entryID uint64) (*models.AuditLogEntry, error) {
	args := m.Called(ctx, entryID)
	return args.Get(0).(*models.AuditLogEntry), args.Error(1)
}

// GetAllAuditLogEntriesCount is a mock function.
func (m *AuditLogEntryDataManager) GetAllAuditLogEntriesCount(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

// GetAllAuditLogEntries is a mock function.
func (m *AuditLogEntryDataManager) GetAllAuditLogEntries(ctx context.Context, results chan []models.AuditLogEntry) error {
	args := m.Called(ctx, results)
	return args.Error(0)
}

// GetAuditLogEntries is a mock function.
func (m *AuditLogEntryDataManager) GetAuditLogEntries(ctx context.Context, filter *models.QueryFilter) (*models.AuditLogEntryList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*models.AuditLogEntryList), args.Error(1)
}

// CreateAuditLogEntry is a mock function.
func (m *AuditLogEntryDataManager) CreateAuditLogEntry(ctx context.Context, input *models.AuditLogEntryCreationInput) error {
	return m.Called(ctx, input).Error(0)
}
