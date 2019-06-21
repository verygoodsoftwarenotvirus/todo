package model

import (
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
)

// RandomWebhookInput creates a random WebhookCreationInput
func RandomWebhookInput() *models.WebhookCreationInput {
	x := &models.WebhookCreationInput{
		Name:        fake.Word(),
		URL:         fake.DomainName(),
		ContentType: "application/json",
		Method:      "POST",
	}
	return x
}
