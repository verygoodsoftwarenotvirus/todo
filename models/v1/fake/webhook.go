package fakemodels

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeWebhook builds a faked Webhook
func BuildFakeWebhook() *models.Webhook {
	return &models.Webhook{
		ID:            fake.Uint64(),
		Name:          fake.Word(),
		ContentType:   fake.FileMimeType(),
		URL:           fake.URL(),
		Method:        fake.HTTPMethod(),
		Events:        []string{fake.Word()},
		DataTypes:     []string{fake.Word()},
		Topics:        []string{fake.Word()},
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
		ArchivedOn:    nil,
		BelongsToUser: fake.Uint64(),
	}
}

// BuildFakeWebhookList builds a faked WebhookList
func BuildFakeWebhookList() *models.WebhookList {
	exampleWebhook1 := BuildFakeWebhook()
	exampleWebhook2 := BuildFakeWebhook()
	exampleWebhook3 := BuildFakeWebhook()
	return &models.WebhookList{
		Pagination: models.Pagination{
			Page:       1,
			Limit:      20,
			TotalCount: 3,
		},
		Webhooks: []models.Webhook{
			*exampleWebhook1,
			*exampleWebhook2,
			*exampleWebhook3,
		},
	}
}

// BuildFakeWebhookUpdateInputFromWebhook builds a faked WebhookUpdateInput
func BuildFakeWebhookUpdateInputFromWebhook(webhook *models.Webhook) *models.WebhookUpdateInput {
	return &models.WebhookUpdateInput{
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

// BuildFakeWebhookCreationInput builds a faked WebhookCreationInput
func BuildFakeWebhookCreationInput() *models.WebhookCreationInput {
	webhook := BuildFakeWebhook()
	return BuildFakeWebhookCreationInputFromWebhook(webhook)
}

// BuildFakeWebhookCreationInputFromWebhook builds a faked WebhookCreationInput
func BuildFakeWebhookCreationInputFromWebhook(webhook *models.Webhook) *models.WebhookCreationInput {
	return &models.WebhookCreationInput{
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
