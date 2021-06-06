package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging"
	"gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
	testutils "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

var (
	urlToUse string
)

func init() {
	ctx := context.Background()
	urlToUse = testutils.DetermineServiceURL().String()
	logger := logging.ProvideLogger(logging.Config{Provider: logging.ProviderZerolog})

	logger.WithValue("url", urlToUse).Info("checking server")
	testutils.EnsureServerIsUp(ctx, urlToUse)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func initializeClient() *httpclient.Client {
	uri, err := url.Parse(urlToUse)
	if err != nil {
		panic(err)
	}

	c, err := httpclient.NewClient(uri)
	if err != nil {
		panic(err)
	}
	return c
}
