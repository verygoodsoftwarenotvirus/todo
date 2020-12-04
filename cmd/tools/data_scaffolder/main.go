package main

import (
	"context"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"net/url"
	"sync"

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

	if count <= 0 {
		logger.Debug("exiting early because the requested amount is already satisfied")
		quitter.Quit(0)
	}

	if uri == "" {
		quitter.ComplainAndQuit("uri must be valid")
	}

	parsedURI, uriParseErr := url.Parse(uri)
	if uriParseErr != nil {
		quitter.ComplainAndQuit(fmt.Errorf("error parsing provided URL: %w", uriParseErr))
	}
	if parsedURI.Scheme == "" {
		quitter.ComplainAndQuit("provided URI missing scheme")
	}

	wg := &sync.WaitGroup{}

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(x int, wg *sync.WaitGroup) {
			// create user.
			createdUser, userCreationErr := testutil.CreateServiceUser(ctx, uri, "", debug)
			if userCreationErr != nil {
				quitter.ComplainAndQuit(fmt.Errorf("error creating user #%d: %w", x, userCreationErr))
			}

			userLogger := logger.
				WithValue("username", createdUser.Username).
				WithValue("password", createdUser.HashedPassword).
				WithValue("totp_secret", createdUser.TwoFactorSecret).
				WithValue("user_id", createdUser.ID).
				WithValue("user_number", x)

			userLogger.Debug("created user")

			var (
				createdOAuth2Client     *types.OAuth2Client
				oauth2ClientCreationErr error
			)

			createdOAuth2Client, oauth2ClientCreationErr = testutil.CreateObligatoryOAuth2Client(ctx, uri, createdUser)
			if oauth2ClientCreationErr != nil {
				quitter.ComplainAndQuit(fmt.Errorf("error creating oauth2 client for user #%d: %w", x, oauth2ClientCreationErr))
			}
			userLogger.WithValue("client_id", createdOAuth2Client.ClientID).Debug("created oauth2 client")

			userClient, userClientInitErr := httpclient.NewClient(
				ctx,
				createdOAuth2Client.ClientID,
				createdOAuth2Client.ClientSecret,
				parsedURI,
				userLogger,
				nil,
				[]string{"*"},
				debug,
			)
			if userClientInitErr != nil {
				quitter.ComplainAndQuit(fmt.Errorf("error initializing API client for user #%d: %w", x, userClientInitErr))
			}
			userLogger.Debug("assigned user API client")

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < count-1; j++ {
					iterationLogger := userLogger.WithValue("iteration", j)

					createdOAuth2Client, oauth2ClientCreationErr = testutil.CreateObligatoryOAuth2Client(ctx, uri, createdUser)
					if oauth2ClientCreationErr != nil {
						quitter.ComplainAndQuit(fmt.Errorf("error creating oauth2 client #%d for user #%d: %w", j, x, oauth2ClientCreationErr))
					}

					iterationLogger.WithValue("client_id", createdOAuth2Client.ClientID).Debug("created oauth2 client")
				}
				wg.Done()
			}(wg)

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < count; j++ {
					iterationLogger := userLogger.WithValue("creating", "webhooks").WithValue("iteration", j)

					createdWebhook, webhookCreationErr := userClient.CreateWebhook(ctx, fakes.BuildFakeWebhookCreationInput())
					if webhookCreationErr != nil {
						quitter.ComplainAndQuit(fmt.Errorf("error creating webhook #%d: %w", j, webhookCreationErr))
					}
					iterationLogger.WithValue("webhook_id", createdWebhook.ID).Debug("created webhook")
				}
				wg.Done()
			}(wg)

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < count; j++ {
					iterationLogger := userLogger.WithValue("creating", "items").WithValue("iteration", j)

					// create item
					createdItem, itemCreationErr := userClient.CreateItem(ctx, fakes.BuildFakeItemCreationInput())
					if itemCreationErr != nil {
						quitter.ComplainAndQuit(fmt.Errorf("error creating item #%d: %w", j, itemCreationErr))
					}
					iterationLogger.WithValue("webhook_id", createdItem.ID).Debug("created item")
				}
				wg.Done()
			}(wg)

			wg.Done()
		}(i, wg)
	}

	wg.Wait()
}
