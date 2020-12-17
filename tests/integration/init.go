package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"github.com/stretchr/testify/require"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/noop"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

const (
	debug         = false
	timeout       = 5 * time.Second
	nonexistentID = 999999999

	clientsDebug = debug
)

var (
	urlToUse    string
	adminClient *httpclient.V1Client

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

	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	adminOAuth2Client, err := testutil.CreateObligatoryOAuth2Client(ctx, urlToUse, premadeAdminUser)
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

func initializeClient(oa2Client *types.OAuth2Client) *httpclient.V1Client {
	uri, err := url.Parse(urlToUse)
	if err != nil {
		panic(err)
	}

	logger := zerolog.NewLogger()
	if !debug {
		logger = noop.NewLogger()
	}

	c, err := httpclient.NewClient(
		context.Background(),
		oa2Client.ClientID,
		oa2Client.ClientSecret,
		uri,
		logger,
		buildHTTPClient(),
		oa2Client.Scopes,
		debug,
	)
	if err != nil {
		panic(err)
	}

	return c
}

func buildSimpleClient(ctx context.Context, t *testing.T) *httpclient.V1Client {
	t.Helper()

	uri, err := url.Parse(urlToUse)
	require.NoError(t, err)

	c, err := httpclient.NewSimpleClient(ctx, uri, false)
	require.NoError(t, err)

	return c
}
