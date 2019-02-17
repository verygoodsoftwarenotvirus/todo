package main

import (
	"log"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/tracing/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/db_bootstrap"
)

const (
	sqliteConnectionDetails = "example.db"

	//postgresSchemaDir         = "database/v1/postgres/schema"
	postgresConnectionDetails = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"
)

func main() {
	logger := zerolog.ProvideLogger(zerolog.ProvideZerologger())

	var (
		db  database.Database
		err error
	)

	tracer, err := tracing.ProvideTracer("db-bootstrap")
	if err != nil {
		log.Printf("error building tracer: %v\n", err)
	}

	switch strings.ToLower(os.Getenv("DATABASE_TO_USE")) {
	case "postgres":
		db, err = postgres.ProvidePostgres(false, logger, tracer, database.ConnectionDetails(postgresConnectionDetails))
	default:
		dbPath := sqliteConnectionDetails
		if len(os.Args) > 1 {
			dbPath = os.Args[1]
			log.Printf("set alternative output path: %q\n", dbPath)
		}
		db, err = sqlite.ProvideSqlite(false, logger, tracer, database.ConnectionDetails(dbPath))
	}

	if err != nil {
		log.Fatalf("error opening database connection: %v\n", err)
	}

	if err = bootstrap.PreloadDatabase(
		db,
		logger,
		tracer,
	); err != nil {
		log.Fatal("error preloading the database: ", err)
	}

}
