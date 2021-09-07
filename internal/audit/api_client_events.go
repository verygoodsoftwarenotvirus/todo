package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

const (
	// APIClientAssignmentKey is the key we use to indicate that an audit log entry is associated with an API client.
	APIClientAssignmentKey = "api_client_id"

	// APIClientCreationEvent events indicate a user created a API client.
	APIClientCreationEvent = "api_client_created"
	// APIClientUpdateEvent events indicate a user updated a API client.
	APIClientUpdateEvent = "api_client_created"
	// APIClientArchiveEvent events indicate a user deleted a API client.
	APIClientArchiveEvent = "api_client_archived"
)

// BuildAPIClientCreationEventEntry builds an entry creation input for when an API client is created.
func BuildAPIClientCreationEventEntry(client *types.APIClient, createdBy string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: APIClientCreationEvent,
		Context: map[string]interface{}{
			APIClientAssignmentKey: client.ID,
			CreationAssignmentKey:  client,
			ActorAssignmentKey:     createdBy,
		},
	}
}

// BuildAPIClientArchiveEventEntry builds an entry creation input for when an API client is archived.
func BuildAPIClientArchiveEventEntry(accountID, clientID, archivedBy string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: APIClientArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:     archivedBy,
			AccountAssignmentKey:   accountID,
			APIClientAssignmentKey: clientID,
		},
	}
}
