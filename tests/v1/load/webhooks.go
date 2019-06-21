package main



import (
	"context"
	"math/rand"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

// FetchRandomWebhook retrieves a random webhook from the list of available webhooks
func FetchRandomWebhook(client *client.V1Client) *models.Webhook {
	webhooks, err := client.GetWebhooks(context.Background(), nil)
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
				return c.BuildGetWebhooksRequest(ctx, nil)
			},
			Weight: 100,
		},
		"GetWebhook": {
			Name: "GetWebhook",
			Action: func() (*http.Request, error) {
				if randomWebhook := FetchRandomWebhook(c); randomWebhook != nil {
					return c.BuildGetWebhookRequest(ctx, randomWebhook.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"CreateWebhook": {
			Name: "CreateWebhook",
			Action: func() (*http.Request, error) {
				return c.BuildCreateWebhookRequest(ctx, model.RandomWebhookInput())
			},
			Weight: 1,
		},
		"UpdateWebhook": {
			Name: "UpdateWebhook",
			Action: func() (*http.Request, error) {
				if randomWebhook := FetchRandomWebhook(c); randomWebhook != nil {
					randomWebhook.Name = model.RandomWebhookInput().Name
					return c.BuildUpdateWebhookRequest(ctx, randomWebhook)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 50,
		},
		"ArchiveWebhook": {
			Name: "ArchiveWebhook",
			Action: func() (*http.Request, error) {
				if randomWebhook := FetchRandomWebhook(c); randomWebhook != nil {
					return c.BuildArchiveWebhookRequest(ctx, randomWebhook.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 50,
		},
	}
}
