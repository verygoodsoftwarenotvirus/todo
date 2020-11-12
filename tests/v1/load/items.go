package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/fake"
)

// fetchRandomItem retrieves a random item from the list of available items.
func fetchRandomItem(ctx context.Context, c *client.V1Client) *models.Item {
	itemsRes, err := c.GetItems(ctx, nil)
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
				ctx := context.Background()

				itemInput := fakemodels.BuildFakeItemCreationInput()

				return c.BuildCreateItemRequest(ctx, itemInput)
			},
			Weight: 100,
		},
		"GetItem": {
			Name: "GetItem",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				randomItem := fetchRandomItem(ctx, c)
				if randomItem == nil {
					return nil, fmt.Errorf("retrieving random item: %w", ErrUnavailableYet)
				}

				return c.BuildGetItemRequest(ctx, randomItem.ID)
			},
			Weight: 100,
		},
		"GetItems": {
			Name: "GetItems",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				return c.BuildGetItemsRequest(ctx, nil)
			},
			Weight: 100,
		},
		"UpdateItem": {
			Name: "UpdateItem",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				if randomItem := fetchRandomItem(ctx, c); randomItem != nil {
					newItem := fakemodels.BuildFakeItemCreationInput()
					randomItem.Name = newItem.Name
					randomItem.Details = newItem.Details
					return c.BuildUpdateItemRequest(ctx, randomItem)
				}

				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"ArchiveItem": {
			Name: "ArchiveItem",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				randomItem := fetchRandomItem(ctx, c)
				if randomItem == nil {
					return nil, fmt.Errorf("retrieving random item: %w", ErrUnavailableYet)
				}

				return c.BuildArchiveItemRequest(ctx, randomItem.ID)
			},
			Weight: 85,
		},
	}
}
