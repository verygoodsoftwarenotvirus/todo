package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

const (
	// APIClientAssignmentKey is the key we use to indicate that an audit log entry is associated with an API client.
	APIClientAssignmentKey = "api_client_id"

	// APIClientCreationEvent events indicate a user created an API client.
	APIClientCreationEvent = "api_client_created"
	// APIClientArchiveEvent events indicate a user deleted an API client.
	APIClientArchiveEvent = "api_client_archived"
)

// BuildAPIClientCreationEventEntry builds an entry creation input for when an API client is created.
func BuildAPIClientCreationEventEntry(client *types.APIClient, createdBy uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: APIClientCreationEvent,
		Context: map[string]interface{}{
			APIClientAssignmentKey: client.ID,
			CreationAssignmentKey:  client,
			ActorAssignmentKey:     createdBy,
		},
	}
}

// BuildAPIClientArchiveEventEntry builds an entry creation input for when an API client is archived.
func BuildAPIClientArchiveEventEntry(accountID, clientID, archivedBy uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: APIClientArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:     archivedBy,
			AccountAssignmentKey:   accountID,
			APIClientAssignmentKey: clientID,
		},
	}
}
