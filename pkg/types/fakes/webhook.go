package fakes

import (
	"net/http"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/segmentio/ksuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// BuildFakeWebhook builds a faked Webhook.
func BuildFakeWebhook() *types.Webhook {
	return &types.Webhook{
		ID:               ksuid.New().String(),
		Name:             fake.UUID(),
		ContentType:      "application/json",
		URL:              fake.URL(),
		Method:           http.MethodPost,
		Events:           []string{fake.Word()},
		DataTypes:        []string{fake.Word()},
		Topics:           []string{fake.Word()},
		CreatedOn:        uint64(uint32(fake.Date().Unix())),
		ArchivedOn:       nil,
		BelongsToAccount: fake.UUID(),
	}
}

// BuildFakeWebhookList builds a faked WebhookList.
func BuildFakeWebhookList() *types.WebhookList {
	var examples []*types.Webhook
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeWebhook())
	}

	return &types.WebhookList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		Webhooks: examples,
	}
}

// BuildFakeWebhookCreationInput builds a faked WebhookCreationInput from an item.
func BuildFakeWebhookCreationInput() *types.WebhookCreationInput {
	webhook := BuildFakeWebhook()
	return BuildFakeWebhookCreationInputFromWebhook(webhook)
}

// BuildFakeWebhookDatabaseCreationInput builds a faked WebhookCreationInput from an item.
func BuildFakeWebhookDatabaseCreationInput() *types.WebhookDatabaseCreationInput {
	webhook := BuildFakeWebhook()
	return BuildFakeWebhookDatabaseCreationInputFromWebhook(webhook)
}

// BuildFakeWebhookCreationInputFromWebhook builds a faked WebhookCreationInput.
func BuildFakeWebhookCreationInputFromWebhook(webhook *types.Webhook) *types.WebhookCreationInput {
	return &types.WebhookCreationInput{
		ID:               webhook.ID,
		Name:             webhook.Name,
		ContentType:      webhook.ContentType,
		URL:              webhook.URL,
		Method:           webhook.Method,
		Events:           webhook.Events,
		DataTypes:        webhook.DataTypes,
		Topics:           webhook.Topics,
		BelongsToAccount: webhook.BelongsToAccount,
	}
}

// BuildFakeWebhookDatabaseCreationInputFromWebhook builds a faked WebhookCreationInput.
func BuildFakeWebhookDatabaseCreationInputFromWebhook(webhook *types.Webhook) *types.WebhookDatabaseCreationInput {
	return &types.WebhookDatabaseCreationInput{
		ID:               webhook.ID,
		Name:             webhook.Name,
		ContentType:      webhook.ContentType,
		URL:              webhook.URL,
		Method:           webhook.Method,
		Events:           webhook.Events,
		DataTypes:        webhook.DataTypes,
		Topics:           webhook.Topics,
		BelongsToAccount: webhook.BelongsToAccount,
	}
}
