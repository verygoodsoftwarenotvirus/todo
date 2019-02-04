package main

import (
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/integration/v1/db_bootstrap"

	"github.com/sirupsen/logrus"
)

const (
	expectedUsername = "username"
	expectedPassword = "password"

	sqliteSchemaDir         = "database/v1/sqlite/schema"
	sqliteConnectionDetails = "example.db"

	postgresSchemaDir         = "database/v1/postgres/schema"
	postgresConnectionDetails = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"

	localTestInstanceURL = "https://localhost"
	defaultSecret        = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultClientID      = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultClientSecret  = defaultSecret
)

func main() {
	logger := logrus.New()
	// logger.SetLevel(logrus.DebugLevel)

	var (
		db        database.Database
		schemaDir string
		err       error
	)

	tracer, err := tracing.ProvideTracer("db-bootstrap")
	if err != nil {
		logger.Debugf("error building tracer: %v\n", err)
	}

	switch strings.ToLower(os.Getenv("DATABASE_TO_USE")) {
	case "postgres":
		schemaDir = postgresSchemaDir
		db, err = postgres.ProvidePostgres(false, logger, tracer, database.ConnectionDetails(postgresConnectionDetails))
	default:
		schemaDir = sqliteSchemaDir
		dbPath := sqliteConnectionDetails
		if len(os.Args) > 1 {
			dbPath = os.Args[1]
			logger.Printf("set alternative output path: %q\n", dbPath)
		}
		db, err = sqlite.ProvideSqlite(false, logger, tracer, database.ConnectionDetails(dbPath))
	}
	if err != nil {
		logger.Fatalf("error opening database connection: %v\n", err)
	}

	bootstrap.PreloadDatabase(db, database.SchemaDirectory(schemaDir))

}
