package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient/requests"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
)

// fetchRandomItem retrieves a random item from the list of available items.
func fetchRandomItem(ctx context.Context, c *httpclient.Client) *types.Item {
	itemsRes, err := c.GetItems(ctx, nil)
	if err != nil || itemsRes == nil || len(itemsRes.Items) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(itemsRes.Items))

	return itemsRes.Items[randIndex]
}

func buildItemActions(c *httpclient.Client, builder *requests.Builder) map[string]*Action {
	return map[string]*Action{
		"CreateItem": {
			Name: "CreateItem",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				itemInput := fakes.BuildFakeItemCreationInput()

				return builder.BuildCreateItemRequest(ctx, itemInput)
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

				return builder.BuildGetItemRequest(ctx, randomItem.ID)
			},
			Weight: 100,
		},
		"GetItems": {
			Name: "GetItems",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				return builder.BuildGetItemsRequest(ctx, nil)
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
					return builder.BuildUpdateItemRequest(ctx, randomItem)
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

				return builder.BuildArchiveItemRequest(ctx, randomItem.ID)
			},
			Weight: 85,
		},
	}
}
