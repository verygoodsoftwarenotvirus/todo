package fakemodels

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeOAuth2Client builds a faked OAuth2Client.
func BuildFakeOAuth2Client() *models.OAuth2Client {
	return &models.OAuth2Client{
		ID:           fake.Uint64(),
		Name:         fake.Word(),
		ClientID:     fake.UUID(),
		ClientSecret: fake.UUID(),
		RedirectURI:  fake.URL(),
		Scopes: []string{
			fake.Word(),
			fake.Word(),
			fake.Word(),
		},
		ImplicitAllowed: false,
		BelongsToUser:   fake.Uint64(),
		CreatedOn:       uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakeOAuth2ClientList builds a faked OAuth2ClientList.
func BuildFakeOAuth2ClientList() *models.OAuth2ClientList {
	exampleOAuth2Client1 := BuildFakeOAuth2Client()
	exampleOAuth2Client2 := BuildFakeOAuth2Client()
	exampleOAuth2Client3 := BuildFakeOAuth2Client()

	return &models.OAuth2ClientList{
		Pagination: models.Pagination{
			Page:  1,
			Limit: 20,
		},
		Clients: []models.OAuth2Client{
			*exampleOAuth2Client1,
			*exampleOAuth2Client2,
			*exampleOAuth2Client3,
		},
	}
}

// BuildFakeOAuth2ClientCreationInputFromClient builds a faked OAuth2ClientCreationInput.
func BuildFakeOAuth2ClientCreationInputFromClient(client *models.OAuth2Client) *models.OAuth2ClientCreationInput {
	return &models.OAuth2ClientCreationInput{
		UserLoginInput: models.UserLoginInput{
			Username:  fake.Username(),
			Password:  fake.Password(true, true, true, true, true, 32),
			TOTPToken: fmt.Sprintf("0%s", fake.Zip()),
		},
		Name:          client.Name,
		Scopes:        client.Scopes,
		ClientID:      client.ClientID,
		ClientSecret:  client.ClientSecret,
		RedirectURI:   client.RedirectURI,
		BelongsToUser: client.BelongsToUser,
	}
}
