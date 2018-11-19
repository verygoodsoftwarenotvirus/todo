package main

import (
	"fmt"
	"log"
	"os"

	// "gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models"
	//
	// "github.com/sirupsen/logrus"
)

const (
	defaultDBPath    = "example.db"
	defaultSchemaDir = "database/sqlite/schema"
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

	// a := auth.NewBcrypt(logrus.New())
	// password, err := a.HashPassword("password")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// ak := &models.AuthToken{
	// 	AppName:   appName,
	// 	Secret:    defaultSecret,
	// 	ExpiresOn: time.Now().Add(365 * (24 * time.Hour)),
	// }
	// if err := db.CreateAuthKey(ak); err != nil {
	// 	log.Fatal(err)
	// }

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
