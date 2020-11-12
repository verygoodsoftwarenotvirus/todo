package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
)

const (
	// OAuth2ClientAssignmentKey is the key we use to indicate that an audit log entry is associated with an oauth2 client.
	OAuth2ClientAssignmentKey = "client_id"
)

// BuildOAuth2ClientCreationEventEntry builds an entry creation input for when an oauth2 client is created.
func BuildOAuth2ClientCreationEventEntry(client *models.OAuth2Client) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: OAuth2ClientCreationEvent,
		Context: map[string]interface{}{
			OAuth2ClientAssignmentKey: client.ID,
			CreationAssignmentKey:     client,
		},
	}
}

// BuildOAuth2ClientArchiveEventEntry builds an entry creation input for when an oauth2 client is archived.
func BuildOAuth2ClientArchiveEventEntry(userID, clientID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: OAuth2ClientArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:        userID,
			OAuth2ClientAssignmentKey: clientID,
		},
	}
}
