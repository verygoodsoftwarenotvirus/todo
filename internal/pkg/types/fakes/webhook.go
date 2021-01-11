package fakes

import (
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeWebhook builds a faked Webhook.
func BuildFakeWebhook() *types.Webhook {
	return &types.Webhook{
		ID:            uint64(fake.Uint32()),
		Name:          fake.Word(),
		ContentType:   "application/json",
		URL:           fake.URL(),
		Method:        http.MethodPost,
		Events:        []string{fake.Word()},
		DataTypes:     []string{fake.Word()},
		Topics:        []string{fake.Word()},
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		ArchivedOn:    nil,
		BelongsToUser: fake.Uint64(),
	}
}

// BuildFakeWebhookList builds a faked WebhookList.
func BuildFakeWebhookList() *types.WebhookList {
	var examples []types.Webhook
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, *BuildFakeWebhook())
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

// BuildFakeWebhookUpdateInputFromWebhook builds a faked WebhookUpdateInput.
func BuildFakeWebhookUpdateInputFromWebhook(webhook *types.Webhook) *types.WebhookUpdateInput {
	return &types.WebhookUpdateInput{
		Name:          webhook.Name,
		ContentType:   webhook.ContentType,
		URL:           webhook.URL,
		Method:        webhook.Method,
		Events:        webhook.Events,
		DataTypes:     webhook.DataTypes,
		Topics:        webhook.Topics,
		BelongsToUser: webhook.BelongsToUser,
	}
}

// BuildFakeWebhookCreationInput builds a faked WebhookCreationInput.
func BuildFakeWebhookCreationInput() *types.WebhookCreationInput {
	webhook := BuildFakeWebhook()
	return BuildFakeWebhookCreationInputFromWebhook(webhook)
}

// BuildFakeWebhookCreationInputFromWebhook builds a faked WebhookCreationInput.
func BuildFakeWebhookCreationInputFromWebhook(webhook *types.Webhook) *types.WebhookCreationInput {
	return &types.WebhookCreationInput{
		Name:          webhook.Name,
		ContentType:   webhook.ContentType,
		URL:           webhook.URL,
		Method:        webhook.Method,
		Events:        webhook.Events,
		DataTypes:     webhook.DataTypes,
		Topics:        webhook.Topics,
		BelongsToUser: webhook.BelongsToUser,
	}
}
