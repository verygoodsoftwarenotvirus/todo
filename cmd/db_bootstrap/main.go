package main

import (
	"log"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/postgres"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/queriers/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1/zerolog"
	"gitlab.com/verygoodsoftwarenotvirus/todo/tests/v1/db_bootstrap"
)

const (
	sqliteConnectionDetails   = "example.db"
	postgresConnectionDetails = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"
)

func main() {
	logger := zerolog.ProvideLogger(zerolog.ProvideZerologger())

	var (
		db  database.Database
		err error
	)

	switch strings.ToLower(os.Getenv("DATABASE_TO_USE")) {
	case "postgres":
		db, err = postgres.ProvidePostgres(false, logger, postgres.ConnectionDetails(postgresConnectionDetails))
	default:
		dbPath := sqliteConnectionDetails
		if len(os.Args) > 1 {
			dbPath = os.Args[1]
			log.Printf("set alternative output path: %q\n", dbPath)
		}
		db, err = sqlite.ProvideSqlite(false, logger, sqlite.Filepath(dbPath))
	}

	if err != nil {
		log.Fatalf("error opening database connection: %v\n", err)
	}

	if err = bootstrap.PreloadDatabase(db, logger); err != nil {
		log.Fatal("error preloading the database: ", err)
	}
}
