package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"gitlab.com/verygoodsoftwarenotvirus/todo/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/lib/logging/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/models/v1"

	"github.com/opentracing/opentracing-go"
)

const (
	expectedUsername = "username"
	expectedPassword = "password"

	sqliteSchemaDir         = "database/v1/sqlite/schema"
	sqliteConnectionDetails = "example.db"

	postgresSchemaDir         = "database/v1/postgres/schema"
	postgresConnectionDetails = "postgres://todo:hunter2@database:5432/todo?sslmode=disable"

	localTestInstanceURL = "http://localhost"
	defaultSecret        = "HEREISASECRETWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultClientID      = "HEREISACLIENTIDWHICHIVEMADEUPBECAUSEIWANNATESTRELIABLY"
	defaultClientSecret  = defaultSecret
)

// PreloadDatabase migrates a postgres database
func PreloadDatabase(
	db database.Database,
	logger logging.Logger,
	tracer opentracing.Tracer,
) error {
	switch strings.ToLower(os.Getenv("DATABASE_TO_USE")) {
	case "postgres":
		println()
	default:
		println()
	}
	ctx := context.Background()

	if !db.IsReady(context.Background()) {
		return errors.New("no database ready")
	}

	if err := db.Migrate(ctx); err != nil {
		return err
	}

	b := auth.ProvideBcrypt(auth.DefaultBcryptHashCost, logger, tracer)
	hp, err := b.HashPassword(ctx, expectedUsername)
	if err != nil {
		return err
	}

	u, err := db.CreateUser(ctx, &models.UserInput{
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
		ctx,
		&models.OAuth2ClientCreationInput{
			UserLoginInput: models.UserLoginInput{Username: u.Username},
			Scopes:         []string{"*"},
			BelongsTo:      u.ID,
		},
	)
	if err != nil {
		logger.Error(err, "error creating oauth client")
		logger.Fatal(err)
	}

	oac.ClientID, oac.ClientSecret, oac.RedirectURI = defaultClientID, defaultClientSecret, localTestInstanceURL
	if err := db.UpdateOAuth2Client(ctx, oac); err != nil {
		return err
	}

	for i := 1; i < 6; i++ {
		if _, err := db.CreateItem(ctx, &models.ItemInput{
			Name:      fmt.Sprintf("example item #%d", i),
			Details:   fmt.Sprintf("example details #%d", i),
			BelongsTo: u.ID,
		}); err != nil {
			return err
		}
	}

	return nil
}
