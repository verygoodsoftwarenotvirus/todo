package main

import (
	"context"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"net/url"

	flag "github.com/spf13/pflag"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

var (
	uri   string
	count int
	debug bool

	quitter = fatalQuitter{}
)

func init() {
	flag.StringVarP(&uri, "url", "u", "", "where the target instance is hosted")
	flag.IntVarP(&count, "count", "c", -1, "how many users/items per user to create")
	flag.BoolVarP(&debug, "debug", "d", false, "whether or not debug mode is enabled")
}

func main() {
	flag.Parse()

	ctx := context.Background()
	logger := zerolog.NewLogger()

	if debug {
		logger.SetLevel(logging.DebugLevel)
	}

	if uri == "" {
		quitter.ComplainAndQuit("uri must be valid")
	}

	parsedURI, err := url.Parse(uri)
	if err != nil {
		quitter.ComplainAndQuit(fmt.Errorf("error parsing provided URL: %w", err))
	}
	if parsedURI.Scheme == "" {
		quitter.ComplainAndQuit("provided URI missing scheme")
	}

	if count <= 0 {
		logger.Debug("exiting early because the requested amount is already satisfied")
		quitter.Quit(0)
	}

	var (
		userClient *httpclient.V1Client
	)

	for i := 0; i < count; i++ {
		// create user.
		createdUser, err := testutil.CreateServiceUser(ctx, uri, "", debug)
		if err != nil {
			quitter.ComplainAndQuit(fmt.Errorf("error creating user #%d: %w", i, err))
		}

		userLogger := logger.
			WithValue("username", createdUser.Username).
			WithValue("user_id", createdUser.ID).
			WithValue("user_number", i)

		userLogger.Debug("created user")

		for j := 0; j < count; j++ {
			iterationLogger := userLogger.WithValue("iteration", j)

			createdOAuth2Client, err := testutil.CreateObligatoryOAuth2Client(ctx, uri, createdUser)
			if err != nil {
				quitter.ComplainAndQuit(fmt.Errorf("error creating oauth2 client #%d for user #%d: %w", j, i, err))
			}
			iterationLogger.WithValue("client_id", createdOAuth2Client.ClientID).Debug("created oauth2 client")

			if j == 0 {
				userClient, err = httpclient.NewClient(
					ctx,
					createdOAuth2Client.ClientID,
					createdOAuth2Client.ClientSecret,
					parsedURI,
					iterationLogger,
					nil,
					[]string{"*"},
					debug,
				)
				if err != nil {
					quitter.ComplainAndQuit(fmt.Errorf("error initializing API client for user #%d: %w", i, err))
				}
				iterationLogger.Debug("assigned user API client")
			}

			createdWebhook, err := userClient.CreateWebhook(ctx, fakes.BuildFakeWebhookCreationInput())
			if err != nil {
				quitter.ComplainAndQuit(fmt.Errorf("error creating webhook #%d: %w", j, err))
			}
			iterationLogger.WithValue("webhook_id", createdWebhook.ID).Debug("created webhook")
		}

		for j := 0; j < count; j++ {
			iterationLogger := userLogger.WithValue("iteration", j)

			// create item
			createdItem, err := userClient.CreateItem(ctx, fakes.BuildFakeItemCreationInput())
			if err != nil {
				quitter.ComplainAndQuit(fmt.Errorf("error creating item #%d: %w", j, err))
			}
			iterationLogger.WithValue("webhook_id", createdItem.ID).Debug("created item")
		}
	}
}
