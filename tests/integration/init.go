package integration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

const (
	debug         = false
	timeout       = 5 * time.Second
	nonexistentID = 999999999
)

var (
	urlToUse string

	adminClientLock   sync.Mutex
	adminCookieClient *httpclient.Client
	adminPASETOClient *httpclient.Client

	premadeAdminUser = &types.User{
		ID:              1,
		TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		Username:        "exampleUser",
		HashedPassword:  "integration-tests-are-cool",
	}
)

func init() {
	ctx, span := tracing.StartSpan(context.Background())
	defer span.End()

	urlToUse = testutil.DetermineServiceURL()
	logger := zerolog.NewLogger()

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	adminCookie, err := utils.GetLoginCookie(ctx, urlToUse, premadeAdminUser)
	if err != nil {
		logger.Fatal(err)
	}

	adminCookieClient = initializeCookiePoweredClient(adminCookie)

	code, err := totp.GenerateCode(premadeAdminUser.TwoFactorSecret, time.Now().UTC())
	if err != nil {
		logger.Fatal(err)
	}

	apiClient, err := adminCookieClient.CreateAPIClient(ctx, adminCookie, &types.APICientCreationInput{
		Name: "admin_paseto_client",
		UserLoginInput: types.UserLoginInput{
			Username:  premadeAdminUser.Username,
			Password:  premadeAdminUser.HashedPassword,
			TOTPToken: code,
		},
	})
	if err != nil {
		logger.Fatal(err)
	}

	secretKey, err := base64.RawURLEncoding.DecodeString(apiClient.ClientSecret)
	if err != nil {
		logger.Fatal(err)
	}

	adminPASETOClient = initializePASETOPoweredClient(apiClient.ClientID, secretKey)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func buildHTTPClient() *http.Client {
	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   timeout,
	}
}

func initializeCookiePoweredClient(cookie *http.Cookie) *httpclient.Client {
	c := httpclient.NewClient(
		httpclient.UsingURI(urlToUse),
		httpclient.UsingLogger(zerolog.NewLogger()),
		httpclient.UsingHTTPClient(buildHTTPClient()),
		httpclient.UsingCookie(cookie),
	)

	if debug {
		if err := c.SetOption(httpclient.WithDebug()); err != nil {
			panic(err)
		}
	}

	return c
}
func initializePASETOPoweredClient(clientID string, secretKey []byte) *httpclient.Client {
	c := httpclient.NewClient(
		httpclient.UsingURI(urlToUse),
		httpclient.UsingLogger(zerolog.NewLogger()),
		httpclient.UsingHTTPClient(buildHTTPClient()),
		httpclient.UsingPASETO(clientID, secretKey),
	)

	if debug {
		if err := c.SetOption(httpclient.WithDebug()); err != nil {
			panic(err)
		}
	}

	return c
}

func buildSimpleClient() *httpclient.Client {
	return httpclient.NewClient(httpclient.UsingURI(urlToUse))
}
