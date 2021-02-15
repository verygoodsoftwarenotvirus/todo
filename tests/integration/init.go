package integration

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/tracing"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

const (
	debug                 = false
	timeout               = 5 * time.Second
	defaultSubtestTimeout = 15 * time.Second
	nonexistentID         = 999999999
)

var (
	urlToUse string

	adminClientLock sync.Mutex
	adminClient     *httpclient.Client

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

	adminCookie, err := testutil.GetLoginCookie(ctx, urlToUse, premadeAdminUser)
	if err != nil {
		logger.Fatal(err)
	}

	adminClient = initializeClient(adminCookie)
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

func initializeClient(cookie *http.Cookie) *httpclient.Client {
	uri := httpclient.MustParseURL(urlToUse)

	c := httpclient.NewClient(
		httpclient.WithURL(uri),
		httpclient.WithLogger(logging.NewNonOperationalLogger()),
		httpclient.WithHTTPClient(buildHTTPClient()),
		httpclient.WithCookieCredentials(cookie),
	)

	if debug {
		c.SetOption(httpclient.WithDebugEnabled())
	}

	return c
}

func buildSimpleClient() *httpclient.Client {
	return httpclient.NewClient(httpclient.WithURL(httpclient.MustParseURL(urlToUse)))
}
