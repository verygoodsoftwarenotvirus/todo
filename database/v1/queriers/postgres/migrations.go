package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gitlab.com/verygoodsoftwarenotvirus/todo/database/v1"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/auth"
	"gitlab.com/verygoodsoftwarenotvirus/todo/internal/v1/exampledata"

	"github.com/GuiaBolso/darwin"
	"github.com/Masterminds/squirrel"
)

var currentMigration float64 = 0

func incrementMigrationVersion() float64 {
	currentMigration++
	return currentMigration
}

var (
	migrations = []darwin.Migration{
		{
			Version:     incrementMigrationVersion(),
			Description: "create users table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"username" TEXT NOT NULL,
				"hashed_password" TEXT NOT NULL,
				"salt" BYTEA NOT NULL,
				"password_last_changed_on" integer,
				"requires_password_change" boolean NOT NULL DEFAULT 'false',
				"two_factor_secret" TEXT NOT NULL,
				"two_factor_secret_verified_on" BIGINT DEFAULT NULL,
				"is_admin" BOOLEAN NOT NULL DEFAULT 'false',
				"status" TEXT NOT NULL DEFAULT 'created',
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				UNIQUE ("username")
			);`,
		},
		{
			Version:     incrementMigrationVersion(),
			Description: "create sessions table for session manager",
			Script: `
			CREATE TABLE sessions (
				token TEXT PRIMARY KEY,
				data BYTEA NOT NULL,
				expiry TIMESTAMPTZ NOT NULL
			);

			CREATE INDEX sessions_expiry_idx ON sessions (expiry);
		`,
		},
		{
			Version:     incrementMigrationVersion(),
			Description: "create oauth2_clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth2_clients (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" TEXT DEFAULT '',
				"client_id" TEXT NOT NULL,
				"client_secret" TEXT NOT NULL,
				"redirect_uri" TEXT DEFAULT '',
				"scopes" TEXT NOT NULL,
				"implicit_allowed" boolean NOT NULL DEFAULT 'false',
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				"belongs_to_user" BIGINT NOT NULL,
				FOREIGN KEY("belongs_to_user") REFERENCES users(id)
			);`,
		},
		{
			Version:     incrementMigrationVersion(),
			Description: "create webhooks table",
			Script: `
			CREATE TABLE IF NOT EXISTS webhooks (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" TEXT NOT NULL,
				"content_type" TEXT NOT NULL,
				"url" TEXT NOT NULL,
				"method" TEXT NOT NULL,
				"events" TEXT NOT NULL,
				"data_types" TEXT NOT NULL,
				"topics" TEXT NOT NULL,
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				"belongs_to_user" BIGINT NOT NULL,
				FOREIGN KEY("belongs_to_user") REFERENCES users(id)
			);`,
		},
		{
			Version:     incrementMigrationVersion(),
			Description: "create audit log table",
			Script: `
			CREATE TABLE IF NOT EXISTS audit_log (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"event_type" TEXT NOT NULL,
				"event_data" JSONB NOT NULL,
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW())
			);`,
		},
		{
			Version:     incrementMigrationVersion(),
			Description: "create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" BIGSERIAL NOT NULL PRIMARY KEY,
				"name" CHARACTER VARYING NOT NULL,
				"details" CHARACTER VARYING NOT NULL DEFAULT '',
				"created_on" BIGINT NOT NULL DEFAULT extract(epoch FROM NOW()),
				"last_updated_on" BIGINT DEFAULT NULL,
				"archived_on" BIGINT DEFAULT NULL,
				"belongs_to_user" BIGINT NOT NULL,
				FOREIGN KEY("belongs_to_user") REFERENCES users(id)
			);`,
		},
	}
)

// buildMigrationFunc returns a sync.Once compatible function closure that will
// migrate a postgres database.
func buildMigrationFunc(db *sql.DB) func() {
	return func() {
		driver := darwin.NewGenericDriver(db, darwin.PostgresDialect{})
		if err := darwin.New(driver, migrations, nil).Migrate(); err != nil {
			panic(fmt.Errorf("error migrating database: %w", err))
		}
	}
}

// Migrate migrates the database. It does so by invoking the migrateOnce function via sync.Once, so it should be
// safe (as in idempotent, though not necessarily recommended) to call this function multiple times.
func (p *Postgres) Migrate(ctx context.Context, authenticator auth.Authenticator, testUserConfig *database.UserCreationConfig) error {
	p.logger.Info("migrating db")
	if !p.IsReady(ctx) {
		return errors.New("db is not ready yet")
	}

	p.migrateOnce.Do(buildMigrationFunc(p.db))

	const usingDemoCodeThatShouldBeDeletedLater = true
	if testUserConfig != nil && usingDemoCodeThatShouldBeDeletedLater {
		for _, x := range exampledata.ExampleUsers {
			query, args, err := p.sqlBuilder.
				Insert(usersTableName).
				Columns(
					usersTableUsernameColumn,
					usersTableHashedPasswordColumn,
					usersTableSaltColumn,
					usersTableTwoFactorColumn,
					usersTableIsAdminColumn,
					usersTableTwoFactorVerifiedOnColumn,
				).
				Values(
					x.Username,
					x.HashedPassword,
					x.Salt,
					x.TwoFactorSecret,
					x.IsAdmin,
					squirrel.Expr(currentUnixTimeQuery),
				).
				ToSql()
			p.logQueryBuildingError(err)

			if _, dbErr := p.db.ExecContext(ctx, query, args...); dbErr != nil {
				return dbErr
			}
		}

		for _, x := range exampledata.ExampleItems {
			for _, y := range x {
				if _, dbErr := p.CreateItem(ctx, y); dbErr != nil {
					return dbErr
				}
			}
		}

		for _, x := range exampledata.ExampleOAuth2Clients {
			query, args := p.buildCreateOAuth2ClientQuery(x)
			if _, dbErr := p.db.ExecContext(ctx, query, args...); dbErr != nil {
				return dbErr
			}
		}

		for _, x := range exampledata.ExampleWebhooks {
			query, args := p.buildCreateWebhookQuery(x)
			if _, dbErr := p.db.ExecContext(ctx, query, args...); dbErr != nil {
				return dbErr
			}
		}
	}

	if testUserConfig != nil {
		hashedPassword, err := authenticator.HashPassword(ctx, testUserConfig.Password)
		if err != nil {
			return fmt.Errorf("error hashing test user password: %w", err)
		}

		query, args, err := p.sqlBuilder.
			Insert(usersTableName).
			Columns(
				usersTableUsernameColumn,
				usersTableHashedPasswordColumn,
				usersTableSaltColumn,
				usersTableTwoFactorColumn,
				usersTableIsAdminColumn,
				usersTableTwoFactorVerifiedOnColumn,
			).
			Values(
				testUserConfig.Username,
				hashedPassword,
				[]byte("aaaaaaaaaaaaaaaa"),
				// `otpauth://totp/todo:username?secret=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=&issuer=todo`
				"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
				testUserConfig.IsAdmin,
				squirrel.Expr(currentUnixTimeQuery),
			).
			ToSql()
		p.logQueryBuildingError(err)

		if _, dbErr := p.db.ExecContext(ctx, query, args...); dbErr != nil {
			p.logger.Error(err, "creating user")
			return fmt.Errorf("error creating test user: %w", dbErr)
		}

		p.logger.WithValue("username", testUserConfig.Username).Debug("created user")
	}

	return nil
}
