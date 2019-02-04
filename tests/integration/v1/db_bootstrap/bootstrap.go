package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

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

// PreloadDatabase migrates a postgres database
func PreloadDatabase(db database.Database, schemaDir database.SchemaDirectory) error {
	switch strings.ToLower(os.Getenv("DATABASE_TO_USE")) {
	case "postgres":
		println()
	default:
		println()
	}

	if !db.IsReady(context.Background()) {
		return errors.New("no database ready")
	}

	if len(schemaDir) > 0 {
		if err := db.Migrate(context.Background(), schemaDir); err != nil {
			return err
		}
	}

	b := auth.NewBcrypt(nil)
	hp, err := b.HashPassword(expectedUsername)
	if err != nil {
		return err
	}

	u, err := db.CreateUser(context.Background(), &models.UserInput{
		Username:        expectedUsername,
		Password:        hp,
		IsAdmin:         true,
		TwoFactorSecret: defaultSecret,
	})
	if err != nil {
		return err
	} else if u.TwoFactorSecret != defaultSecret {
		return errors.New("wtf")
	}

	oac, err := db.CreateOAuth2Client(
		context.Background(),
		&models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{Username: u.Username},
			Scopes:         []string{"*"},
			BelongsTo:      u.ID,
		},
	)
	if err != nil {
		logger.Fatalf("error creating oauth client: %v", err)
	}

	oac.ClientID, oac.ClientSecret, oac.RedirectURI = defaultClientID, defaultClientSecret, localTestInstanceURL
	if err := db.UpdateOAuth2Client(context.Background(), oac); err != nil {
		return err
	}

	for i := 1; i < 6; i++ {
		if _, err := db.CreateItem(context.Background(), &models.ItemInput{
			Name:      fmt.Sprintf("example item #%d", i),
			Details:   fmt.Sprintf("example details #%d", i),
			BelongsTo: u.ID,
		}); err != nil {
			return err
		}
	}

	return nil
}
