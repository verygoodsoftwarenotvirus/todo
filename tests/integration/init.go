package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/client/http"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

const (
	debug         = true
	timeout       = 5 * time.Second
	nonexistentID = 999999999
)

var (
	urlToUse    string
	adminClient *client.V1Client
	todoClient  *client.V1Client

	premadeAdminUser = &types.User{
		ID:              1,
		TwoFactorSecret: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		Username:        "exampleUser",
		HashedPassword:  "integration-tests-are-cool",
	}
)

func init() {
	ctx, span := tracing.StartSpan(context.Background(), "init")
	defer span.End()

	urlToUse = testutil.DetermineServiceURL()
	logger := zerolog.NewLogger()

	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	ogUser, err := testutil.CreateObligatoryUser(urlToUse, debug)
	if err != nil {
		logger.Fatal(err)
	}

	oa2Client, err := testutil.CreateObligatoryClient(ctx, urlToUse, ogUser)
	if err != nil {
		logger.Fatal(err)
	}

	clientsDebug := urlToUse == "" // change this to change debug log behavior
	todoClient = initializeClient(oa2Client)
	todoClient.Debug = clientsDebug

	adminOAuth2Client, err := testutil.CreateObligatoryClient(ctx, urlToUse, premadeAdminUser)
	if err != nil {
		logger.Fatal(err)
	}

	adminClient = initializeClient(adminOAuth2Client)
	adminClient.Debug = clientsDebug

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func buildHTTPClient() *http.Client {
	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   timeout,
	}
}

func initializeClient(oa2Client *types.OAuth2Client) *client.V1Client {
	uri, err := url.Parse(urlToUse)
	if err != nil {
		panic(err)
	}

	c, err := client.NewClient(
		context.Background(),
		oa2Client.ClientID,
		oa2Client.ClientSecret,
		uri,
		zerolog.NewLogger(),
		buildHTTPClient(),
		oa2Client.Scopes,
		debug,
	)
	if err != nil {
		panic(err)
	}
	return c
}
