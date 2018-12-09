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
		log.Printf("set alternative output path: %q", dbPath)
	}

	dbcfg := database.Config{ConnectionString: dbPath}
	db, err := sqlite.NewSqlite(dbcfg)
	if err != nil {
		log.Fatalf("error opening sqlite connection: %v", err)
	}

	if err := db.Migrate(defaultSchemaDir); err != nil {
		log.Fatalf("error performing migration: %v", err)
	}

	b := auth.NewBcrypt(nil)
	hp, err := b.HashPassword(ExpectedPassword)
	if err != nil {
		log.Fatalf("error hashing password: %v", err)
	}

	if _, err = db.CreateUser(&models.UserInput{
		Username:   ExpectedUsername,
		Password:   hp,
		TOTPSecret: defaultSecret,
	}); err != nil {
		log.Fatalf("error creating user: %v", err)
	}

	oac, err := db.CreateOauthClient(&models.OauthClientInput{Scopes: []string{"*"}})
	if err != nil {
		log.Fatalf("error creating user: %v", err)
	}

	log.Printf(`
	client_id: %q
client_secret: %q
`, oac.ClientID, oac.ClientSecret)

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
