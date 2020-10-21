package integration

import (
	"context"
	"fmt"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/tracing"
	"net/http"
	"net/url"
	"strings"
	"time"

	client "gitlab.com/verygoodsoftwarenotvirus/todo/client/v1/http"
	models "gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/testutil"

	"gitlab.com/verygoodsoftwarenotvirus/logging/v2/zerolog"
)

const (
	debug         = true
	nonexistentID = 999999999
)

var (
	urlToUse   string
	todoClient *client.V1Client
)

func init() {
	ctx, span := tracing.StartSpan(context.Background(), "init")
	defer span.End()

	urlToUse = testutil.DetermineServiceURL()
	logger := zerolog.NewLogger()

	logger.WithValue("url", urlToUse).Info("checking server")
	testutil.EnsureServerIsUp(urlToUse)

	ogUser, err := testutil.CreateObligatoryUser(urlToUse, debug)
	if err != nil {
		logger.Fatal(err)
	}

	// make the user an admin
	dbURL, dbVendor := testutil.DetermineDatabaseURL()
	tempCfg := config.ServerConfig{
		Metrics: config.MetricsSettings{
			DBMetricsCollectionInterval: config.DefaultDatabaseMetricsCollectionInterval,
		},
		Database: config.DatabaseSettings{
			Provider:          dbVendor,
			ConnectionDetails: database.ConnectionDetails(dbURL),
			RunMigrations:     false,
		},
	}

	dbConn, dbConnectionErr := tempCfg.ProvideDatabaseConnection(logger)
	if dbConnectionErr != nil {
		logger.Fatal(dbConnectionErr)
	}

	db, dbClientInitErr := tempCfg.ProvideDatabaseClient(ctx, logger, dbConn)
	if dbClientInitErr != nil {
		logger.Fatal(dbClientInitErr)
	}

	if makeAdminErr := db.MakeUserAdmin(ctx, ogUser.ID); makeAdminErr != nil {
		logger.Fatal(makeAdminErr)
	}

	oa2Client, err := testutil.CreateObligatoryClient(urlToUse, ogUser)
	if err != nil {
		logger.Fatal(err)
	}

	todoClient = initializeClient(oa2Client)
	todoClient.Debug = urlToUse == "" // change this for debug logs

	fiftySpaces := strings.Repeat("\n", 50)
	fmt.Printf("%s\tRunning tests%s", fiftySpaces, fiftySpaces)
}

func buildHTTPClient() *http.Client {
	return &http.Client{
		Transport: http.DefaultTransport,
		Timeout:   5 * time.Second,
	}
}

func initializeClient(oa2Client *models.OAuth2Client) *client.V1Client {
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
