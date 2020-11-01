package fakemodels

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeAuditLogEntry builds a faked item.
func BuildFakeAuditLogEntry() *models.AuditLogEntry {
	return &models.AuditLogEntry{
		ID:        fake.Uint64(),
		EventType: models.SuccessfulLoginEvent,
		Context:   map[string]interface{}{"fake": "true"},
		CreatedOn: uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakeAuditLogEntryList builds a faked AuditLogEntryList.
func BuildFakeAuditLogEntryList() *models.AuditLogEntryList {
	exampleAuditLogEntry1 := BuildFakeAuditLogEntry()
	exampleAuditLogEntry2 := BuildFakeAuditLogEntry()
	exampleAuditLogEntry3 := BuildFakeAuditLogEntry()

	return &models.AuditLogEntryList{
		Pagination: models.Pagination{
			Page:  1,
			Limit: 20,
		},
		Entries: []models.AuditLogEntry{
			*exampleAuditLogEntry1,
			*exampleAuditLogEntry2,
			*exampleAuditLogEntry3,
		},
	}
}

// BuildFakeAuditLogEntryCreationInput builds a faked AuditLogEntryCreationInput.
func BuildFakeAuditLogEntryCreationInput() *models.AuditLogEntryCreationInput {
	item := BuildFakeAuditLogEntry()
	return BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(item)
}

// BuildFakeAuditLogEntryCreationInputFromAuditLogEntry builds a faked AuditLogEntryCreationInput from an item.
func BuildFakeAuditLogEntryCreationInputFromAuditLogEntry(entry *models.AuditLogEntry) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: entry.EventType,
		Context:   entry.Context,
	}
}
