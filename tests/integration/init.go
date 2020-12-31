package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

const (
	debug         = false
	timeout       = 5 * time.Second
	nonexistentID = 999999999
)

var (
	urlToUse string

	adminClientLock sync.Mutex
	adminClient     *httpclient.V1Client

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

	adminOAuth2Client, err := testutil.CreateObligatoryOAuth2Client(ctx, urlToUse, premadeAdminUser)
	if err != nil {
		logger.Fatal(err)
	}

	adminClient = initializeClient(adminOAuth2Client)
	adminClient.SetOption(httpclient.WithDebugEnabled())

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
	uri := httpclient.MustParseURL(urlToUse)

	c := httpclient.NewClient(
		httpclient.WithHTTPClient(buildHTTPClient()),
		httpclient.WithURL(uri),
		httpclient.WithOAuth2ClientCredentials(
			httpclient.BuildClientCredentialsConfig(
				uri,
				oa2Client.ClientID,
				oa2Client.ClientSecret,
				oa2Client.Scopes...,
			),
		),
	)

	if debug {
		c.SetOption(httpclient.WithDebugEnabled())
	} else {
		l := zerolog.NewLogger()
		if !debug {
			l.SetLevel(logging.InfoLevel)
		}
		c.SetOption(httpclient.WithLogger(l))
	}

	return c
}

func buildSimpleClient() *httpclient.V1Client {
	return httpclient.NewClient(httpclient.WithURL(httpclient.MustParseURL(urlToUse)))
}
