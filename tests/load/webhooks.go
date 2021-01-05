package main

import (
	"context"
	"math/rand"
	"net/http"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"
)

// fetchRandomWebhook retrieves a random webhook from the list of available webhooks.
func fetchRandomWebhook(c *httpclient.Client) *types.Webhook {
	webhooks, err := c.GetWebhooks(context.Background(), nil)
	if err != nil || webhooks == nil || len(webhooks.Webhooks) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(webhooks.Webhooks))

	return &webhooks.Webhooks[randIndex]
}

func buildWebhookActions(c *httpclient.Client) map[string]*Action {
	return map[string]*Action{
		"GetWebhooks": {
			Name: "GetWebhooks",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				return c.BuildGetWebhooksRequest(ctx, nil)
			},
			Weight: 100,
		},
		"GetWebhook": {
			Name: "GetWebhook",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				if randomWebhook := fetchRandomWebhook(c); randomWebhook != nil {
					return c.BuildGetWebhookRequest(ctx, randomWebhook.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"CreateWebhook": {
			Name: "CreateWebhook",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				exampleInput := fakes.BuildFakeWebhookCreationInput()
				return c.BuildCreateWebhookRequest(ctx, exampleInput)
			},
			Weight: 1,
		},
		"UpdateWebhook": {
			Name: "UpdateWebhook",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				if randomWebhook := fetchRandomWebhook(c); randomWebhook != nil {
					randomWebhook.Name = fakes.BuildFakeWebhook().Name
					return c.BuildUpdateWebhookRequest(ctx, randomWebhook)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 50,
		},
		"ArchiveWebhook": {
			Name: "ArchiveWebhook",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				if randomWebhook := fetchRandomWebhook(c); randomWebhook != nil {
					return c.BuildArchiveWebhookRequest(ctx, randomWebhook.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 50,
		},
	}
}
