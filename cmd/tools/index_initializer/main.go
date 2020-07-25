package main

import (
	"context"
	"log"
	"time"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/config"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/search/bleve"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	flag "github.com/spf13/pflag"
	"gitlab.com/verygoodsoftwarenotvirus/logging/v1/zerolog"
)

var (
	indexOutputPath string
	typeName        string

	dbConnectionDetails string
	databaseType        string

	deadline time.Duration

	validTypeNames = map[string]struct{}{
		"item": {},
	}

	validDatabaseTypes = map[string]struct{}{
		config.PostgresProviderKey: {},
		config.MariaDBProviderKey:  {},
		config.SqliteProviderKey:   {},
	}
)

const (
	outputPathVerboseFlagName   = "output"
	dbConnectionVerboseFlagName = "db_connection"
	dbTypeVerboseFlagName       = "db_type"
)

func init() {
	flag.StringVarP(&indexOutputPath, outputPathVerboseFlagName, "o", "", "output path for bleve index")
	flag.StringVarP(&typeName, "type", "t", "", "which type to create bleve index for")

	flag.StringVarP(&dbConnectionDetails, dbConnectionVerboseFlagName, "c", "", "connection string for the relevant database")
	flag.StringVarP(&databaseType, dbTypeVerboseFlagName, "b", "", "which type of database to connect to")

	flag.DurationVarP(&deadline, "deadline", "d", time.Minute, "amount of time to spend adding to the index")
}

func main() {
	flag.Parse()
	logger := zerolog.NewZeroLogger().WithName("search_index_initializer")
	ctx := context.Background()

	if indexOutputPath == "" {
		log.Fatalf("No output path specified, please provide one via the --%s flag", outputPathVerboseFlagName)
		return
	} else if _, ok := validTypeNames[typeName]; !ok {
		log.Fatalf("Invalid type name %q specified, one of [ 'item' ] expected", typeName)
		return
	} else if dbConnectionDetails == "" {
		log.Fatalf("No database connection details %q specified, please provide one via the --%s flag", dbConnectionDetails, dbConnectionVerboseFlagName)
		return
	} else if _, ok := validDatabaseTypes[databaseType]; !ok {
		log.Fatalf("Invalid database type %q specified, please provide one via the --%s flag", databaseType, dbTypeVerboseFlagName)
		return
	}

	im, err := bleve.NewBleveIndexManager(search.IndexPath(indexOutputPath), search.IndexName(typeName), logger)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &config.ServerConfig{
		Database: config.DatabaseSettings{
			Provider:          databaseType,
			ConnectionDetails: database.ConnectionDetails(dbConnectionDetails),
		},
		Metrics: config.MetricsSettings{
			DBMetricsCollectionInterval: time.Second,
		},
	}

	// connect to our database.
	logger.Debug("connecting to database")
	rawDB, err := cfg.ProvideDatabaseConnection(logger)
	if err != nil {
		log.Fatalf("error establishing connection to database: %v", err)
	}

	// establish the database client.
	logger.Debug("setting up database client")
	dbClient, err := cfg.ProvideDatabaseClient(ctx, logger, rawDB)
	if err != nil {
		log.Fatalf("error initializing database client: %v", err)
	}

	switch typeName {
	case "item":
		outputChan := make(chan []models.Item)
		if queryErr := dbClient.GetAllItems(ctx, outputChan); queryErr != nil {
			log.Fatalf("error fetching items from database: %v", err)
		}

		for {
			select {
			case items := <-outputChan:
				for _, x := range items {
					if searchIndexErr := im.Index(ctx, x.ID, x); searchIndexErr != nil {
						logger.WithValue("id", x.ID).Error(searchIndexErr, "error adding to search index")
					}
				}
			case <-time.After(deadline):
				logger.Info("terminating")
				return
			}
		}
	default:
		log.Fatal("this should never occur")
	}
}