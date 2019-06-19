package main

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"time"

	http2 "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil/rand/model"
)

var (
	// ErrUnavailableYet is a sentinel error value
	ErrUnavailableYet = errors.New("can't do this yet")
)

type (
	// actionFunc represents a thing you can do
	actionFunc func() (*http.Request, error)

	// Action is a wrapper struct around some important values
	Action struct {
		Action actionFunc
		Weight int
		Name   string
	}
)

func fetchRandomItem(client *http2.V1Client) *models.Item {
	itemsRes, err := client.GetItems(context.Background(), nil)
	if err != nil || itemsRes == nil || len(itemsRes.Items) == 0 {
		return nil
	}

	randIndex := rand.Intn(len(itemsRes.Items))
	return &itemsRes.Items[randIndex]
}

func fetchRandomOAuth2Client(client *http2.V1Client) *models.OAuth2Client {
	clientsRes, err := client.GetOAuth2Clients(context.Background(), nil)
	if err != nil || clientsRes == nil || len(clientsRes.Clients) <= 1 {
		return nil
	}

	var selectedClient *models.OAuth2Client
	for selectedClient == nil {
		ri := rand.Intn(len(clientsRes.Clients))
		c := &clientsRes.Clients[ri]
		if c.ClientID != "FIXME" {
			selectedClient = c
		}
	}

	return selectedClient
}

// RandomAction takes a client and returns a closure which is an action
func RandomAction(client *http2.V1Client) *Action {
	ctx := context.Background()
	allActions := map[string]*Action{
		"GetHealthCheck": {
			Name:   "GetHealthCheck",
			Action: client.BuildHealthCheckRequest,
			Weight: 100,
		},
		/*
			///////////////////
			// Items
		*/
		"CreateItem": {
			Name: "CreateItem",
			Action: func() (*http.Request, error) {
				return client.BuildCreateItemRequest(ctx, model.RandomItemInput())
			},
			Weight: 100,
		},
		"GetItem": {
			Name: "GetItem",
			Action: func() (*http.Request, error) {
				if randomItem := fetchRandomItem(client); randomItem != nil {
					return client.BuildGetItemRequest(ctx, randomItem.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"GetItems": {
			Name: "GetItems",
			Action: func() (*http.Request, error) {
				return client.BuildGetItemsRequest(ctx, nil)
			},
			Weight: 100,
		},
		"UpdateItem": {
			Name: "UpdateItem",
			Action: func() (*http.Request, error) {
				if randomItem := fetchRandomItem(client); randomItem != nil {
					randomItem.Name = model.RandomItemInput().Name
					return client.BuildUpdateItemRequest(ctx, randomItem)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"ArchiveItem": {
			Name: "ArchiveItem",
			Action: func() (*http.Request, error) {
				if randomItem := fetchRandomItem(client); randomItem != nil {
					return client.BuildArchiveItemRequest(ctx, randomItem.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 85,
		},
		/*
			///////////////////
			// Users
		*/
		"CreateUser": {
			Name: "CreateUser",
			Action: func() (*http.Request, error) {
				ui := model.RandomUserInput()
				return client.BuildCreateUserRequest(ctx, ui)
			},
			Weight: 100,
		},
		/*
			///////////////////
			// OAuth2 Clients
		*/
		"CreateOAuth2Client": {
			Name: "CreateOAuth2Client",
			Action: func() (*http.Request, error) {
				ui := model.RandomUserInput()
				u, err := client.CreateUser(ctx, ui)
				if err != nil {
					return client.BuildHealthCheckRequest()
				}

				cookie, err := client.Login(ctx, u.Username, ui.Password, u.TwoFactorSecret)
				if err != nil {
					return client.BuildHealthCheckRequest()
				}

				req, err := client.BuildCreateOAuth2ClientRequest(
					ctx,
					cookie,
					model.RandomOAuth2ClientInput(
						u.Username,
						ui.Password,
						u.TwoFactorSecret,
					),
				)
				return req, err
			},
			Weight: 100,
		},

		"GetOAuth2Client": {
			Name: "GetOAuth2Client",
			Action: func() (*http.Request, error) {
				if randomOAuth2Client := fetchRandomOAuth2Client(client); randomOAuth2Client != nil {
					return client.BuildGetOAuth2ClientRequest(ctx, randomOAuth2Client.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"GetOAuth2Clients": {
			Name: "GetOAuth2Clients",
			Action: func() (*http.Request, error) {
				return client.BuildGetOAuth2ClientsRequest(ctx, nil)
			},
			Weight: 100,
		},
		/*
			///////////////////
			// Webhooks
		*/
		"GetWebhooks": {
			Name: "GetWebhooks",
			Action: func() (*http.Request, error) {
				return client.BuildGetWebhooksRequest(ctx, nil)
			},
			Weight: 100,
		},
		// "CreateWebhook": {
		// 	Name: "CreateWebhook",
		// 	Action: func() (*http.Request, error) {
		// 		return client.BuildCreateWebhookRequest(ctx, model.RandomWebhookInput())
		// 	},
		// 	Weight: 1,
		// },
		// "GetWebhook": {
		// 	Name: "GetWebhook",
		// 	Action: func() (*http.Request, error) {
		// 		if randomWebhook := fetchRandomWebhook(client); randomWebhook != nil {
		// 			return client.BuildGetWebhookRequest(ctx, randomWebhook.ID)
		// 		}
		// 		return nil, ErrUnavailableYet
		// 	},
		// 	Weight: 100,
		// },
		// "UpdateWebhook": {
		// 	Name: "UpdateWebhook",
		// 	Action: func() (*http.Request, error) {
		// 		if randomWebhook := fetchRandomWebhook(client); randomWebhook != nil {
		// 			randomWebhook.Name = model.RandomWebhookInput().Name
		// 			return client.BuildUpdateWebhookRequest(ctx, randomWebhook)
		// 		}
		// 		return nil, ErrUnavailableYet
		// 	},
		// 	Weight: 100,
		// },
		// "ArchiveWebhook": {
		// 	Name: "ArchiveWebhook",
		// 	Action: func() (*http.Request, error) {
		// 		if randomWebhook := fetchRandomWebhook(client); randomWebhook != nil {
		// 			return client.BuildArchiveWebhookRequest(ctx, randomWebhook.ID)
		// 		}
		// 		return nil, ErrUnavailableYet
		// 	},
		// 	Weight: 100,
		// },
		//
	}

	totalWeight := 0
	for _, rb := range allActions {
		totalWeight += rb.Weight
	}

	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(totalWeight)

	for _, rb := range allActions {
		r -= rb.Weight
		if r <= 0 {
			return rb
		}
	}

	return nil
}
