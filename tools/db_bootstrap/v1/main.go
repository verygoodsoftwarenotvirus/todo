package main

import (
	"fmt"
	"log"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"
)

const (
	ExpectedUsername = "username"
	ExpectedPassword = "password"

	defaultDBPath    = "example.db"
	defaultSchemaDir = "database/v1/sqlite/schema"
	defaultSecret    = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
)

func main() {
	dbPath := defaultDBPath
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	dbcfg := database.Config{
		ConnectionString: dbPath,
	}
	db, err := sqlite.NewSqlite(dbcfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(defaultSchemaDir); err != nil {
		log.Fatal(err)
	}

	b := auth.NewBcrypt(nil)
	hp, _ := b.HashPassword(ExpectedPassword)

	db.CreateUser(&models.UserInput{
		Username:   ExpectedUsername,
		TOTPSecret: defaultSecret,
		Password:   hp,
	})

	for i := 1; i < 6; i++ {
		exampleItem := &models.ItemInput{
			Name:    fmt.Sprintf("example item #%d", i),
			Details: fmt.Sprintf("example details #%d", i),
		}

		_, err := db.CreateItem(exampleItem)
		if err != nil {
			log.Fatalf("error creating item #%d", i)
		}
	}
}
