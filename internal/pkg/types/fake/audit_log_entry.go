package fake

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeAuditLogEntry builds a faked item.
func BuildFakeAuditLogEntry() *types.AuditLogEntry {
	return &types.AuditLogEntry{
		ID:        fake.Uint64(),
		EventType: audit.SuccessfulLoginEvent,
		Context:   map[string]interface{}{"fake": "true"},
		CreatedOn: uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakeAuditLogEntryList builds a faked AuditLogEntryList.
func BuildFakeAuditLogEntryList() *types.AuditLogEntryList {
	exampleAuditLogEntry1 := BuildFakeAuditLogEntry()
	exampleAuditLogEntry2 := BuildFakeAuditLogEntry()
	exampleAuditLogEntry3 := BuildFakeAuditLogEntry()

	return &types.AuditLogEntryList{
		Pagination: types.Pagination{
			Page:  1,
			Limit: 20,
		},
		Entries: []types.AuditLogEntry{
			*exampleAuditLogEntry1,
			*exampleAuditLogEntry2,
			*exampleAuditLogEntry3,
		},
	}
}

// BuildFakeAuditLogEntryCreationInput builds a faked AuditLogEntryCreationInput.
func BuildFakeAuditLogEntryCreationInput() *types.AuditLogEntryCreationInput {
	item := BuildFakeAuditLogEntry()
	return BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(item)
}

// BuildFakeAuditLogEntryCreationInputFromAuditLogEntry builds a faked AuditLogEntryCreationInput from an item.
func BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(entry *types.AuditLogEntry) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: entry.EventType,
		Context:   entry.Context,
	}
}
