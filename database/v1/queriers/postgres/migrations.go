// +build !migrations

package postgres

import (
	"context"
	// "database/sql"

	"github.com/GuiaBolso/darwin"
	"github.com/pkg/errors"
)

var (
	migrations = []darwin.Migration{
		{
			Version:     1,
			Description: "Create user table",
			Script: `
			CREATE TABLE IF NOT EXISTS users (
				"id" bigserial NOT NULL PRIMARY KEY,
				"username" text NOT NULL,
				"hashed_password" text NOT NULL,
				"password_last_changed_on" integer,
				"two_factor_secret" text NOT NULL,
				"is_admin" bool NOT NULL DEFAULT 'false',
				"created_on" bigint NOT NULL DEFAULT extract(epoch FROM NOW()),
				"updated_on" bigint,
				"archived_on" bigint,
				UNIQUE ("username")
			);`,
		},
		{
			Version:     2,
			Description: "Add OAuth2 Clients table",
			Script: `
			CREATE TABLE IF NOT EXISTS oauth_clients (
				"id" bigserial NOT NULL PRIMARY KEY,
				"client_id" text NOT NULL,
				"client_secret" text NOT NULL,
				"redirect_uri" text DEFAULT '',
				"scopes" text NOT NULL,
				"implicit_allowed" BOOLEAN NOT NULL DEFAULT 'false',
				"created_on" bigint NOT NULL DEFAULT extract(epoch FROM NOW()),
				"updated_on" bigint DEFAULT NULL,
				"archived_on" bigint DEFAULT NULL,
				"belongs_to" bigint NOT NULL,
				FOREIGN KEY(belongs_to) REFERENCES users(id)
			);`,
		},
		{
			Version:     3,
			Description: "Create items table",
			Script: `
			CREATE TABLE IF NOT EXISTS items (
				"id" bigserial NOT NULL PRIMARY KEY,
				"name" text NOT NULL,
				"details" text NOT NULL DEFAULT '',
				"created_on" bigint NOT NULL DEFAULT extract(epoch FROM NOW()),
				"updated_on" bigint,
				"completed_on" bigint,
				"belongs_to" bigint NOT NULL,
				FOREIGN KEY ("belongs_to") REFERENCES "users"("id")
			);`,
		},
	}
)

// Migrate migrates a postgres database
func (p *Postgres) Migrate(ctx context.Context) error {
	p.logger.Info("migrating database")
	if !p.IsReady(ctx) {
		return errors.New("database is not ready yet")
	}

	driver := darwin.NewGenericDriver(p.database, darwin.PostgresDialect{})
	err := darwin.New(driver, migrations, nil).Migrate()

	if err != nil {
		p.logger.Error(err, "migrating database")
	}

	return err
}
