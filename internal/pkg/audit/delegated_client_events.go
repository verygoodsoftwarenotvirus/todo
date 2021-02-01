package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// DelegatedClientAssignmentKey is the key we use to indicate that an audit log entry is associated with an oauth2 client.
	DelegatedClientAssignmentKey = "delegated_client_id"

	// DelegatedClientCreationEvent events indicate a user created an item.
	DelegatedClientCreationEvent = "oauth2_client_created"
	// DelegatedClientArchiveEvent events indicate a user deleted an item.
	DelegatedClientArchiveEvent = "oauth2_client_archived"
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
