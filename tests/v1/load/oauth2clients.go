package main

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	fakemodels "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1/fake"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	"github.com/pquerna/otp/totp"
)

// fetchRandomOAuth2Client retrieves a random client from the list of available clients.
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

func mustBuildCode(totpSecret string) string {
	code, err := totp.GenerateCode(totpSecret, time.Now().UTC())
	if err != nil {
		panic(err)
	}
	return code
}

func buildOAuth2ClientActions(c *client.V1Client) map[string]*Action {
	return map[string]*Action{
		"CreateOAuth2Client": {
			Name: "CreateOAuth2Client",
			Action: func() (*http.Request, error) {
				ctx := context.Background()
				ui := fakemodels.BuildFakeUserCreationInput()
				u, err := c.CreateUser(ctx, ui)
				if err != nil {
					return nil, err
				}

				twoFactorSecret, err := testutil.ParseTwoFactorSecretFromBase64EncodedQRCode(u.TwoFactorQRCode)
				if err != nil {
					return nil, err
				}

				uli := &models.UserLoginInput{
					Username:  ui.Username,
					Password:  ui.Password,
					TOTPToken: mustBuildCode(twoFactorSecret),
				}

				cookie, err := c.Login(ctx, uli)
				if err != nil {
					return nil, err
				}

				req, err := c.BuildCreateOAuth2ClientRequest(
					ctx,
					cookie,
					&models.OAuth2ClientCreationInput{
						UserLoginInput: *uli,
					},
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
		"GetOAuth2ClientsForUser": {
			Name: "GetOAuth2ClientsForUser",
			Action: func() (*http.Request, error) {
				return c.BuildGetOAuth2ClientsRequest(context.Background(), nil)
			},
			Weight: 100,
		},
	}
}
