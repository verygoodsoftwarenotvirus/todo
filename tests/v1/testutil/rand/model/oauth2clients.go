package randmodel

import (
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

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
