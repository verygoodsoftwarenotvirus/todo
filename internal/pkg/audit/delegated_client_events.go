package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// DelegatedClientAssignmentKey is the key we use to indicate that an audit log entry is associated with an oauth2 client.
	DelegatedClientAssignmentKey = "delegated_client_id"

	// DelegatedClientCreationEvent events indicate a user created a delegated client.
	DelegatedClientCreationEvent = "delegated_client_created"
	// DelegatedClientUpdateEvent events indicate a user updated a delegated client.
	DelegatedClientUpdateEvent = "delegated_client_created"
	// DelegatedClientArchiveEvent events indicate a user deleted a delegated client.
	DelegatedClientArchiveEvent = "delegated_client_archived"
)

// BuildDelegatedClientCreationEventEntry builds an entry creation input for when an oauth2 client is created.
func BuildDelegatedClientCreationEventEntry(client *types.DelegatedClient) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: DelegatedClientCreationEvent,
		Context: map[string]interface{}{
			DelegatedClientAssignmentKey: client.ID,
			CreationAssignmentKey:        client,
		},
	}
}

// BuildDelegatedClientUpdateEventEntry builds an entry creation input for when an item is updated.
func BuildDelegatedClientUpdateEventEntry(userID, clientID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: DelegatedClientUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:           userID,
			DelegatedClientAssignmentKey: clientID,
			ChangesAssignmentKey:         changes,
		},
	}
}

// BuildDelegatedClientArchiveEventEntry builds an entry creation input for when an oauth2 client is archived.
func BuildDelegatedClientArchiveEventEntry(userID, clientID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: DelegatedClientArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:           userID,
			DelegatedClientAssignmentKey: clientID,
		},
	}
}
