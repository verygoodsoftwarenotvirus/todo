package main

import (
	"context"
	"math/rand"
	"net/http"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	randmodel "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil/rand/model"
)

// fetchRandomWebhook retrieves a random webhook from the list of available webhooks
func fetchRandomWebhook(c *client.V1Client) *models.Webhook {
	webhooks, err := c.GetWebhooks(context.Background(), nil)
	if err != nil || webhooks == nil || len(webhooks.Webhooks) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(webhooks.Webhooks))
	return &webhooks.Webhooks[randIndex]
}

func buildWebhookActions(c *client.V1Client) map[string]*Action {
	return map[string]*Action{
		"GetWebhooks": {
			Name: "GetWebhooks",
			Action: func() (*http.Request, error) {
				return c.BuildGetWebhooksRequest(context.Background(), nil)
			},
			Weight: 100,
		},
		"GetWebhook": {
			Name: "GetWebhook",
			Action: func() (*http.Request, error) {
				if randomWebhook := fetchRandomWebhook(c); randomWebhook != nil {
					return c.BuildGetWebhookRequest(context.Background(), randomWebhook.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"CreateWebhook": {
			Name: "CreateWebhook",
			Action: func() (*http.Request, error) {
				return c.BuildCreateWebhookRequest(context.Background(), randmodel.RandomWebhookInput())
			},
			Weight: 1,
		},
		"UpdateWebhook": {
			Name: "UpdateWebhook",
			Action: func() (*http.Request, error) {
				if randomWebhook := fetchRandomWebhook(c); randomWebhook != nil {
					randomWebhook.Name = randmodel.RandomWebhookInput().Name
					return c.BuildUpdateWebhookRequest(context.Background(), randomWebhook)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 50,
		},
		"ArchiveWebhook": {
			Name: "ArchiveWebhook",
			Action: func() (*http.Request, error) {
				if randomWebhook := fetchRandomWebhook(c); randomWebhook != nil {
					return c.BuildArchiveWebhookRequest(context.Background(), randomWebhook.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 50,
		},
	}
}
