package main

import (
	"context"
	"fmt"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/httpclient"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/observability/keys"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/testutil"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/pkg/types"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

var (
	debug     bool
	urlToUse  string
	oa2Client *types.OAuth2Client
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

	oa2Client, err = testutil.CreateObligatoryOAuth2Client(ctx, urlToUse, u)
	if err != nil {
		logger.Fatal(err)
	}

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func initializeClient(oa2Client *types.OAuth2Client) *httpclient.Client {
	uri := httpclient.MustParseURL(urlToUse)

	c := httpclient.NewClient(
		httpclient.WithURL(uri),
		httpclient.WithLogger(zerolog.NewLogger()),
		httpclient.WithOAuth2ClientCredentials(
			httpclient.BuildClientCredentialsConfig(
				uri,
				oa2Client.ClientID,
				oa2Client.ClientSecret,
				oa2Client.Scopes...,
			),
		),
	)

	return c
}
