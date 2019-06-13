package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/zerolog"
	"testing"

	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	// "github.com/stretchr/testify/assert"
	"go.opencensus.io/trace"
)

func TestExport(test *testing.T) {
	test.SkipNow()

	test.Run("Exporting", func(T *testing.T) {
		T.Run("should be exportable", func(t *testing.T) {
			tctx := context.Background()
			ctx, span := trace.StartSpan(tctx, t.Name())
			defer span.End()

			// create user
			x, y, cookie := buildDummyUser(test)
			assert.NotNil(test, cookie)

			clientInput := buildDummyOAuth2ClientInput(test, x.Username, y.Password, x.TwoFactorSecret)
			premade, err := todoClient.CreateOAuth2Client(tctx, cookie, clientInput)
			checkValueAndError(test, premade, err)

			c, err := client.NewClient(
				context.Background(),
				premade.ClientID,
				premade.ClientSecret,
				todoClient.URL,
				zerolog.NewZeroLogger(),
				buildHTTPClient(),
				premade.Scopes,
				true,
			)
			checkValueAndError(test, c, err)

			// Create item
			item, err := c.CreateItem(
				ctx,
				&models.ItemInput{
					Name:    "name",
					Details: "details",
				})
			checkValueAndError(t, item, err)

			// Create webhook
			webhook, err := c.CreateWebhook(
				ctx,
				&models.WebhookInput{
					Name: "name",
				})
			checkValueAndError(t, webhook, err)

			// Create OAuth2Client
			oac, err := c.CreateOAuth2Client(
				ctx,
				cookie,
				buildDummyOAuth2ClientInput(
					t,
					x.Username,
					y.Password,
					x.TwoFactorSecret,
				),
			)
			checkValueAndError(t, oac, err)

			expected := &models.DataExport{
				User: models.User{
					ID:                    x.ID,
					Username:              x.Username,
					TwoFactorSecret:       x.TwoFactorSecret,
					PasswordLastChangedOn: x.PasswordLastChangedOn,
					IsAdmin:               x.IsAdmin,
					CreatedOn:             x.CreatedOn,
					UpdatedOn:             x.UpdatedOn,
					ArchivedOn:            x.ArchivedOn,
				},
				Items:         []models.Item{*item},
				Webhooks:      []models.Webhook{*webhook},
				OAuth2Clients: []models.OAuth2Client{*oac},
			}

			actual, err := c.ExportData(ctx)
			checkValueAndError(t, actual, err)
			assert.Equal(t, expected, actual)
		})
	})
}
