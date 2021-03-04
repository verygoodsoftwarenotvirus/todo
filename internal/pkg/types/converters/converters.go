package converters

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
)

// ConvertAuditLogEntryCreationInputToEntry converts an AuditLogEntryCreationInput to an AuditLogEntry.
func ConvertAuditLogEntryCreationInputToEntry(e *types.AuditLogEntryCreationInput) *types.AuditLogEntry {
	return &types.AuditLogEntry{
		EventType: e.EventType,
		Context:   e.Context,
	}
}

// ConvertAccountToAccountUpdateInput creates an AccountUpdateInput struct from an item.
func ConvertAccountToAccountUpdateInput(x *types.Account) *types.AccountUpdateInput {
	return &types.AccountUpdateInput{
		Name:          x.Name,
		BelongsToUser: x.BelongsToUser,
	}
}

// ConvertWebhookToWebhookUpdateInput creates an WebhookUpdateInput struct from a webhook.
func ConvertWebhookToWebhookUpdateInput(x *types.Webhook) *types.WebhookUpdateInput {
	return &types.WebhookUpdateInput{
		Name:             x.Name,
		ContentType:      x.ContentType,
		URL:              x.URL,
		Method:           x.Method,
		Events:           x.Events,
		DataTypes:        x.DataTypes,
		Topics:           x.Topics,
		BelongsToAccount: x.BelongsToAccount,
	}
}

// ConvertItemToItemUpdateInput creates an ItemUpdateInput struct from an item.
func ConvertItemToItemUpdateInput(x *types.Item) *types.ItemUpdateInput {
	return &types.ItemUpdateInput{
		Name:    x.Name,
		Details: x.Details,
	}
}

// ConvertAccountSubscriptionPlanToPlanUpdateInput creates an AccountSubscriptionPlanUpdateInput struct from a plan.
func ConvertAccountSubscriptionPlanToPlanUpdateInput(x *types.AccountSubscriptionPlan) *types.AccountSubscriptionPlanUpdateInput {
	return &types.AccountSubscriptionPlanUpdateInput{
		Name:        x.Name,
		Description: x.Description,
		Price:       x.Price,
		Period:      x.Period,
	}
}
