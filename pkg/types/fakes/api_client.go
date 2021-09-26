package fakes

import (
	"fmt"

	fake "github.com/brianvoe/gofakeit/v5"
	"github.com/segmentio/ksuid"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
)

// BuildFakeAPIClient builds a faked APIClient.
func BuildFakeAPIClient() *types.APIClient {
	return &types.APIClient{
		ID:            ksuid.New().String(),
		Name:          fake.Password(true, true, true, false, false, 32),
		ClientID:      ksuid.New().String(),
		ClientSecret:  []byte(fake.Password(true, true, true, true, true, 32)),
		BelongsToUser: fake.UUID(),
		CreatedOn:     uint64(uint32(fake.Date().Unix())),
	}
}

// BuildFakeAPIClientCreationResponseFromClient builds a faked APIClientCreationResponse.
func BuildFakeAPIClientCreationResponseFromClient(client *types.APIClient) *types.APIClientCreationResponse {
	return &types.APIClientCreationResponse{
		ID:           client.ID,
		ClientID:     client.ClientID,
		ClientSecret: string(client.ClientSecret),
	}
}

// BuildFakeAPIClientList builds a faked APIClientList.
func BuildFakeAPIClientList() *types.APIClientList {
	var examples []*types.APIClient
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeAPIClient())
	}

	return &types.APIClientList{
		Pagination: types.Pagination{
			Page:          1,
			Limit:         20,
			FilteredCount: exampleQuantity / 2,
			TotalCount:    exampleQuantity,
		},
		Clients: examples,
	}
}

// BuildFakeAPIClientCreationInput builds a faked APIClientCreationInput.
func BuildFakeAPIClientCreationInput() *types.APIClientCreationInput {
	client := BuildFakeAPIClient()

	return &types.APIClientCreationInput{
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

// BuildFakeAPIClientCreationInputFromClient builds a faked APIClientCreationInput.
func BuildFakeAPIClientCreationInputFromClient(client *types.APIClient) *types.APIClientCreationInput {
	return &types.APIClientCreationInput{
		ID: client.ID,
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
