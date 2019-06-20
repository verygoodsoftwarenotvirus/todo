package model

import (
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/icrowley/fake"
	"github.com/pquerna/otp/totp"
)

// RandomItemCreationInput creates a random ItemInput
func RandomItemCreationInput() *models.ItemCreationInput {
	x := &models.ItemCreationInput{
		Name:    fake.Word(),
		Details: fake.Sentence(),
	}

	return x
}

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

func mustBuildCode(totpSecret string) string {
	code, err := totp.GenerateCode(totpSecret, time.Now().UTC())
	if err != nil {
		panic(err)
	}
	return code
}

// RandomOAuth2ClientInput creates a random OAuth2ClientCreationInput
func RandomOAuth2ClientInput(username, password, totpToken string) *models.OAuth2ClientCreationInput {
	x := &models.OAuth2ClientCreationInput{
		UserLoginInput: models.UserLoginInput{
			Username:  username,
			Password:  password,
			TOTPToken: mustBuildCode(totpToken),
		},
	}

	return x
}

// RandomUserInput creates a random UserInput
func RandomUserInput() *models.UserInput {
	// I had difficulty ensuring these values were unique, even when fake.Seed was called. Could've been fake's fault,
	// could've been docker's fault. In either case, it wasn't worth the time to investigate and determine the culprit.
	username := fake.UserName() + fake.HexColor() + fake.Country()
	x := &models.UserInput{
		Username: username,
		Password: fake.Password(64, 128, true, true, true),
	}
	return x
}
