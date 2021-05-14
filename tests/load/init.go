package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	zerolog "gitlab.com/verygoodsoftwarenotvirus/todo/internal/observability/logging/zerolog"
	httpclient2 "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"

	testutil "gitlab.com/verygoodsoftwarenotvirus/todo/tests/utils"
)

var (
	urlToUse string
)

func init() {
	ctx := context.Background()
	urlToUse = testutil.DetermineServiceURL().String()
	logger := zerolog.NewLogger()

	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(ctx, urlToUse)

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func initializeClient() *httpclient2.Client {
	uri, err := url.Parse(urlToUse)
	if err != nil {
		panic(err)
	}

	c, err := httpclient2.NewClient(uri)
	if err != nil {
		panic(err)
	}
	return c
}
