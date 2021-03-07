package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types/fakes"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	flag "github.com/spf13/pflag"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

var (
	uri            string
	count          int
	debug          bool
	singleUserMode bool

	singleUser *types.User

	quitter = fatalQuitter{}
)

func init() {
	flag.StringVarP(&uri, "url", "u", "", "where the target instance is hosted")
	flag.IntVarP(&count, "count", "c", -1, "how many users/items per user to create")
	flag.BoolVarP(&debug, "debug", "d", false, "whether or not debug mode is enabled")
	flag.BoolVarP(&singleUserMode, "single-user-mode", "s", false, "whether or not single user mode is enabled")
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
	logger := zerolog.NewLogger()

	if debug {
		logger.SetLevel(logging.DebugLevel)
	}

	if count <= 0 {
		logger.Debug("exiting early because the requested amount is already satisfied")
		quitter.Quit(0)
	}

	if count == 1 && !singleUserMode {
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

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(x int, wg *sync.WaitGroup) {
			// create user.
			createdUser, userCreationErr := utils.CreateServiceUser(ctx, uri, "")
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

			cookie, cookieErr := utils.GetLoginCookie(ctx, uri, createdUser)
			if cookieErr != nil {
				quitter.ComplainAndQuit(fmt.Errorf("getting cookie: %v", cookieErr))
			}

			userClient, err := httpclient.NewClient(
				httpclient.UsingURL(parsedURI),
				httpclient.UsingLogger(userLogger),
				httpclient.UsingCookie(cookie),
			)
			if err != nil {
				quitter.ComplainAndQuit(fmt.Errorf("error initializing client: %w", err))
			}

			userLogger.Debug("assigned user API client")

			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				for j := 0; j < count; j++ {
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
				for j := 0; j < count; j++ {
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
authentication:  %s
2FA token: %s

`, singleUser.Username, singleUser.HashedPassword, buildTOTPTokenForSecret(singleUser.TwoFactorSecret))
		}
	}
}
