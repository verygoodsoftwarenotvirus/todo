package fakes

import (
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	fake "github.com/brianvoe/gofakeit/v5"
)

// BuildFakeDelegatedClient builds a faked DelegatedClient.
func BuildFakeDelegatedClient() *types.DelegatedClient {
	return &types.DelegatedClient{
		ID:            uint64(fake.Uint32()),
		ExternalID:    fake.UUID(),
		Name:          fake.Word(),
		ClientID:      fake.UUID(),
		ClientSecret:  []byte(fake.UUID()),
		BelongsToUser: fake.Uint64(),
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakeDelegatedClientCreationResponseFromClient builds a faked DelegatedClientCreationResponse.
func BuildFakeDelegatedClientCreationResponseFromClient(client *types.DelegatedClient) *types.DelegatedClientCreationResponse {
	return &types.DelegatedClientCreationResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: string(client.ClientSecret),
	}
}

// BuildFakeDelegatedClientList builds a faked DelegatedClientList.
func BuildFakeDelegatedClientList() *types.DelegatedClientList {
	var examples []*types.DelegatedClient
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeDelegatedClient())
	}

	return &types.DelegatedClientList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		Clients: examples,
	}
}

// BuildFakeDelegatedClientCreationInput builds a faked DelegatedClientCreationInput.
func BuildFakeDelegatedClientCreationInput() *types.DelegatedClientCreationInput {
	client := BuildFakeDelegatedClient()

	return &types.DelegatedClientCreationInput{
		UserLoginInput: types.UserLoginInput{
			Username:  fake.Username(),
			Password:  fake.Password(true, true, true, true, true, 32),
			TOTPToken: fmt.Sprintf("0%s", fake.Zip()),
		},
		Name:          client.Name,
		ClientID:      client.ClientID,
		BelongsToUser: client.BelongsToUser,
	}
}

// BuildFakeDelegatedClientCreationInputFromClient builds a faked DelegatedClientCreationInput.
func BuildFakeDelegatedClientCreationInputFromClient(client *types.DelegatedClient) *types.DelegatedClientCreationInput {
	return &types.DelegatedClientCreationInput{
		UserLoginInput: types.UserLoginInput{
			Username:  fake.Username(),
			Password:  fake.Password(true, true, true, true, true, 32),
			TOTPToken: fmt.Sprintf("0%s", fake.Zip()),
		},
		Name:          client.Name,
		ClientID:      client.ClientID,
		ClientSecret:  client.ClientSecret,
		BelongsToUser: client.BelongsToUser,
	}
}
