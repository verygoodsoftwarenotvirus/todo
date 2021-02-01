package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// OAuth2ClientAssignmentKey is the key we use to indicate that an audit log entry is associated with an oauth2 client.
	OAuth2ClientAssignmentKey = "client_id"

	// OAuth2ClientCreationEvent events indicate a user created an item.
	OAuth2ClientCreationEvent = "oauth2_client_created"
	// OAuth2ClientArchiveEvent events indicate a user deleted an item.
	OAuth2ClientArchiveEvent = "oauth2_client_archived"
)

// BuildOAuth2ClientCreationEventEntry builds an entry creation input for when an oauth2 client is created.
func BuildOAuth2ClientCreationEventEntry(client *types.OAuth2Client) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: OAuth2ClientCreationEvent,
		Context: map[string]interface{}{
			OAuth2ClientAssignmentKey: client.ID,
			CreationAssignmentKey:     client,
		},
	}
}

// BuildOAuth2ClientArchiveEventEntry builds an entry creation input for when an oauth2 client is archived.
func BuildOAuth2ClientArchiveEventEntry(userID, clientID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: OAuth2ClientArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:        userID,
			OAuth2ClientAssignmentKey: clientID,
		},
	}
}
