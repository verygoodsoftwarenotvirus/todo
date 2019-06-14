package integration

import (
	"context"
	"testing"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opencensus.io/trace"
)

func TestExport(test *testing.T) {
	test.Parallel()

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
					PasswordLastChangedOn: x.PasswordLastChangedOn,
					IsAdmin:               x.IsAdmin,
					CreatedOn:             x.CreatedOn,
					UpdatedOn:             x.UpdatedOn,
					ArchivedOn:            x.ArchivedOn,
				},
				Items:         []models.Item{*item},
				Webhooks:      []models.Webhook{*webhook},
				OAuth2Clients: []models.OAuth2Client{*premade, *oac},
			}

			actual, err := c.ExportData(ctx)
			checkValueAndError(t, actual, err)

			require.Equal(t, len(expected.Items), len(actual.Items))
			for i := range actual.Items {
				e := &expected.Items[i]
				a := &actual.Items[i]
				checkItemEquality(t, e, a)
			}

			require.Equal(t, len(expected.Webhooks), len(actual.Webhooks))
			for i := range actual.Webhooks {
				e := &expected.Webhooks[i]
				a := &actual.Webhooks[i]
				checkWebhookEquality(t, e, a)
			}

			require.Equal(t, len(expected.OAuth2Clients), len(actual.OAuth2Clients))
			for i := range actual.OAuth2Clients {
				e := &expected.OAuth2Clients[i]
				a := &actual.OAuth2Clients[i]
				checkOAuth2ClientEquality(t, e, a)
			}

			assert.Equal(t, expected, actual)
		})
	})
}
