package main

import (
	"context"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient/requests"
	"math/rand"
	"net/http"

	models "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	client "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
)

// fetchRandomItem retrieves a random item from the list of available items.
func fetchRandomItem(ctx context.Context, c *client.Client) *models.Item {
	itemsRes, err := c.GetItems(ctx, nil)
	if err != nil || itemsRes == nil || len(itemsRes.Items) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(itemsRes.Items))

	return itemsRes.Items[randIndex]
}

func buildItemActions(c *client.Client, builder *requests.Builder) map[string]*Action {
	return map[string]*Action{
		"CreateItem": {
			Name: "CreateItem",
			Action: func() (*http.Request, error) {
				ctx := context.Background()

				itemInput := fakemodels.BuildFakeItemCreationInput()

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
					newItem := fakemodels.BuildFakeItemCreationInput()
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
