package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
)

// fetchRandomItem retrieves a random item from the list of available items.
func fetchRandomItem(ctx context.Context, c *httpclient.Client) *types.Item {
	itemsRes, err := c.GetItems(ctx, nil)
	if err != nil || itemsRes == nil || len(itemsRes.Items) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(itemsRes.Items))

	return &itemsRes.Items[randIndex]
}

func buildItemActions(c *httpclient.Client) map[string]*Action {
	return map[string]*Action{
		"CreateItem": {
			Name: "CreateItem",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				itemInput := fakes.BuildFakeItemCreationInput()

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
					newItem := fakes.BuildFakeItemCreationInput()
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
