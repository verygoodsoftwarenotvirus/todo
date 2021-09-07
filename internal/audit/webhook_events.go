package audit

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"

	"github.com/segmentio/ksuid"
)

const (
	// WebhookAssignmentKey is the key we use to indicate that an audit log entry is associated with a webhook.
	WebhookAssignmentKey = "webhook_id"

	// WebhookCreationEvent events indicate a user created a webhook.
	WebhookCreationEvent = "webhook_created"
	// WebhookUpdateEvent events indicate a user updated a webhook.
	WebhookUpdateEvent = "webhook_updated"
	// WebhookArchiveEvent events indicate a user deleted a webhook.
	WebhookArchiveEvent = "webhook_archived"
)

// BuildWebhookCreationEventEntry builds an entry creation input for when a webhook is created.
func BuildWebhookCreationEventEntry(webhook *types.Webhook, createdByUser string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: WebhookCreationEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:    createdByUser,
			CreationAssignmentKey: webhook,
			WebhookAssignmentKey:  webhook.ID,
			AccountAssignmentKey:  webhook.BelongsToAccount,
		},
	}
}

// BuildWebhookUpdateEventEntry builds an entry creation input for when a webhook is updated.
func BuildWebhookUpdateEventEntry(changedByUser, accountID, webhookID string, changes []*types.FieldChangeSummary) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: WebhookUpdateEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   changedByUser,
			AccountAssignmentKey: accountID,
			WebhookAssignmentKey: webhookID,
			ChangesAssignmentKey: changes,
		},
	}
}

// BuildWebhookArchiveEventEntry builds an entry creation input for when a webhook is archived.
func BuildWebhookArchiveEventEntry(archivedByUser, accountID, webhookID string) *types.AuditLogEntryCreationInput {
	return &types.AuditLogEntryCreationInput{
		ID:        ksuid.New().String(),
		EventType: WebhookArchiveEvent,
		Context: map[string]interface{}{
			ActorAssignmentKey:   archivedByUser,
			AccountAssignmentKey: accountID,
			WebhookAssignmentKey: webhookID,
		},
	}
}
