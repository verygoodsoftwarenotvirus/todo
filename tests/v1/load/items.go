package main

import (
	"context"
	"math/rand"
	"net/http"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil/rand/model"
)

// FetchRandomItem retrieves a random item from the list of available items
func FetchRandomItem(c *client.V1Client) *models.Item {
	itemsRes, err := c.GetItems(context.Background(), nil)
	if err != nil || itemsRes == nil || len(itemsRes.Items) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(itemsRes.Items))
	return &itemsRes.Items[randIndex]
}

func buildItemActions(c *client.V1Client) map[string]*Action {
	return map[string]*Action{
		"CreateItem": {
			Name: "CreateItem",
			Action: func() (*http.Request, error) {
				return c.BuildCreateItemRequest(context.Background(), model.RandomItemCreationInput())
			},
			Weight: 100,
		},
		"GetItem": {
			Name: "GetItem",
			Action: func() (*http.Request, error) {
				if randomItem := FetchRandomItem(c); randomItem != nil {
					return c.BuildGetItemRequest(context.Background(), randomItem.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"GetItems": {
			Name: "GetItems",
			Action: func() (*http.Request, error) {
				return c.BuildGetItemsRequest(context.Background(), nil)
			},
			Weight: 100,
		},
		"UpdateItem": {
			Name: "UpdateItem",
			Action: func() (*http.Request, error) {
				if randomItem := FetchRandomItem(c); randomItem != nil {
					randomItem.Name = model.RandomItemCreationInput().Name
					return c.BuildUpdateItemRequest(context.Background(), randomItem)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"ArchiveItem": {
			Name: "ArchiveItem",
			Action: func() (*http.Request, error) {
				if randomItem := FetchRandomItem(c); randomItem != nil {
					return c.BuildArchiveItemRequest(context.Background(), randomItem.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 85,
		},
	}
}
