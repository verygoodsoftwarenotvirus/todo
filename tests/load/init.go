package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/logging/zerolog"
	client "gitlab.com/verygoodsoftwarenotvirus/todo/pkg/client/httpclient"
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

func buildHTTPClient() *http.Client {
	httpc := &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}

	return httpc
}

func initializeClient() *client.Client {
	uri, err := url.Parse(urlToUse)
	if err != nil {
		panic(err)
	}

	c, err := client.NewClient(uri)
	if err != nil {
		panic(err)
	}
	return c
}
