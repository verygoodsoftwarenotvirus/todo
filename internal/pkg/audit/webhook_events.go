package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

const (
	// WebhookAssignmentKey is the key we use to indicate that an audit log entry is associated with a webhook.
	WebhookAssignmentKey = "webhook_id"

	// WebhookCreationEvent events indicate a user created an item.
	WebhookCreationEvent = "webhook_created"
	// WebhookUpdateEvent events indicate a user updated an item.
	WebhookUpdateEvent = "webhook_updated"
	// WebhookArchiveEvent events indicate a user deleted an item.
	WebhookArchiveEvent = "webhook_archived"
)

// BuildWebhookCreationEventEntry builds an entry creation input for when a webhook is created.
func BuildWebhookCreationEventEntry(webhook *types.Webhook) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: WebhookCreationEvent,
		Context: map[string]interface{}{
			CreationAssignmentKey: webhook,
			WebhookAssignmentKey:  webhook.ID,
		},
	}
}

// BuildWebhookUpdateEventEntry builds an entry creation input for when a webhook is updated.
func BuildWebhookUpdateEventEntry(userID, webhookID uint64, changes []types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: WebhookUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			WebhookAssignmentKey: webhookID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildWebhookArchiveEventEntry builds an entry creation input for when a webhook is archived.
func BuildWebhookArchiveEventEntry(userID, webhookID uint64) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		EventType: WebhookArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   userID,
			WebhookAssignmentKey: webhookID,
		},
	}
}
