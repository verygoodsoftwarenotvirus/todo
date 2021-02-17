package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/httpclient"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
)

var (
	debug    bool
	urlToUse string
	cookie   *http.Cookie
)

func init() {
	ctx := context.Background()
	urlToUse = testutil.DetermineServiceURL()
	logger := zerolog.NewLogger()

	logger.WithValue(keys.URLKey, urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	u, err := testutil.CreateServiceUser(ctx, urlToUse, "", debug)
	if err != nil {
		logger.Fatal(err)
	}

	cookie, err = testutil.GetLoginCookie(ctx, urlToUse, u)
	if err != nil {
		logger.Fatal(err)
	}

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func initializeClient() *httpclient.Client {
	c := httpclient.NewClient(
		httpclient.UsingURI(urlToUse),
		httpclient.UsingLogger(zerolog.NewLogger()),
		httpclient.UsingCookie(cookie),
	)

	return c
}
