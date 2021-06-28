package fakes

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/audit"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeAuditLogEntry builds a faked audit log entry.
func BuildFakeAuditLogEntry() *types.AuditLogEntry {
	return &types.AuditLogEntry{
		ID:         uint64(fake.Uint32()),
		ExternalID: fake.UUID(),
		EventType:  audit.SuccessfulLoginEvent,
		Context:    map[string]interface{}{"fakes": "true"},
		CreatedOn:  uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakeAuditLogEntryList builds a faked AuditLogEntryList.
func BuildFakeAuditLogEntryList() *types.AuditLogEntryList {
	var examples []*types.AuditLogEntry
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeAuditLogEntry())
	}

	return &types.AuditLogEntryList{
		Pagination: types.Pagination{
			Page:  1,
			Limit: 20,
			//FilteredCount: exampleQuantity / 2,
			TotalCount: exampleQuantity,
		},
		Entries: examples,
	}
}

// BuildFakeAuditLogEntryCreationInput builds a faked AuditLogEntryCreationInput.
func BuildFakeAuditLogEntryCreationInput() *types.AuditLogEntryCreationInput {
	entry := BuildFakeAuditLogEntry()
	return BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(entry)
}

// BuildFakeAuditLogEntryCreationInputFromAuditLogEntry builds a faked AuditLogEntryCreationInput from an audit log entry.
func BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(entry *types.AuditLogEntry) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: entry.EventType,
		Context:   entry.Context,
	}
}
