package main

import (
	"fmt"
	"os"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1/sqlite"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

const (
	ExpectedUsername = "username"
	ExpectedPassword = "password"

	defaultDBPath       = "example.db"
	defaultSchemaDir    = "database/v1/sqlite/schema"
	defaultSecret       = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultClientID     = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultClientSecret = defaultSecret
)

func main() {
	logger = logrus.New()
	// logger.SetLevel(logrus.DebugLevel)

	dbPath := defaultDBPath
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
		logger.Printf("set alternative output path: %q", dbPath)
	}

	dbcfg := database.Config{
		Logger: logger,
		// Debug:            true,
		ConnectionString: dbPath,
	}
	db, err := sqlite.NewSqlite(dbcfg)
	if err != nil {
		logger.Fatalf("error opening sqlite connection: %v", err)
	}

	if err := db.Migrate(defaultSchemaDir); err != nil {
		logger.Fatalf("error performing migration: %v", err)
	}

	b := auth.NewBcrypt(nil)
	hp, err := b.HashPassword(ExpectedPassword)
	if err != nil {
		logger.Fatalf("error hashing password: %v", err)
	}

	u, err := db.CreateUser(&models.UserInput{Username: ExpectedUsername, Password: hp}, defaultSecret)
	if err != nil {
		logger.Fatalf("error creating user: %v", err)
	} else if u.TwoFactorSecret != defaultSecret {
		logger.Fatal("wtf")
	}

	oac, err := db.CreateOauth2Client(
		&models.Oauth2ClientInput{
			UserLoginInput: models.UserLoginInput{Username: u.Username},
			Scopes:         []string{"*"},
		},
	)
	if err != nil {
		logger.Fatalf("error creating oauth client: %v", err)
	}

	reverseSecret := []rune(defaultSecret)
	for i, j := 0, len(reverseSecret)-1; i < j; i, j = i+1, j-1 {
		reverseSecret[i], reverseSecret[j] = reverseSecret[j], reverseSecret[i]
	}

	oac.ClientID, oac.ClientSecret = defaultClientID, defaultClientSecret
	if err := db.UpdateOauth2Client(oac); err != nil {
		logger.Fatalf("error overriding oauth client secrets: %v", err)
	}

	for i := 1; i < 6; i++ {
		exampleItem := &models.ItemInput{
			Name:    fmt.Sprintf("example item #%d", i),
			Details: fmt.Sprintf("example details #%d", i),
		}

		_, err := db.CreateItem(exampleItem)
		if err != nil {
			logger.Fatalf("error creating item #%d", i)
		}
	}
}
