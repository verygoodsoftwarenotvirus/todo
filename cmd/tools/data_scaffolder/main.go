package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/types/fakes"
	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"github.com/pquerna/otp/totp"
	flag "github.com/spf13/pflag"
)

var (
	uri            string
	userCount      uint16
	dataCount      uint16
	debug          bool
	singleUserMode bool

	singleUser *types.User

	quitter = fatalQuitter{}
)

func init() {
	flag.StringVarP(&uri, "url", "u", "", "where the target instance is hosted")
	flag.Uint16VarP(&userCount, "user-count", "c", 0, "how many users to create")
	flag.Uint16VarP(&dataCount, "data-count", "d", 0, "how many accounts/api clients/etc per user to create")
	flag.BoolVarP(&debug, "debug", "z", false, "whether debug mode is enabled")
	flag.BoolVarP(&singleUserMode, "single-user-mode", "s", false, "whether single user mode is enabled")
}

func clearTheScreen() {
	fmt.Println("\x1b[2J")
	fmt.Printf("\x1b[0;0H")
}

func buildTOTPTokenForSecret(secret string) string {
	secret = strings.ToUpper(secret)
	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		panic(err)
	}

	if !totp.Validate(code, secret) {
		panic("this shouldn't happen")
	}

	return code
}

func main() {
	flag.Parse()

	ctx := context.Background()
	logger := logging.ProvideLogger(logging.Config{Provider: logging.ProviderZerolog})

	if debug {
		logger.SetLevel(logging.DebugLevel)
	}

	if dataCount <= 0 {
		logger.Debug("exiting early because the requested amount is already satisfied")
		quitter.Quit(0)
	}

	if dataCount == 1 && !singleUserMode {
		singleUserMode = true
	}

	if uri == "" {
		quitter.ComplainAndQuit("uri must be valid")
	}

	parsedURI, uriParseErr := url.Parse(uri)
	if uriParseErr != nil {
		quitter.ComplainAndQuit(fmt.Errorf("parsing provided url: %w", uriParseErr))
	}
	if parsedURI.Scheme == "" {
		quitter.ComplainAndQuit("provided URI missing scheme")
	}

	wg := &sync.WaitGroup{}

	for i := 0; i < int(userCount); i++ {
		wg.Add(1)
		go func(x int, wg *sync.WaitGroup) {
			createdUser, userCreationErr := testutil.CreateServiceUser(ctx, uri, "")
			if userCreationErr != nil {
				quitter.ComplainAndQuit(fmt.Errorf("creating user #%d: %w", x, userCreationErr))
			}

			if x == 0 && singleUserMode {
				singleUser = createdUser
			}

			userLogger := logger.
				WithValue("username", createdUser.Username).
				WithValue("password", createdUser.HashedPassword).
				WithValue("totp_secret", createdUser.TwoFactorSecret).
				WithValue("user_id", createdUser.ID).
				WithValue("user_number", x)

			userLogger.Debug("created user")

			cookie, cookieErr := testutil.GetLoginCookie(ctx, uri, createdUser)
			if cookieErr != nil {
				quitter.ComplainAndQuit(fmt.Errorf("getting cookie: %v", cookieErr))
			}

			userClient, err := httpclient.NewClient(parsedURI, httpclient.UsingLogger(userLogger), httpclient.UsingCookie(cookie))
			if err != nil {
				quitter.ComplainAndQuit(fmt.Errorf("initializing client: %w", err))
			}

			userLogger.Debug("assigned user API client")

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < int(dataCount); j++ {
					iterationLogger := userLogger.WithValue("creating", "accounts").WithValue("iteration", j)

					createdAccount, accountCreationError := userClient.CreateAccount(ctx, fakes.BuildFakeAccountCreationInput())
					if accountCreationError != nil {
						quitter.ComplainAndQuit(fmt.Errorf("creating account #%d: %w", j, accountCreationError))
					}

					iterationLogger.WithValue(keys.AccountIDKey, createdAccount.ID).Debug("created account")
				}
				wg.Done()
			}(wg)

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < int(dataCount); j++ {
					iterationLogger := userLogger.WithValue("creating", "api_clients").WithValue("iteration", j)

					code, codeErr := totp.GenerateCode(strings.ToUpper(createdUser.TwoFactorSecret), time.Now().UTC())
					if codeErr != nil {
						quitter.ComplainAndQuit(fmt.Errorf("creating API Client #%d: %w", j, codeErr))
					}

					fakeInput := fakes.BuildFakeAPIClientCreationInput()

					createdAPIClient, apiClientCreationErr := userClient.CreateAPIClient(ctx, cookie, &types.APIClientCreationInput{
						UserLoginInput: types.UserLoginInput{
							Username:  createdUser.Username,
							Password:  createdUser.HashedPassword,
							TOTPToken: code,
						},
						Name: fakeInput.Name,
					})
					if apiClientCreationErr != nil {
						quitter.ComplainAndQuit(fmt.Errorf("API Client webhook #%d: %w", j, apiClientCreationErr))
					}

					iterationLogger.WithValue(keys.APIClientDatabaseIDKey, createdAPIClient.ID).Debug("created API Client")
				}
				wg.Done()
			}(wg)

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < int(dataCount); j++ {
					iterationLogger := userLogger.WithValue("creating", "webhooks").WithValue("iteration", j)

					createdWebhook, webhookCreationErr := userClient.CreateWebhook(ctx, fakes.BuildFakeWebhookCreationInput())
					if webhookCreationErr != nil {
						quitter.ComplainAndQuit(fmt.Errorf("creating webhook #%d: %w", j, webhookCreationErr))
					}

					iterationLogger.WithValue(keys.WebhookIDKey, createdWebhook.ID).Debug("created webhook")
				}
				wg.Done()
			}(wg)

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < int(dataCount); j++ {
					iterationLogger := userLogger.WithValue("creating", "items").WithValue("iteration", j)

					// create item
					createdItem, itemCreationErr := userClient.CreateItem(ctx, fakes.BuildFakeItemCreationInput())
					if itemCreationErr != nil {
						quitter.ComplainAndQuit(fmt.Errorf("creating item #%d: %w", j, itemCreationErr))
					}

					iterationLogger.WithValue(keys.WebhookIDKey, createdItem.ID).Debug("created item")
				}
				wg.Done()
			}(wg)

			wg.Done()
		}(i, wg)
	}

	wg.Wait()

	if singleUserMode && singleUser != nil {
		logger.Debug("engage single user mode!")

		for range time.Tick(1 * time.Second) {
			clearTheScreen()
			fmt.Printf(`

username:  %s
passwords:  %s
2FA token: %s

`, singleUser.Username, singleUser.HashedPassword, buildTOTPTokenForSecret(singleUser.TwoFactorSecret))
		}
	}
}
