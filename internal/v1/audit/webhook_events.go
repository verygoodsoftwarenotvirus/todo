package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	// WebhookAssignmentKey is the key we use to indicate that an audit log entry is associated with a webhook.
	WebhookAssignmentKey = "webhook_id"
)

// BuildWebhookCreationEventEntry builds an entry creation input for when a webhook is created.
func BuildWebhookCreationEventEntry(webhook *models.Webhook) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: WebhookCreationEvent,
		Context: map[string]interface{}{
			CreationAssignmentKey: webhook,
			WebhookAssignmentKey:  webhook.ID,
		},
	}
}

// BuildWebhookUpdateEventEntry builds an entry creation input for when a webhook is updated.
func BuildWebhookUpdateEventEntry(userID, webhookID uint64, changes []models.FieldChangeSummary) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: WebhookUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			WebhookAssignmentKey: webhookID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildWebhookArchiveEventEntry builds an entry creation input for when a webhook is archived.
func BuildWebhookArchiveEventEntry(userID, webhookID uint64) *models.AuditLogEntryCreationInput {
	return &models.AuditLogEntryCreationInput{
		EventType: WebhookArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			WebhookAssignmentKey: webhookID,
		},
	}
}
