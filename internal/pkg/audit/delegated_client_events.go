package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// APIClientAssignmentKey is the key we use to indicate that an audit log entry is associated with an oauth2 client.
	APIClientAssignmentKey = "api_client_id"

	// APIClientCreationEvent events indicate a user created a API client.
	APIClientCreationEvent = "api_client_created"
	// APIClientUpdateEvent events indicate a user updated a API client.
	APIClientUpdateEvent = "api_client_created"
	// APIClientArchiveEvent events indicate a user deleted a API client.
	APIClientArchiveEvent = "api_client_archived"
)

// BuildAPIClientCreationEventEntry builds an entry creation input for when an oauth2 client is created.
func BuildAPIClientCreationEventEntry(client *types.APIClient) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: APIClientCreationEvent,
		Context: map[string]interface{}{
			APIClientAssignmentKey: client.ID,
			CreationAssignmentKey:  client,
		},
	}
}

// BuildAPIClientArchiveEventEntry builds an entry creation input for when an oauth2 client is archived.
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
