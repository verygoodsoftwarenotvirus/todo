package fakes

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeOAuth2Client builds a faked OAuth2Client.
func BuildFakeOAuth2Client() *types.OAuth2Client {
	return &types.OAuth2Client{
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
func BuildFakeOAuth2ClientList() *types.OAuth2ClientList {
	var examples []types.OAuth2Client
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, *BuildFakeOAuth2Client())
	}

	return &types.OAuth2ClientList{
		Pagination: types.Pagination{
			Page:       1,
			Limit:      20,
			TotalCount: exampleQuantity,
		},
		Clients: examples,
	}
}

// BuildFakeOAuth2ClientCreationInputFromClient builds a faked OAuth2ClientCreationInput.
func BuildFakeOAuth2ClientCreationInputFromClient(client *types.OAuth2Client) *types.OAuth2ClientCreationInput {
	return &types.OAuth2ClientCreationInput{
		UserLoginInput: types.UserLoginInput{
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
