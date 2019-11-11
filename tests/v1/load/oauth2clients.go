package main

import (
	"context"
	"math/rand"
	"net/http"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	randmodel "gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil/rand/model"
)

// fetchRandomOAuth2Client retrieves a random client from the list of available clients
func fetchRandomOAuth2Client(c *client.V1Client) *models.OAuth2Client {
	clientsRes, err := c.GetOAuth2Clients(context.Background(), nil)
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

func buildOAuth2ClientActions(c *client.V1Client) map[string]*Action {
	return map[string]*Action{
		"CreateOAuth2Client": {
			Name: "CreateOAuth2Client",
			Action: func() (*http.Request, error) {
				ui := randmodel.RandomUserInput()
				u, err := c.CreateUser(context.Background(), ui)
				if err != nil {
					return c.BuildHealthCheckRequest()
				}

				cookie, err := c.Login(context.Background(), u.Username, ui.Password, u.TwoFactorSecret)
				if err != nil {
					return c.BuildHealthCheckRequest()
				}

				req, err := c.BuildCreateOAuth2ClientRequest(
					context.Background(),
					cookie,
					randmodel.RandomOAuth2ClientInput(
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
				if randomOAuth2Client := fetchRandomOAuth2Client(c); randomOAuth2Client != nil {
					return c.BuildGetOAuth2ClientRequest(context.Background(), randomOAuth2Client.ID)
				}
				return nil, ErrUnavailableYet
			},
			Weight: 100,
		},
		"GetOAuth2Clients": {
			Name: "GetOAuth2Clients",
			Action: func() (*http.Request, error) {
				return c.BuildGetOAuth2ClientsRequest(context.Background(), nil)
			},
			Weight: 100,
		},
	}
}
